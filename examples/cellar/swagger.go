package cellar

import (
	"log"

	swagger "goa.design/goa/examples/cellar/gen/swagger"
)

// swagger service example implementation.
// The example methods log the requests and return zero values.
type swaggerSvc struct {
	logger *log.Logger
}

// NewSwagger returns the swagger service implementation.
func NewSwagger(logger *log.Logger) swagger.Service {
	return &swaggerSvc{logger}
}
