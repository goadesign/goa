package goa

import "context"

type (
	// InterceptorInfo contains information about the request shared between
	// all interceptors in the service chain. It provides access to the service name,
	// method name, endpoint function, and request payload.
	InterceptorInfo struct {
		// Name of service handling request
		Service string
		// Name of method handling request
		Method string
		// Endpoint of request, can be used for retrying
		Endpoint Endpoint
		// Payload of request
		RawPayload any
	}

	// NextFunc is a function that will continue the request processing chain.
	NextFunc func(ctx context.Context) (any, error)
)
