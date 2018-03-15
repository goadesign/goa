package goa

import "context"

const (
	// ContextKeyMethod is the name of the context key used to store the
	// name of the method as defined in the design. The generated transport
	// code initializes the corresponding value prior to invoking the
	// endpoint.
	ContextKeyMethod contextKey = iota + 1

	// ContextKeyService is the name of the context key used to store the
	// name of the service as defined in the design. The generated transport
	// code initializes the corresponding value prior to invoking the
	// endpoint.
	ContextKeyService
)

type (
	// private type used to define context keys.
	contextKey int
)

// Endpoint exposes service methods to remote clients independently of the
// underlying transport.
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)
