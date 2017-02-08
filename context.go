package goa

import "golang.org/x/net/context"

// Keys used to store data in context.
const (
	endpointKey key = iota + 1
	serviceKey
)

// key is the type used to store internal values in the context.
// Context provides typed accessor methods to these values.
type key int

// NewContext builds a new goa request context. The context contains the service
// and endpoint names as set by the generated code. This can be leveraged by
// middlewares such as the security middleware to implement cross concern
// behavior.
func NewContext(ctx context.Context, s, e string) context.Context {
	ctx = context.WithValue(ctx, serviceKey, s)
	ctx = context.WithValue(ctx, endpointKey, e)

	return ctx
}

// ContextService extracts the service name from the given context.
func ContextService(ctx context.Context) string {
	if c := ctx.Value(serviceKey); c != nil {
		return c.(string)
	}
	return ""
}

// ContextEndpoint extracts the endpoint name from the given context.
func ContextEndpoint(ctx context.Context) string {
	if a := ctx.Value(endpointKey); a != nil {
		return a.(string)
	}
	return ""
}
