package middleware

// middlewareKey is the private type used for goa middlewares to store values in the context.
// It is private to avoid possible collisions with keys used by other packages.
type middlewareKey int

const (
	// ReqIDKey is the context key used by the RequestID middleware to store the request ID value.
	reqIDKey middlewareKey = iota + 1

	// Keys used to record trace in context.
	traceKey
	spanKey
	parentSpanKey
)
