package apiclient

//go:generate go tool oapi-codegen -package=apiclient -generate=types,client -o=client.go ../../docs/openapi.normalized.json
