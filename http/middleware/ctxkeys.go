package middleware

type (
	// private type used to define context keys
	ctxKey int
)

const (
	// RequestIDKey is the request context key used to store the request ID
	// created by the RequestID middleware.
	RequestIDKey ctxKey = iota + 1

	// TraceIDKey is the request context key used to store the current Trace
	// ID if any.
	TraceIDKey

	// TraceSpanIDKey is the request context key used to store the current
	// trace span ID if any.
	TraceSpanIDKey

	// TraceParentSpanIDKey is the request context key used to store the current
	// trace parent span ID if any.
	TraceParentSpanIDKey
)
