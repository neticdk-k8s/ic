# Inventory CLI (ic)

[![ci](https://github.com/neticdk-k8s/ic/actions/workflows/main.yml/badge.svg)](https://github.com/neticdk-k8s/ic/actions/workflows/main.yml)
[![tag](https://img.shields.io/github/tag/neticdk-k8s/ic.svg)](https://github.com/neticdk-k8s/ic/tags/)

This is the CLI used to interact with k8s-inventory-server.

## Installation

<details>
<summary>From Release Distribution on GitHub</summary>

This only works on MacOS and Linux:

```bash
tag=$(
    curl --silent -H "Accept: application/vnd.github.v3+json" \
    https://api.github.com/repos/neticdk-k8s/ic/releases/latest \
    | jq -r .tag_name
)

curl -L https://github.com/neticdk-k8s/ic/releases/download/${tag}/ic-${tag}-$(uname -s|tr A-Z a-z)-$(uname -m).tar.gz \
    | tar xzf - /usr/local/bin/ic
```

For Windows, go to the [release
page](https://github.com/neticdk-k8s/ic/releases/latest) and download the zip
archive.

</details>

<details>
<summary>From Source</summary>

### Using `go install`

You will need:

- go

Run:

```bash
go install github.com/neticdk-k8s/ic@latest
```

The executable will be installed in `$GOPATH/bin` Add it to your `PATH` if you haven't
already.

### Using `make install`

You will need:

- go
- golangci-lint

Checkout this repository and run:

```bash
make install
```

The executable will be installed in `$GOPATH/bin` Add it to your `PATH` if you
haven't already.

</details>

## Introduction

Basic usage:

```bash
ic COMMAND [flags]
```

### Authentication

Most commands require authentication. By default, browser based OICD
authentication will be used.

If you want to use keyboard based OICD authentication you can use the
`--oidc-grant-type authcode-keyboard` flag.

`ic` will try to refresh the token on every run.

Tokens are cached in the default user cache directory for the Operating System
`ic` is running on:

- `~/Library/Caches/ic/oidc-login/` on MacOS
- `$XDG_CACHE_HOME` (typically `$HOME/.cache`) on Linux
- `%LocalAppData%` on Windows

### Output

Commands will output log messages and errors to `stderr` and normal output to
`stdout`.

The default format is `text` which in usually means tables. In most cases
commands will support `json` output via the `--output-format json` flag.

Colors and other flashy things are disabled while running in a non-interactive
environment (e.g. when redirecting output to a log file). This can be controlled
via the `--interactive` flag.

If you don't want headers printed you can use the `--no-headers` flag. This can
be useful for piping output to other commands.

## Commands and Usage

See [docs/ic.md](docs/ic.md) for more documentation on the commands.

You may also run `ic help`.

## Development

### Configuration for Local Development

You might want to create a configuration file named `ic.toml` in the root
directory that looks something like this:

```toml
log-level="debug"
oidc-issuer-url="http://localhost:8080/realms/test"
api-server="http://localhost:8087"
```

### Using api-token and curl with the API server

Use it like this:

```bash
curl -H "Authorization: Bearer $(bin/ic api-token 2>/dev/null)" http://localhost:8087/clusters | jq .
```

### Code Generation

#### Mocks

Interface mocks are generated using [mockery](https://github.com/vektra/mockery)

The command used is:

```bash
mockery --with-expecter --inpackage --name <interface name>
```

### OpenAPI Client Code

The inventory server provides an OpenAPI 2.0 spec.

We use [oapi-codegen](https://github.com/deepmap/oapi-codegen) to generate the
client code. See [docs/openapi.md](docs/openapi.md).

### Make Targets

- `make build` builds `bin/ic`
- `make test` runs tests
- `make install` builds and install the `ic` command
- `make docker-build` builds a docker image
- `make gen` runs code generation
- `make doc` generates command line documentation in `docs/`
- `make release-patch` tags and pushes the next patch release
- `make release-minor` tags and pushes a new minor release
