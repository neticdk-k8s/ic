# Inventory CLI (ic)

[![ci](https://github.com/neticdk-k8s/ic/actions/workflows/main.yml/badge.svg)](https://github.com/neticdk-k8s/ic/actions/workflows/main.yml)
[![tag](https://img.shields.io/github/tag/neticdk-k8s/ic.svg)](https://github.com/neticdk-k8s/ic/tags/)

This is the CLI used to interact with k8s-inventory-server.

## Documentation

See [docs/ic.md](docs/ic.md) for more documentation on the commands.

## Development

### Configuration for Local Development

You might want to create a configuration file named `ic.toml` in the root
directory that looks something like this:

```toml
log-level="debug"
oidc-issuer-url="http://localhost:8080/realms/test"
api-server="http://localhost:8087"
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
client code. See [`docs/openapi.md`](docs/openapi.md).

### Make Targets

- `make build` builds `bin/ic`
- `make test` runs tests
- `make docker-build` builds a docker image
- `make gen` runs code generation
- `make doc` generates command line documentation in `docs/`
- `make release-patch` tags and pushes the next patch release
- `make release-minor` tags and pushes a new minor release
