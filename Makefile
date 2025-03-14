.DEFAULT_GOAL := build

.PHONY: build
build: lint test
	@hack/do.sh build

.PHONY: build-all
build-all:
	@hack/do.sh build -a

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

.PHONY: lint-more
lint-more:
	@hack/do.sh lint_more

.PHONY: test
test:
	@hack/do.sh test

.PHONY: act
act:
	hack/do.sh act

.PHONY: docs
docs:
	@hack/do.sh docs

.PHONY: completions
completions:
	@hack/do.sh completions

.PHONY: vet
vet:
	@hack/do.sh vet

.PHONY: race
race:
	@hack/do.sh race


.PHONY: bench
bench:
	@hack/do.sh bench

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

.PHONY: dev-deps
dev-deps:
	@hack/do.sh dev_deps
