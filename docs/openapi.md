# Generating client code

The code used to talk to inventory server is generated and placed in
`internal/apiclient/client.go`.

`oapi-codegen` is used to generate the client code.

It only understands OpenAPI 3.0 but the server uses OpenAPI 2.0/swagger.

So in order to generate the client code, you need to convert the spec from 2.0
to 3.0.

There are multiple tools that can do so, e.g. the node module `swagger2openapi`.

Run it like this:

```bash
npx swagger2openapi -o docs/openapi.normalized.json path/to/openapi-v2-spec.json
```

The spec can be obtained from the server source code in
`docs/data/openapi.normalized.json`.

Once converted run `go generate`:

```bash
go generate ./...
```

Or manually:

```bash
oapi-codegen -package=apiclient -generate=types,client -o=internal/apiclient/client.go docs/openapi.normalized.json
```
