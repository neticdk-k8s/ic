.DEFAULT_GOAL := build
VERSION ?= $(shell (git describe --tags --always --match=v* 2> /dev/null || echo v0) | sed -e s/^v//)
COMMIT = $(shell git rev-parse HEAD)
MODULEPATH := $(shell go mod edit -json 2> /dev/null | jq -r '.Module.Path')
GOPRIVATE := "github.com/neticdk-k8s/scs-domain-model"
GITHUB_USER := $(shell printf "protocol=https\\nhost=github.com\\n" | git credential-manager get | grep username | cut -d= -f2)
GITHUB_TOKEN := $(shell printf "protocol=https\\nhost=github.com\\n" | git credential-manager get | grep password | cut -d= -f2)

BIN = $(CURDIR)/bin
$(BIN):
	@mkdir -p $@

PLATFORM=local

.PHONY: bin/ic
bin/ic:
	@DOCKER_BUILDKIT=1 docker build --target bin \
		--output bin/ \
		--platform ${PLATFORM} \
		--tag netic/k8s-inventory-cli \
		cmd/ic/main.go
	@DOCKER_BUILDKIT=1 docker build --platform ${PLATFORM} \
		--tag netic/k8s-inventory-cli \
		cmd/ic/main.go

.PHONY: release-patch
release-patch:
	@echo "Releasing patch version..."
	@hack/release.sh patch

.PHONY: release-minor
release-minor:
	@echo "Releasing minor version..."
	@hack/release.sh minor

# Runs go lint
.PHONY: lint
lint:
	@echo "Linting..."
	@golangci-lint run

# Runs go clean
.PHONY: clean
clean:
	@echo "Cleaning..."
	@go clean

# Runs go fmt
.PHONY: fmt
fmt:
	@echo "Formatting..."
	@go fmt ./...

# Runs go build
.PHONY: build
build: clean fmt lint | $(BIN)
	@echo "Building k8s-inventory-cli..."
	CGO_ENABLED=0 go build -o $(BIN)/ic \
		-v \
		-a \
		-tags release \
		-ldflags '-s -w -X ${MODULEPATH}/internal/version.VERSION=$(VERSION) -X ${MODULEPATH}internal/version.COMMIT=$(COMMIT)' \
		cmd/ic/main.go

# Runs go build
.PHONY: build2
build2: clean fmt | $(BIN)
	@echo "Building k8s-inventory-cli..."
	CGO_ENABLED=0 go build -o $(BIN)/ic \
		-v \
		-tags release \
		-ldflags '-s -w -X ${MODULEPATH}/internal/version.VERSION=$(VERSION) -X ${MODULEPATH}/internal/version.COMMIT=$(COMMIT)' \
		cmd/ic/main.go

# Build docker image
.PHONY: docker-build
docker-build:
	@echo "Building k8s-inventory-ic image..."
	DOCKER_BUILDKIT=1 docker build --network=host --progress=plain --no-cache --build-arg GITHUB_USER=${GITHUB_USER} --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} --build-arg MODULEPATH=${MODULEPATH} --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -t netic/k8s-inventory-cli .

# Tag and push docker image
.PHONY: docker-push
docker-push:
	docker tag netic/k8s-inventory-cli:latest registry.netic.dk/netic/k8s-inventory-cli:latest
	docker push registry.netic.dk/netic/k8s-inventory-cli:latest
	docker tag netic/k8s-inventory-cli:latest registry.netic.dk/netic/k8s-inventory-cli:${VERSION}
	docker push registry.netic.dk/netic/k8s-inventory-cli:${VERSION}

# Build, tag and push docker image
.PHONY: docker-all
docker-all: docker-build docker-push

.PHONY: docker-cross
docker-cross:
	docker buildx build --platform linux/amd64 --build-arg GITHUB_USER=${GITHUB_USER} --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} --build-arg MODULEPATH=${MODULEPATH} --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -t registry.netic.dk/netic/k8s-inventory-cli:latest -t registry.netic.dk/netic/k8s-inventory-cli:${VERSION} --push .
