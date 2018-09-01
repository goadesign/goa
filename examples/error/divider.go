package divider

import (
	"context"
	"fmt"
	"log"

	dividersvc "goa.design/goa/examples/error/gen/divider"
)

// divider service example implementation.
// The example methods log the requests and return zero values.
type dividerSvc struct {
	logger *log.Logger
}

// NewDivider returns the divider service implementation.
func NewDivider(logger *log.Logger) dividersvc.Service {
	return &dividerSvc{logger}
}

// IntegerDivide implements integer_divide.
func (s *dividerSvc) IntegerDivide(ctx context.Context, p *dividersvc.IntOperands) (int, error) {
	if p.B == 0 {
		return 0, dividersvc.MakeDivByZero(fmt.Errorf("right operand cannot be 0"))
	}
	if p.A%p.B != 0 {
		return 0, dividersvc.MakeHasRemainder(fmt.Errorf("remainder is %d", p.A%p.B))
	}
	return p.A / p.B, nil
}

// Divide implements divide.
func (s *dividerSvc) Divide(ctx context.Context, p *dividersvc.FloatOperands) (float64, error) {
	if p.B == 0 {
		return 0, dividersvc.MakeDivByZero(fmt.Errorf("right operand cannot be 0"))
	}
	return p.A / p.B, nil
}
