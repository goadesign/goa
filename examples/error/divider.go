package divider

import (
	"context"
	"fmt"

	dividersvc "goa.design/goa/examples/error/gen/divider"
	goalog "goa.design/goa/logging"
)

// divider service example implementation.
// The example methods log the requests and return zero values.
type dividerSvc struct {
	logger goalog.Logger
}

// Required for compatibility with Service interface
func (s *dividerSvc) GetLogger() goalog.Logger {
	return s.logger
}

// NewDivider returns the divider service implementation.
func NewDivider(logger goalog.Logger) dividersvc.Service {
	return &dividerSvc{logger: logger}
}

// IntegerDivide implements integer_divide.
func (s *dividerSvc) IntegerDivide(ctx context.Context, p *dividersvc.IntOperands) (res int, err error) {
	if p.B == 0 {
		return 0, dividersvc.MakeDivByZero(fmt.Errorf("right operand cannot be 0"))
	}
	if p.A%p.B != 0 {
		return 0, dividersvc.MakeHasRemainder(fmt.Errorf("remainder is %d", p.A%p.B))
	}
	return p.A / p.B, nil
}

// Divide implements divide.
func (s *dividerSvc) Divide(ctx context.Context, p *dividersvc.FloatOperands) (res float64, err error) {
	if p.B == 0 {
		return 0, dividersvc.MakeDivByZero(fmt.Errorf("right operand cannot be 0"))
	}
	return p.A / p.B, nil
}
