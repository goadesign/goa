package middleware

import (
	"context"
	"net/http"
)

// RequestID returns a middleware which initializes the context with a unique value under the
// RequestIDKey key.
func RequestID() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), RequestIDKey, shortID())
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
