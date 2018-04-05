// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// divider client
//
// Command:
// $ goa gen goa.design/goa/examples/error/design -o
// $(GOPATH)/src/goa.design/goa/examples/error

package dividersvc

import (
	"context"

	goa "goa.design/goa"
)

// Client is the "divider" service client.
type Client struct {
	IntegerDivideEndpoint goa.Endpoint
	DivideEndpoint        goa.Endpoint
}

// NewClient initializes a "divider" service client given the endpoints.
func NewClient(integerDivide, divide goa.Endpoint) *Client {
	return &Client{
		IntegerDivideEndpoint: integerDivide,
		DivideEndpoint:        divide,
	}
}

// IntegerDivide calls the "integer_divide" endpoint of the "divider" service.
// IntegerDivide may return the following errors:
//	- "has_remainder" (type *goa.ServiceError): integer division has remainder
//	- error: internal error
func (c *Client) IntegerDivide(ctx context.Context, p *IntOperands) (res int, err error) {
	var ires interface{}
	ires, err = c.IntegerDivideEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(int), nil
}

// Divide calls the "divide" endpoint of the "divider" service.
func (c *Client) Divide(ctx context.Context, p *FloatOperands) (res float64, err error) {
	var ires interface{}
	ires, err = c.DivideEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(float64), nil
}
