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
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func createMsmContainerPatch(tuple *podSpecAndMeta, annotationValue string) (patch []patchOperation) {
	uid := int64(1337)
	msmProxyContainer := corev1.Container{
		Name:            getSidecar(),
		Image:           fmt.Sprintf("%s/%s:%s", getRepo(), getSidecar(), getTag()),
		ImagePullPolicy: getPullPolicyValue(),
		Ports: []corev1.ContainerPort{
			{
				Name:          "rtsp",
				ContainerPort: 8554,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "rtp",
				ContainerPort: 8050,
				Protocol:      corev1.ProtocolUDP,
			},
			{
				Name:          "rtcp",
				ContainerPort: 8051,
				Protocol:      corev1.ProtocolUDP,
			},
		},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:                &uid,
			RunAsGroup:               &uid,
			AllowPrivilegeEscalation: new(bool),
		},
		Env: []corev1.EnvVar{
			{
				Name: "LOG_LVL",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: msmConfigMap,
						},
						Key: "LOG_LVL",
					},
				},
			},
			{
				Name: "MSM_CONTROL_PLANE",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: msmConfigMap,
						},
						Key: "MSM_CONTROL_PLANE",
					},
				},
			},
			{
				Name: "MSM_DATA_PLANE",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: msmConfigMap,
						},
						Key: "MSM_DATA_PLANE",
					},
				},
			},
		},
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
