package api

import _ "embed"

//go:embed service.swagger.json
var OpenAPISpec []byte
