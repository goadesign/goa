package middleware

type (
	// private type used to define context keys
	ctxKey int
)

const (
	// RequestIDKey is the request context key used to store the request ID
	// created by the RequestID middleware.
	RequestIDKey ctxKey = iota + 1
)
