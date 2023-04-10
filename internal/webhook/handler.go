/*
 * Copyright (c) 2022 Cisco and/or its affiliates.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

// handle is the http handler for msm webhook admission requests
func (w *MsmWebhook) handle(rw http.ResponseWriter, r *http.Request) {
	body, err := w.readRequest(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}

	msmAdmissionWebhookReview := v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Request:  nil,
		Response: nil,
	}
	requestReview, err := w.parseAdmissionReview(body)
	if err != nil {
		msmAdmissionWebhookReview.Response = &v1.AdmissionResponse{
			UID:     "",
			Allowed: false,
			Result: &metav1.Status{
				TypeMeta: metav1.TypeMeta{
					Kind:       "",
					APIVersion: "",
				},
				ListMeta: metav1.ListMeta{
					SelfLink:           "",
					ResourceVersion:    "",
					Continue:           "",
					RemainingItemCount: nil,
				},
				Status:  "",
				Message: err.Error(),
				Reason:  "",
				Details: nil,
				Code:    0,
			},
			Patch:            nil,
			PatchType:        nil,
			AuditAnnotations: nil,
			Warnings:         nil,
		}
	} else if r.URL.Path == mutateMethod {
		msmAdmissionWebhookReview.Response = w.mutate(requestReview.Request)
	}
	msmAdmissionWebhookReview.Response.UID = requestReview.Request.UID
	msmAdmissionWebhookReview.Kind = admissionReviewKind
	msmAdmissionWebhookReview.APIVersion = admissionReviewAPIVersion

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

func (w *MsmWebhook) parseAdmissionReview(body []byte) (*v1.AdmissionReview, error) {
	r := &v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Request:  nil,
		Response: nil,
	}
	if _, _, err := w.deserializer.Decode(body, nil, r); err != nil {
		return nil, err
	}
	return r, nil
}
