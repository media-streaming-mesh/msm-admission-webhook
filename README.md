# MSM Admission Webhooks

Admission webhooks are HTTP callbacks that receive admission requests and do something with them.
There are two types of admission webhooks, validating admission webhook and mutating admission webhook.
Current MSM implementation uses mutating admission webhooks, in which you can change requests to enforce custom defaults.
More specifically MSM uses MutatingAdmissionWebhooks to automatically inject the MSM stub (sidecar proxy) into an application pod.

## Usage
The easiest way to get started with the MSM Admission Webhook is by using
a deployment from the MSM deployment repo.  Some instructions can
be found here
[MSM Admission Webhook Helm deployment](https://github.com/media-streaming-mesh/deployments-kubernetes/tree/main/deployments/msm-helm)

### Pod and Deployment mutation
At the time of writing the default `webhookconfiguration` will cause
the server to be called for all pods and all deployments across all
namespaces. The logic will check for an annotation on the object
in order to proceed with mutation.  The annotation key should be
`sidecar.mediastreamingmesh.io/inject` it's value does not matter.

## Implementation Details

The MutatingAdmissionWebhook needs three objects to function:

1. MutatingWebhookConfiguration
   - A MutatingAdmissionWebhook needs to be registered with the K8s API server by providing a MutatingWebhookConfiguration.
1. MutatingAdmissionWebhook itself
   - A MutatingAdmissionWebhook is a plugin-style admission controller that can be configured into the K8s API server. The MutatingAdmissionWebhook plugin gets the list of interested admission webhooks from MutatingWebhookConfiguration. Then the MutatingAdmissionWebhook observes the allowed request to the API server and intercepts any matching rules in admission webhooks and calls them in parallel.
1. Webhook Admission Server
   Webhook Admission Server is just plain http server that adhere to Kubernetes API. For each request to the K8s API Server, the MutatingAdmissionWebhook sends an admissionReview to the relevant webhook admission server and patch the requested resource

### MSM Mutating Admission Webhook

In the MSM implementation, the msm-admission-webhook mutating configuration gets injected during the MSM installation process.
In order for the controller to modify the pod specification in runtime by introducing the MSM stub (sidecar proxy) container to the actual pod specification it needs to be triggered at the object level.
This is the most important part for the pod injection to work as the Deployment or Pod specifications need to be annotated with the following MSM key "sidecar.mediastreamingmesh.io/inject"="true".
If the above annotation is found, the controller is triggered and returns the modified object back to the admission webhook for object validation. After validation, the modified pod is deployed with the sidecar container running alongside the application container(s).

#### MSM Mutating Admission Webhook Overview

- [MSM Admission Webhook Helm chart](https://github.com/media-streaming-mesh/deployments-kubernetes/blob/master/examples/features/cni/msm-admission-webhook.yaml)
    - `msm-admission-webhook-svc` service for the actual Webhook admission server
    - `msm-admission-webhook` deployment of the admission webhook server where the image of the injected sidecar has to be defined
    - creates service-account `msm-admission-webhook-sa` and `ClusterRoleBinding` to allow any queries for pods from K8s API

## Building the image
The image is built in a docker container. The Dockerfile is at the base of the repo.  It can be built with
```bash
docker build -t $HUB/msm-admission-webhook:$TAG .
```
where $HUB and $TAG can be set to match the helm values used
during installation.
