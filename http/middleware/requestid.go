package middleware

import (
	"context"
	"net/http"
)

type (
	// RequestIDOption uses a constructor pattern to customize middleware
	RequestIDOption func(*requestIDOption) *requestIDOption

	// requestIDOption is the struct storing all the options.
	requestIDOption struct {
		// useXRequestIDHeader is true to use incoming "X-Request-Id" headers,
		// instead of always generating unique IDs, when present in request.
		// defaults to always-generate.
		useXRequestIDHeader bool
		// xRequestHeaderLimit is positive to truncate incoming "X-Request-Id"
		// headers at the specified length. defaults to no limit.
		xRequestHeaderLimit int
	}
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
func RequestID(options ...RequestIDOption) func(http.Handler) http.Handler {
	o := new(requestIDOption)
	for _, option := range options {
		o = option(o)
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var id string
			if o.useXRequestIDHeader {
				id = r.Header.Get("X-Request-Id")
				if o.xRequestHeaderLimit > 0 && len(id) > o.xRequestHeaderLimit {
					id = id[:o.xRequestHeaderLimit]
				} else if id == "" {
					id = shortID()
				}
			} else {
				id = shortID()
			}
			ctx := context.WithValue(r.Context(), RequestIDKey, id)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UseXRequestIDHeaderOption enables/disables using "X-Request-Id" header.
func UseXRequestIDHeaderOption(f bool) RequestIDOption {
	return func(o *requestIDOption) *requestIDOption {
		o.useXRequestIDHeader = f
		return o
	}
}

// XRequestHeaderLimitOption sets the option for using "X-Request-Id" header.
func XRequestHeaderLimitOption(limit int) RequestIDOption {
	return func(o *requestIDOption) *requestIDOption {
		o.xRequestHeaderLimit = limit
		return o
	}
}
