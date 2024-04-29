#!/bin/sh

set -eu

NOLINT=${NOLINT:-0}

set_version() {
	VERSION=$(git describe --tags --always --match=v* 2>/dev/null || echo v0 | sed -e s/^v//)
}

set_default_go_opts() {
	export CGO_ENABLED=0
	DEFAULT_GO_OPTS="-v -tags release -ldflags '-s -w -X main.version=${VERSION}'"
}

set_github_credentials() {
	GITHUB_USER=$(printf "protocol=https\\nhost=github.com\\n" | git credential-manager get | grep username | cut -d= -f2)
	GITHUB_TOKEN=$(printf "protocol=https\\nhost=github.com\\n" | git credential-manager get | grep password | cut -d= -f2)
}

clean() {
	go clean
}

fmt() {
	go fmt
}

lint() {
	which -s golangci-lint || (echo "golangci-lint not found - install it with 'brew install golangci-lint' or similar" && exit 1)
	golangci-lint run
}

gen() {
	go generate ./...
}

test() {
	go test ./... "$*"
}

docs() {
	go run docs/gen.go
}

build() {
	set_version
	set_default_go_opts
	clean
	fmt
	[ "${NOLINT}" -eq 0 ] && lint
	eval go build -o bin/ "$DEFAULT_GO_OPTS" "$*"
}

docker_build() {
	set_version
	set_github_credentials
	docker buildx build --progress plain --build-arg GITHUB_USERNAME="${GITHUB_USER}" --build-arg GITHUB_TOKEN="${GITHUB_TOKEN}" --build-arg VERSION="${VERSION}" -t registry.netic.dk/netic/ic:latest -t registry.netic.dk/netic/ic:"${VERSION}" --load .
}

docker_push() {
	set_version
	set_github_credentials
	docker buildx build --progress plain --build-arg GITHUB_USERNAME="${GITHUB_USER}" --build-arg GITHUB_TOKEN="${GITHUB_TOKEN}" --build-arg VERSION="${VERSION}" -t registry.netic.dk/netic/ic:latest -t registry.netic.dk/netic/ic:"${VERSION}" --push .
}

install() {
	set_version
	set_default_go_opts
	eval go install "$DEFAULT_GO_OPTS" "$*"
}

release() {
	bump="${1:-patch}"
	case "$bump" in
	patch | minor)
		hack/release.sh "${bump}"
		;;
	*)
		echo "unsupported release type: ${bump}"
		exit 1
		;;
	esac
}

"$@"
