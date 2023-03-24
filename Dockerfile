# Build the manager binary
FROM golang:1.19 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY cmd/ cmd/
COPY internal/ internal/

# get the Go modules - need to cache this!
RUN go mod download

# Build
RUN GOOS=linux go build -a -o msm-admission-webhook cmd/main.go

# runtime
FROM ubuntu
WORKDIR /
COPY --from=builder /workspace/msm-admission-webhook .

ENTRYPOINT ["/msm-admission-webhook"]
