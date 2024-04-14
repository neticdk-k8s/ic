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
	docker buildx build --target bin \
		--output bin/ \
		--platform ${PLATFORM} \
		--tag netic/ic .
	docker buildx build --platform ${PLATFORM} \
		--tag netic/ic .

.PHONY: release-patch
release-patch:
	@echo "Releasing patch version..."
	@hack/release.sh patch

.PHONY: release-minor
release-minor:
	@echo "Releasing minor version..."
	@hack/release.sh minor

.PHONY: lint
lint:
	@echo "Linting..."
	@golangci-lint run

.PHONY: clean
clean:
	@echo "Cleaning..."
	@go clean

.PHONY: fmt
fmt:
	@echo "Formatting..."
	@go fmt ./...

.PHONY: gen
gen:
	@echo "Generating code..."
	@go generate ./...

.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...

.PHONY: doc
doc:
	@echo "Generating documentation..."
	@go run docs/gen.go

.PHONY: build-all
build-all: clean fmt lint | $(BIN)
	@echo "Building ic..."
	CGO_ENABLED=0 go build -o $(BIN)/ic \
		-v \
		-a \
		-tags release \
		-ldflags '-s -w -X main.version=$(VERSION)'

.PHONY: build
build: clean fmt lint | $(BIN)
	@echo "Building ic..."
	CGO_ENABLED=0 go build -o $(BIN)/ic \
		-v \
		-tags release \
		-ldflags '-s -w -X main.version=$(VERSION)'

.PHONY: build-nolint
build-nolint: clean fmt | $(BIN)
	@echo "Building ic..."
	CGO_ENABLED=0 go build -o $(BIN)/ic \
		-v \
		-tags release \
		-ldflags '-s -w -X main.version=$(VERSION)'

.PHONY: install
install: clean fmt lint | $(BIN)
	@echo "Installing ic..."
	CGO_ENABLED=0 go install \
		-v \
		-tags release \
		-ldflags '-s -w -X main.version=$(VERSION)'

.PHONY: docker-build
docker-build:
	@echo "Building docker image..."
	docker buildx build --progress plain --build-arg GITHUB_USER=${GITHUB_USER} --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} --build-arg MODULEPATH=${MODULEPATH} --build-arg VERSION=$(VERSION) -t registry.netic.dk/netic/ic:latest -t registry.netic.dk/netic/ic:${VERSION} --load .

.PHONY: docker-build-push
docker-build-push:
	@echo "Building docker image..."
	docker buildx build --progress plain --platform linux/arm64,linux/amd64 --build-arg GITHUB_USER=${GITHUB_USER} --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} --build-arg MODULEPATH=${MODULEPATH} --build-arg VERSION=$(VERSION) -t registry.netic.dk/netic/ic:latest -t registry.netic.dk/netic/ic:${VERSION} --push .
