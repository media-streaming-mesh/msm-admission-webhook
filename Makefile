# Image URL to use all building/pushing image targets
TAG ?= latest 
IMAGE_REPOSITORY ?= ciscolabs
IMG ?= ${IMAGE_REPOSITORY}:$(TAG)
GOLANGCI_VERSION = 1.52.2

check: fumpt vet lint ## Run tests and linters

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b ./bin v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint ## Run linter
	bin/golangci-lint run -c .golangci.yaml --timeout 3m

# Build image binary
.PHONY: build
build:
	docker build -t ${IMG} .

# Run go fmt against code
fmt:
	go fmt ./...

# Run go fumpt against code
fumpt:
	gofumpt -d -w .

# Run go vet against code
vet:
	go vet ./...
