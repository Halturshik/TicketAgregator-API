package docs

import _ "embed"

// OpenAPI содержит HTTP-контракт, который встраивается в API-бинарник.
//
//go:embed openapi.yaml
var OpenAPI []byte
