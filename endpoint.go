package goa

import "context"

const (
	// MethodKey is the request context key used to store the name of the
	// method as defined in the design. The generated transport code
	// initializes the corresponding value prior to invoking the endpoint.
	MethodKey contextKey = iota + 1

	// ServiceKey is the request context key used to store the name of the
	// service as defined in the design. The generated transport code
	// initializes the corresponding value prior to invoking the endpoint.
	ServiceKey
)

type (
	// Contextualizer defines the behavior of something in a context.
	Contextualizer interface {
		// Context returns the underlying context.
		Context() context.Context
		// SetContext updates the underlying context with the given context.
		SetContext(context.Context)
	}

	// Endpoint exposes service methods to remote clients independently of the
	// underlying transport.
	Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

	// private type used to define context keys.
	contextKey int
)
