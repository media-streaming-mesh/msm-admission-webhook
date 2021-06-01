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

	// defaults values
	defaultPort      = 443
	defaultNamespace = "default"
	defaultRepo      = "networkservicemesh"
	defaultTag       = "latest"

	// msm-specific values
	msmAnnotationKey = "media-streaming-mesh.io"
	msmMamespace     = "MSM_NAMESPACE"
	pullPolicyEnv    = "IMAGE_PULL_POLICY"
	repoEnv          = "REPO"
	tagEnv           = "TAG"

	// k8s-specific values
	mutateMethod   = "/mutate"
	deployment     = "Deployment"
	pod            = "Pod"
	containersPath = "/spec/containers"
)
