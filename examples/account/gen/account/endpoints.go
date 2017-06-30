// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// account endpoints
//
// Command:
// $ goa server goa.design/goa.v2/examples/account/design

package account

import (
	"context"

	goa "goa.design/goa.v2"
)

type (
	// Endpoints wraps the account service endpoints.
	Endpoints struct {
		Create goa.Endpoint
		List   goa.Endpoint
		Show   goa.Endpoint
		Delete goa.Endpoint
	}
)

// NewEndpoints wraps the methods of a account service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	ep := new(Endpoints)

	ep.Create = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*CreateAccount)
		return s.Create(ctx, p)
	}

	ep.List = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*ListAccount)
		return s.List(ctx, p)
	}

	ep.Show = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*ShowPayload)
		return s.Show(ctx, p)
	}

	ep.Delete = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*DeletePayload)
		return nil, s.Delete(ctx, p)
	}

	return ep
}
