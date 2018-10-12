package cellar

import (
	swagger "goa.design/goa/examples/cellar/gen/swagger"
	goalog "goa.design/goa/logging"
)

// swagger service example implementation.
// The example methods log the requests and return zero values.
type swaggerSvc struct {
	logger goalog.Logger
}

// Required for compatibility with Service interface
func (s *swaggerSvc) GetLogger() goalog.Logger {
	return s.logger
}

// NewSwagger returns the swagger service implementation.
func NewSwagger(logger goalog.Logger) swagger.Service {
	return &swaggerSvc{logger: logger}
}
