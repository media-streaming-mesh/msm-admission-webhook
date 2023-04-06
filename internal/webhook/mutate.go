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
	"net/url"

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

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

func (w *MsmWebhook) mutate(request *v1.AdmissionRequest) *v1.AdmissionResponse {
	w.Log.Debugf("AdmissionReview for request UID %s, Kind %s, "+
		"Resource %s, Name %s, Namespace %s, Operation %s ",
		request.UID, request.Kind, request.Resource, request.Name,
		request.Namespace, request.Operation)

	if !isSupportKind(request) {
		return okReviewResponse()
	}

	metaAndSpec, err := w.getMetaAndSpec(request)
	if err != nil {
		return errorReviewResponse(err)
	}

	value, ok := w.msmAnnotationValue(ignoredNamespaces, metaAndSpec)
	if !ok {
		w.Log.Infof("Skipping validation for %s/%s due to policy check", metaAndSpec.meta.Namespace, metaAndSpec.meta.Name)
		return okReviewResponse()
	}

	if err = w.validateAnnotationValue(value); err != nil {
		return errorReviewResponse(err)
	}

	// todo - set limits
	// todo - init container duplication

	// create container to inject into pod
	patch := createMsmContainerPatch(metaAndSpec, value)
	w.applyDeploymentKind(patch, request.Kind.Kind)
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return errorReviewResponse(err)
	}

	w.Log.Debugf("AdmissionResponse, patch=%v\n", string(patchBytes))
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

func (w *MsmWebhook) msmAnnotationValue(ignoredNamespaceList []string, tuple *podSpecAndMeta) (string, bool) {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredNamespaceList {
		if tuple.meta.Namespace == namespace {
			w.Log.Infof("Skip validation for %v for it's in special namespace:%v", tuple.meta.Name, tuple.meta.Namespace)
			return "", false
		}
	}

	annotations := tuple.meta.GetAnnotations()
	if annotations == nil {
		w.Log.Info("No annotations, skip")
		return "", false
	}

	value, ok := annotations[msmAnnotationKey]
	return value, ok
}

func (w *MsmWebhook) getMetaAndSpec(request *v1.AdmissionRequest) (*podSpecAndMeta, error) {
	result := &podSpecAndMeta{}
	switch request.Kind.Kind {
	case deployment:
		var d appsv1.Deployment
		if err := json.Unmarshal(request.Object.Raw, &d); err != nil {
			w.Log.Errorf("Could not unmarshal raw object: %v", err)
			return nil, err
		}
		result.meta = &d.ObjectMeta
		result.spec = &d.Spec.Template.Spec
	case pod:
		var p corev1.Pod
		if err := json.Unmarshal(request.Object.Raw, &p); err != nil {
			w.Log.Errorf("Could not unmarshal raw object: %v", err)
			return nil, err
		}
		result.meta = &p.ObjectMeta
		result.spec = &p.Spec
	case statefulSet:
		var ss appsv1.StatefulSet
		if err := json.Unmarshal(request.Object.Raw, &ss); err != nil {
			w.Log.Errorf("Could not unmarshal raw object: %v", err)
			return nil, err
		}
		result.meta = &ss.ObjectMeta
		result.spec = &ss.Spec.Template.Spec
	case daemonSet:
		var ds appsv1.StatefulSet
		if err := json.Unmarshal(request.Object.Raw, &ds); err != nil {
			w.Log.Errorf("Could not unmarshal raw object: %v", err)
			return nil, err
		}
		result.meta = &ds.ObjectMeta
		result.spec = &ds.Spec.Template.Spec
	}

	return result, nil
}
