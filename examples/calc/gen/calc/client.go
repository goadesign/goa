// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// calc client
//
// Command:
// $ goa gen goa.design/goa/examples/calc/design

package calcsvc

import (
	"context"

	goa "goa.design/goa"
)

// Client is the "calc" service client.
type Client struct {
	AddEndpoint   goa.Endpoint
	AddedEndpoint goa.Endpoint
}

// NewClient initializes a "calc" service client given the endpoints.
func NewClient(add, added goa.Endpoint) *Client {
	return &Client{
		AddEndpoint:   add,
		AddedEndpoint: added,
	}
}

// Add calls the "add" endpoint of the "calc" service.
func (c *Client) Add(ctx context.Context, p *AddPayload) (res int, err error) {
	var ires interface{}
	ires, err = c.AddEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(int), nil
}

// Added calls the "added" endpoint of the "calc" service.
func (c *Client) Added(ctx context.Context, p map[string][]int) (res int, err error) {
	var ires interface{}
	ires, err = c.AddedEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(int), nil
}
