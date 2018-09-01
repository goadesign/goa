package calc

import (
	"log"

	openapi "goa.design/goa/examples/calc/gen/openapi"
)

// openapi service example implementation.
// The example methods log the requests and return zero values.
type openapiSvc struct {
	logger *log.Logger
}

// NewOpenapi returns the openapi service implementation.
func NewOpenapi(logger *log.Logger) openapi.Service {
	return &openapiSvc{logger}
}
