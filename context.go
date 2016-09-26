package goa

import "context"

type (
	// key is the type used to store values in the context.
	key int
)

// Keys used to store data in context.
const (
	logKey key = iota + 1
)

// WithLogger sets the request context logger and returns the resulting new context.
func WithLogger(ctx context.Context, logger LogAdapter) context.Context {
	return context.WithValue(ctx, logKey, logger)
}

// WithLogContext instantiates a new logger by appending the given key/value pairs to the context
// logger and setting the resulting logger in the context.
func WithLogContext(ctx context.Context, keyvals ...interface{}) context.Context {
	logger := ContextLogger(ctx)
	if logger == nil {
		return ctx
	}
	nl := logger.New(keyvals...)
	return WithLogger(ctx, nl)
}

// ContextLogger extracts the logger from the given context.
func ContextLogger(ctx context.Context) LogAdapter {
	if v := ctx.Value(logKey); v != nil {
		return v.(LogAdapter)
	}
	return nil
}
