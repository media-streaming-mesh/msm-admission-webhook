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

const (
	// return codes
	couldNotEncodeReview = "could not encode response: %v"
	couldNotWriteReview  = "could not write response: %v"
	invalidContentType   = "invalid Content-Type=%v, expect \"application/json\""
	emptyBody            = "empty body"
	unsupportedKind      = "kind %v is not supported"

	// msm-config values
	defaultPort    = 443
	defaultRepo    = "ciscolabs"
	defaultTag     = "latest"
	defaultSidecar = "msm-rtsp-stub"
	pullPolicyEnv  = "IMAGE_PULL_POLICY"
	repoEnv        = "REPO"
	tagEnv         = "TAG"
	sidecarEnv     = "MSM_SIDECAR"
	msmLogLvlEnv   = "MSM_LOG_LVL"
	defaultLogLvl  = "WARN"
	msmCpEnv       = "MSM_CONTROL_PLANE"
	msmDpEnv       = "MSM_DATA_PLANE"

	// msm-specific values
	msmAnnotationKey = "sidecar.mediastreamingmesh.io/inject"
	msmAnnotation    = "mediastreamingmesh.io"
	msmServiceName   = "msm-admission-webhook-svc"
	msmName          = "msm-admission-webhook"
	msmConfigMap     = "msm-sidecar-cfg"

	// k8s-specific values
	deployment                = "Deployment"
	pod                       = "Pod"
	daemonSet                 = "DaemonSet"
	statefulSet               = "StatefulSet"
	mutateMethod              = "/mutate"
	deploymentSubPath         = "/spec/template"
	volumePath                = "/spec/volumes"
	containersPath            = "/spec/containers"
	admissionReviewKind       = "AdmissionReview"
	admissionReviewAPIVersion = "admission.k8s.io/v1"

	// Downward API Injection values
	podName      = "MSM_POD_NAME"
	podNamePath  = "metadata.name"
	nodeName     = "MSM_NODE_NAME"
	nodeNamePath = "spec.nodeName"
	nsName       = "MSM_POD_NAMESPACE"
	nsPath       = "metadata.namespace"
	podIPName    = "MSM_POD_IP"
	podIPPath    = "status.podIP"
	saName       = "MSM_POD_SERVICE_ACCOUNT"
	saPath       = "spec.serviceAccountName"
)
