# Inventory CLI (ic)

[![ci](https://github.com/neticdk-k8s/ic/actions/workflows/main.yml/badge.svg)](https://github.com/neticdk-k8s/ic/actions/workflows/main.yml)
[![tag](https://img.shields.io/github/tag/neticdk-k8s/ic.svg)](https://github.com/neticdk-k8s/ic/tags/)

This is the CLI used to interact with k8s-inventory-server.

See [`docs/`](docs/) for more documentation on the commands.

## Development

You might want to create a config file named `ic.toml` in the root directory
that looks something like this:

```toml
log-level="debug"
oidc-issuer-url="http://localhost:8080/realms/test"
api-server="http://localhost:8087"
```
