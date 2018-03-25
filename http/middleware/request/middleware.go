package request

import (
	"net/http"
	"context"
)

func New() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "request", r)
			req := r.WithContext(ctx)

			h.ServeHTTP(w, req)
		})
	}
}