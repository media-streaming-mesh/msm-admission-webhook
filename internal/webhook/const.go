package webhook

import v1 "k8s.io/api/core/v1"

const (

	// certificate files
	certFile = "/etc/webhook/certs/" + v1.TLSCertKey
	keyFile  = "/etc/webhook/certs/" + v1.TLSPrivateKeyKey

	// return codes
	couldNotEncodeReview = "could not encode response: %v"
	couldNotWriteReview  = "could not write response: %v"
	invalidContentType   = "invalid Content-Type=%v, expect \"application/json\""
	emptyBody            = "empty body"
	unsupportedKind      = "kind %v is not supported"
	deploymentSubPath    = "/spec/template"
	volumePath           = "/spec/volumes"
	containersPath       = "/spec/containers"

	// defaults values
	defaultPort    = 443
	defaultRepo    = "ciscolabs"
	defaultTag     = "latest"
	defaultSidecar = "msm-proxy"
	pullPolicyEnv  = "IMAGE_PULL_POLICY"
	repoEnv        = "REPO"
	tagEnv         = "TAG"
	sidecarEnv     = "MSM_SIDECAR"

	// msm-specific values
	msmName          = "msm-admission-webhook"
	msmAnnotationKey = "sidecar.mediastreamingmesh.io/inject"
	msmAnnotation    = "mediastreamingmesh.io"
	msmServiceName   = "msm-admission-webhook-svc"
	msmNamespace     = "default"
	msmVolume        = "msm-volume"
	msmVolumeCfg     = "msm-proxy-cfg"

	// k8s-specific values
	mutateMethod = "/mutate"
	deployment   = "Deployment"
	pod          = "Pod"
	daemonSet    = "DaemonSet"
	statefulSet  = "StatefulSet"
)
