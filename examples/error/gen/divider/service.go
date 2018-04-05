// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// divider service
//
// Command:
// $ goa gen goa.design/goa/examples/error/design -o
// $(GOPATH)/src/goa.design/goa/examples/error

package dividersvc

import (
	"context"

	"goa.design/goa"
)

// Service is the divider service interface.
type Service interface {
	// IntegerDivide implements integer_divide.
	IntegerDivide(context.Context, *IntOperands) (int, error)
	// Divide implements divide.
	Divide(context.Context, *FloatOperands) (float64, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "divider"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [2]string{"integer_divide", "divide"}

// IntOperands is the payload type of the divider service integer_divide method.
type IntOperands struct {
	// Left operand
	A int
	// Right operand
	B int
}

// FloatOperands is the payload type of the divider service divide method.
type FloatOperands struct {
	// Left operand
	A float64
	// Right operand
	B float64
}

// MakeDivByZero builds a goa.ServiceError from an error.
func MakeDivByZero(err error) *goa.ServiceError {
	return &goa.ServiceError{
		Name:    "div_by_zero",
		ID:      goa.NewErrorID(),
		Message: err.Error(),
	}
}

// MakeTimeout builds a goa.ServiceError from an error.
func MakeTimeout(err error) *goa.ServiceError {
	return &goa.ServiceError{
		Name:      "timeout",
		ID:        goa.NewErrorID(),
		Message:   err.Error(),
		Temporary: true,
		Timeout:   true,
	}
}

// MakeHasRemainder builds a goa.ServiceError from an error.
func MakeHasRemainder(err error) *goa.ServiceError {
	return &goa.ServiceError{
		Name:    "has_remainder",
		ID:      goa.NewErrorID(),
		Message: err.Error(),
	}
}
