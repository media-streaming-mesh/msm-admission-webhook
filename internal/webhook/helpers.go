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
	"errors"
	"net/url"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isSupportKind(request *v1.AdmissionRequest) bool {
	rk := request.Kind.Kind
	return rk == pod || rk == deployment || rk == statefulSet || rk == daemonSet
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

func (w *MsmWebhook) validateAnnotationValue(value string) error {
	urls, err := parseAnnotationValue(value)
	w.Log.Debugf("Annotation result: %v", urls)
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

func getRepo() string {
	repo := os.Getenv(repoEnv)
	if repo == "" {
		return defaultRepo
	}

	return repo
}

func getSidecar() string {
	sidecar := os.Getenv(sidecarEnv)
	if sidecar == "" {
		return defaultSidecar
	}

	return sidecar
}

func getTag() string {
	tag := os.Getenv(tagEnv)
	if tag == "" {
		return defaultTag
	}

	return tag
}

func (w *MsmWebhook) applyDeploymentKind(patches []patchOperation, kind string) {
	switch kind {
	case pod:
		return
	case deployment:
		for i := 0; i < len(patches); i++ {
			patches[i].Path = deploymentSubPath + patches[i].Path
		}
	case daemonSet:
		for i := 0; i < len(patches); i++ {
			patches[i].Path = deploymentSubPath + patches[i].Path
		}
	case statefulSet:
		for i := 0; i < len(patches); i++ {
			patches[i].Path = deploymentSubPath + patches[i].Path
		}
	default:
		w.Log.Fatalf(unsupportedKind, kind)
	}
}
