package middleware

var (
	// RequestIDKey is the request context key used to store the request ID
	// created by the RequestID middleware.
	RequestIDKey = "goa-request-id"

	// TraceIDKey is the request context key used to store the current Trace
	// ID if any.
	TraceIDKey = "goa-trace-id"

	// TraceSpanIDKey is the request context key used to store the current
	// trace span ID if any.
	TraceSpanIDKey = "goa-trace-span-id"

	// TraceParentSpanIDKey is the request context key used to store the current
	// trace parent span ID if any.
	TraceParentSpanIDKey = "goa-trace-parent-span-id"
)
