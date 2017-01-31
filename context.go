package goa

import "context"

// Private type used to store values in context.
type key int

const (
	serviceKey key = iota + 1
	endpointKey
)

// WithService creates a new context with the given value set for the name of the
// service. Use ContextService to extract the value later. This is mainly
// intended for use by generated code.
func WithService(ctx context.Context, service string) context.Context {
	return context.WithValue(ctx, serviceKey, service)
}

// WithEndpoint creates a new context with the given value set for the name of
// the endpoint. Use ContextEndpoint to extract the value later. This is mainly
// intended for use by generated code.
func WithEndpoint(ctx context.Context, endpoint string) context.Context {
	return context.WithValue(ctx, endpointKey, endpoint)
}

// ContextService extracts the name of the service from the context.
func ContextService(ctx context.Context) string {
	if n := ctx.Value(serviceKey); n != nil {
		return n.(string)
	}
	return ""
}

// ContextEndpoint extracts the name of the endpoint from the context.
func ContextEndpoint(ctx context.Context) string {
	if n := ctx.Value(endpointKey); n != nil {
		return n.(string)
	}
	return ""
}
