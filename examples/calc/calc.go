package calc

import (
	"context"

	calcsvc "goa.design/goa/examples/calc/gen/calc"
	goalog "goa.design/goa/logging"
)

// calc service example implementation.
// The example methods log the requests and return zero values.
type calcSvc struct {
	logger goalog.Logger
}

// Required for compatibility with Service interface
func (s *calcSvc) GetLogger() goalog.Logger {
	return s.logger
}

// NewCalc returns the calc service implementation.
func NewCalc(logger goalog.Logger) calcsvc.Service {
	return &calcSvc{logger: logger}
}

// Add implements add.
func (s *calcSvc) Add(ctx context.Context, p *calcsvc.AddPayload) (res int, err error) {
	return p.A + p.B, nil
}
