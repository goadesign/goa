package middleware

import (
	"context"
	"net/http"

	"goa.design/goa/v3/middleware"
)

// RequestID returns a middleware, which initializes the context with a unique
// value under the RequestIDKey key. Optionally uses the incoming "X-Request-Id"
// header, if present, with or without a length limit to use as request ID. the
// default behavior is to always generate a new ID.
//
// examples of use:
//  service.Use(middleware.RequestID())
//
//  // enable options for using "X-Request-Id" header with length limit.
//  service.Use(middleware.RequestID(
//    middleware.UseXRequestIDHeaderOption(true),
//    middleware.XRequestHeaderLimitOption(128)))
func RequestID(options ...middleware.RequestIDOption) func(http.Handler) http.Handler {
	o := middleware.NewRequestIDOptions(options...)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if o.IsUseRequestID() {
				if id := r.Header.Get("X-Request-Id"); id != "" {
					ctx = context.WithValue(ctx, middleware.RequestIDKey, id)
				}
			}
			ctx = middleware.GenerateRequestID(ctx, o)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UseXRequestIDHeaderOption enables/disables using "X-Request-Id" header.
func UseXRequestIDHeaderOption(f bool) middleware.RequestIDOption {
	return middleware.UseRequestIDOption(f)
}

// XRequestHeaderLimitOption sets the option for using "X-Request-Id" header.
func XRequestHeaderLimitOption(limit int) middleware.RequestIDOption {
	return middleware.RequestIDLimitOption(limit)
}
