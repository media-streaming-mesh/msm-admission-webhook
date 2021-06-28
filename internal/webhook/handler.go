package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// handle
func (w *MsmWebhook) handle(rw http.ResponseWriter, r *http.Request) {

	body, err := w.readRequest(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}

	msmAdmissionWebhookReview := v1.AdmissionReview{}
	requestReview, err := w.parseAdmissionReview(body)
	if err != nil {
		msmAdmissionWebhookReview.Response = &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		if r.URL.Path == mutateMethod {
			msmAdmissionWebhookReview.Response = w.mutate(requestReview.Request)
		}
	}
	msmAdmissionWebhookReview.Response.UID = requestReview.Request.UID
	msmAdmissionWebhookReview.Kind = "AdmissionReview"
	msmAdmissionWebhookReview.APIVersion = "admission.k8s.io/v1"

	resp, err := json.Marshal(msmAdmissionWebhookReview)
	if err != nil {
		w.Log.Errorf("Can't encode response: %v", err)
		http.Error(rw, fmt.Sprintf(couldNotEncodeReview, err), http.StatusInternalServerError)
	}

	if _, err := rw.Write(resp); err != nil {
		w.Log.Errorf("Can't write response: %v", err)
		http.Error(rw, fmt.Sprintf(couldNotWriteReview, err), http.StatusInternalServerError)
	}
}

// readRequest handles the intercepting request of the api server
func (w *MsmWebhook) readRequest(r *http.Request) ([]byte, error) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		w.Log.Error(emptyBody)
		return nil, errors.New(emptyBody)
	}
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		msg := fmt.Sprintf(invalidContentType, contentType)
		w.Log.Error(msg)
		return nil, errors.New(msg)
	}
	return body, nil
}

// parseAdmissionReview
func (w *MsmWebhook) parseAdmissionReview(body []byte) (*v1.AdmissionReview, error) {
	r := &v1.AdmissionReview{}
	if _, _, err := w.deserializer.Decode(body, nil, r); err != nil {
		return nil, err
	}
	return r, nil
}
