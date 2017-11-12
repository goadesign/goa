// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage endpoints
//
// Command:
// $ goa gen goa.design/goa/examples/cellar/design

package storage

import (
	"context"

	goa "goa.design/goa"
)

// Endpoints wraps the "storage" service endpoints.
type Endpoints struct {
	List   goa.Endpoint
	Show   goa.Endpoint
	Add    goa.Endpoint
	Remove goa.Endpoint
}

// NewEndpoints wraps the methods of a storage service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		List:   NewListEndpoint(s),
		Show:   NewShowEndpoint(s),
		Add:    NewAddEndpoint(s),
		Remove: NewRemoveEndpoint(s),
	}
}

// NewListEndpoint returns an endpoint function that calls method "list" of
// service "storage".
func NewListEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.List(ctx)
	}
}

// NewShowEndpoint returns an endpoint function that calls method "show" of
// service "storage".
func NewShowEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*ShowPayload)
		return s.Show(ctx, p)
	}
}

// NewAddEndpoint returns an endpoint function that calls method "add" of
// service "storage".
func NewAddEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*Bottle)
		return s.Add(ctx, p)
	}
}

// NewRemoveEndpoint returns an endpoint function that calls method "remove" of
// service "storage".
func NewRemoveEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*RemovePayload)
		return nil, s.Remove(ctx, p)
	}
}
