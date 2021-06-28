package webhook

import (
	"encoding/json"
	"net/url"

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
	w.Log.Infof("AdmissionReview for req=%v", request)

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
	applyDeploymentKind(patch, request.Kind.Kind)
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return errorReviewResponse(err)
	}

	w.Log.Infof("AdmissionResponse, patch=%v\n", string(patchBytes))
	return createReviewResponse(patchBytes)
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

func getMetaAndSpec(request *v1.AdmissionRequest) (*podSpecAndMeta, error) {
	result := &podSpecAndMeta{}
	switch request.Kind.Kind {
	case deployment:
		var d appsv1.Deployment
		if err := json.Unmarshal(request.Object.Raw, &d); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			return nil, err
		}
		result.meta = &d.ObjectMeta
		result.spec = &d.Spec.Template.Spec
	case pod:
		var p corev1.Pod
		if err := json.Unmarshal(request.Object.Raw, &p); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			return nil, err
		}
		result.meta = &p.ObjectMeta
		result.spec = &p.Spec
	case statefulSet:
		var ss appsv1.StatefulSet
		if err := json.Unmarshal(request.Object.Raw, &ss); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			return nil, err
		}
		result.meta = &ss.ObjectMeta
		result.spec = &ss.Spec.Template.Spec
	case daemonSet:
		var ds appsv1.StatefulSet
		if err := json.Unmarshal(request.Object.Raw, &ds); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			return nil, err
		}
		result.meta = &ds.ObjectMeta
		result.spec = &ds.Spec.Template.Spec
	}

	return result, nil
}
