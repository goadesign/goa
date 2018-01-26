// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage endpoints
//
// Command:
// $ goa gen goa.design/goa/examples/cellar/design -o
// $(GOPATH)/src/goa.design/goa/examples/cellar

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
	Rate   goa.Endpoint
}

// NewEndpoints wraps the methods of the "storage" service with endpoints.
func NewEndpoints(s Service) *Endpoints {
	return &Endpoints{
		List:   NewListEndpoint(s),
		Show:   NewShowEndpoint(s),
		Add:    NewAddEndpoint(s),
		Remove: NewRemoveEndpoint(s),
		Rate:   NewRateEndpoint(s),
	}
}

// NewListEndpoint returns an endpoint function that calls the method "list" of
// service "storage".
func NewListEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.List(ctx)
	}
}

// NewShowEndpoint returns an endpoint function that calls the method "show" of
// service "storage".
func NewShowEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*ShowPayload)
		return s.Show(ctx, p)
	}
}

// NewAddEndpoint returns an endpoint function that calls the method "add" of
// service "storage".
func NewAddEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*Bottle)
		return s.Add(ctx, p)
	}
}

// NewRemoveEndpoint returns an endpoint function that calls the method
// "remove" of service "storage".
func NewRemoveEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*RemovePayload)
		return nil, s.Remove(ctx, p)
	}
}

// NewRateEndpoint returns an endpoint function that calls the method "rate" of
// service "storage".
func NewRateEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(map[uint32][]string)
		return nil, s.Rate(ctx, p)
	}
}
