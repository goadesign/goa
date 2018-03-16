package middleware

import (
	"context"
	"net/http"
)

// RequestContext returns a middleware which initializes the request context.
func RequestContext(ctx context.Context) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestContextKeyVals returns a middleware which adds the given key/value pairs to the
// request context.
func RequestContextKeyVals(keyvals ...interface{}) func(http.Handler) http.Handler {
	if len(keyvals)%2 != 0 {
		panic("initctx: invalid number of key/value elements, must be an even number")
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for i := 0; i < len(keyvals); i += 2 {
				r = r.WithContext(context.WithValue(r.Context(), keyvals[i], keyvals[i+1]))
			}
			h.ServeHTTP(w, r)
		})
	}
}
