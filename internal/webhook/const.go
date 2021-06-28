package webhook

const (
	// return codes
	couldNotEncodeReview = "could not encode response: %v"
	couldNotWriteReview  = "could not write response: %v"
	invalidContentType   = "invalid Content-Type=%v, expect \"application/json\""
	emptyBody            = "empty body"
	unsupportedKind      = "kind %v is not supported"

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
	msmAnnotationKey = "sidecar.mediastreamingmesh.io/inject"
	msmAnnotation    = "mediastreamingmesh.io"
	msmServiceName   = "msm-admission-webhook-svc"
	msmName          = "msm-admission-webhook"
	msmNamespace     = "default"
	msmVolume        = "msm-volume"
	msmVolumeCfg     = "msm-proxy-cfg"

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
)
