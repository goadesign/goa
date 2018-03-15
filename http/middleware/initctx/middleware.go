/*
Package initctx provides a middleware that initializes the request context.
*/
package initctx

import (
	"context"
	"net/http"
)

// New returns a middleware which initializes the request context.
func New(ctx context.Context) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
