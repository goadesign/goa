// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// calc service
//
// Command:
// $ goa gen goa.design/goa/examples/calc/design -o
// $(GOPATH)/src/goa.design/goa/examples/calc

package calcsvc

import (
	"context"
)

// The calc service performs operations on numbers
type Service interface {
	// Add implements add.
	Add(context.Context, *AddPayload) (int, error)
}

// AddPayload is the payload type of the calc service add method.
type AddPayload struct {
	// Left operand
	A int
	// Right operand
	B int
}
