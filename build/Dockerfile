FROM alpine:latest

RUN apk --no-cache add ca-certificates

LABEL description="MSM Admission Webhook Sidecar Injector"

COPY msm-admission-webhook /bin/msm-admission-webhook

CMD ["/bin/msm-admission-webhook"]