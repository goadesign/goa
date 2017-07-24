// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage endpoints
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package storage

import (
	"context"

	goa "goa.design/goa.v2"
)

type (
	// Endpoints wraps the storage service endpoints.
	Endpoints struct {
		Add    goa.Endpoint
		List   goa.Endpoint
		Show   goa.Endpoint
		Remove goa.Endpoint
	}
)

// NewEndpoints wraps the methods of a storage service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	ep := new(Endpoints)

	ep.Add = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*Bottle)
		return s.Add(ctx, p)
	}

	ep.List = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.List(ctx)
	}

	ep.Show = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*ShowPayload)
		return s.Show(ctx, p)
	}

	ep.Remove = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*RemovePayload)
		return nil, s.Remove(ctx, p)
	}

	return ep
}
