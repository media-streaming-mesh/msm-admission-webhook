package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	v1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NSUrl struct {
	NsName string
	Intf   string
	Params url.Values
}

type podSpecAndMeta struct {
	meta *metav1.ObjectMeta
	spec *corev1.PodSpec
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

func (w *MsmWebhook) mutate(request *v1.AdmissionRequest) *v1.AdmissionResponse {
	w.Log.Infof("AdmissionReview for =%v", request)

	if !isSupportKind(request) {
		return okReviewResponse()
	}

	metaAndSpec, err := getMetaAndSpec(request)
	if err != nil {
		return errorReviewResponse(err)
	}

	value, ok := msmAnnotationValue(ignoredNamespaces, metaAndSpec)
	if !ok {
		logrus.Infof("Skipping validation for %s/%s due to policy check", metaAndSpec.meta.Namespace, metaAndSpec.meta.Name)
		return okReviewResponse()
	}

	if err = validateAnnotationValue(value); err != nil {
		return errorReviewResponse(err)
	}

	// todo - set limits
	// todo - init container duplication

	// create container to inject into pod
	patch := createMsmContainerPatch(metaAndSpec, value)
	// applyDeploymentKind(patch, request.Kind.Kind)
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return errorReviewResponse(err)
	}

	w.Log.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return createReviewResponse(patchBytes)
}

func createMsmContainerPatch(tuple *podSpecAndMeta, annotationValue string) (patch []patchOperation) {
	msmProxyContainer := corev1.Container{
		Name:            "msm-proxy-sidecar",
		Command:         []string{"/bin/test"},
		Image:           fmt.Sprintf("%s/%s:%s", "test", "test", "test"),
		ImagePullPolicy: getPullPolicyValue(),
	}

	patch = append(patch, addContainer(tuple.spec, []corev1.Container{msmProxyContainer})...)

	return patch
}

func addContainer(spec *corev1.PodSpec, containers []corev1.Container) (patch []patchOperation) {
	first := len(spec.Containers) == 0
	for i := 0; i < len(containers); i++ {
		value := &containers[i]
		path := containersPath
		if first {
			first = false
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}

	return patch
}

func isSupportKind(request *v1.AdmissionRequest) bool {
	return request.Kind.Kind == pod || request.Kind.Kind == deployment
}

func errorReviewResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func okReviewResponse() *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Allowed: true,
	}
}

func createReviewResponse(data []byte) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Allowed: true,
		Patch:   data,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func msmAnnotationValue(ignoredNamespaceList []string, tuple *podSpecAndMeta) (string, bool) {

	// skip special kubernetes system namespaces
	for _, namespace := range ignoredNamespaceList {
		if tuple.meta.Namespace == namespace {
			logrus.Infof("Skip validation for %v for it's in special namespace:%v", tuple.meta.Name, tuple.meta.Namespace)
			return "", false
		}
	}

	annotations := tuple.meta.GetAnnotations()
	if annotations == nil {
		logrus.Info("No annotations, skip")
		return "", false
	}

	value, ok := annotations[msmAnnotationKey]
	return value, ok
}

func validateAnnotationValue(value string) error {
	urls, err := parseAnnotationValue(value)
	logrus.Infof("Annotation result: %v", urls)
	return err
}

func parseAnnotationValue(value string) ([]*NSUrl, error) {
	var result []*NSUrl
	urls := strings.Split(value, ",")
	for _, u := range urls {
		nsurl, err := parseNSUrl(u)
		if err != nil {
			return nil, err
		}
		result = append(result, nsurl)
	}
	return result, nil
}

func parseNSUrl(urlString string) (*NSUrl, error) {
	result := &NSUrl{}
	// Remove possible leading spaces from network service name
	urlString = strings.Trim(urlString, " ")
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	path := strings.Split(url.Path, "/")
	if len(path) > 2 {
		return nil, errors.New("Invalid NSUrl format")
	}
	if len(path) == 2 {
		if len(path[1]) > 15 {
			return nil, errors.New("Interface part cannot exceed 15 characters")
		}
		result.Intf = path[1]
	}
	result.NsName = path[0]
	result.Params = url.Query()
	return result, nil
}

func getPullPolicyValue() corev1.PullPolicy {
	pullPolicy := os.Getenv(pullPolicyEnv)
	if pullPolicy == "" {
		return corev1.PullIfNotPresent
	}

	return corev1.PullPolicy(pullPolicy)
}

func getMetaAndSpec(request *v1.AdmissionRequest) (*podSpecAndMeta, error) {
	result := &podSpecAndMeta{}
	if request.Kind.Kind == deployment {
		var deployment appsv1.Deployment
		if err := json.Unmarshal(request.Object.Raw, &deployment); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			return nil, err
		}
		result.meta = &deployment.ObjectMeta
		result.spec = &deployment.Spec.Template.Spec
	}
	if request.Kind.Kind == pod {
		var pod corev1.Pod
		if err := json.Unmarshal(request.Object.Raw, &pod); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			return nil, err
		}
		result.meta = &pod.ObjectMeta
		result.spec = &pod.Spec
	}
	return result, nil
}
