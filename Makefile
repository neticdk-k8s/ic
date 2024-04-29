.DEFAULT_GOAL := build

.PHONY: clean
clean:
	@hack/do.sh clean

.PHONY: fmt
fmt:
	@hack/do.sh fmt

.PHONY: lint
lint:
	@hack/do.sh lint

.PHONY: gen
gen:
	@hack/do.sh gen

.PHONY: test
test:
	@hack/do.sh test

.PHONY: docs
docs:
	@hack/do.sh docs

.PHONY: build-all
build-all:
	@hack/do.sh build -a

.PHONY: build
build:
	@hack/do.sh build

.PHONY: build-nolint
build-nolint:
	@NOLINT=1 hack/do.sh build

.PHONY: release-patch
release-patch:
	@hack/do.sh release patch

.PHONY: release-minor
release-minor:
	@hack/do.sh release minor

.PHONY: install
install:
	@hack/do.sh install

.PHONY: docker-build
docker-build:
	@hack/do.sh docker_build

.PHONY: docker-build-push
docker-push:
	@hack/do.sh docker_push
