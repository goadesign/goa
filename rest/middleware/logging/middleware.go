package logging

import (
	"net"
	"net/http"
	"time"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/rest"
)

// New returns a middleware that logs incoming requests and outgoing responses.
// If dump is true the middleware logs the request and response bodies.
func New(logger goa.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				startedAt = time.Now()
			)

			logger.Info(
				r.Context(),
				r.Method, r.URL.String(),
				"from", from(r),
			)

			rw := rest.WrapResponseWriter(w)
			h.ServeHTTP(rw, r)

			logger.Info(
				r.Context(),
				"status", rw.StatusCode,
				"bytes", rw.ContentLength,
				"time", time.Since(startedAt).String(),
			)
		})
	}
}

// from makes a best effort to compute the request client IP.
func from(req *http.Request) string {
	if f := req.Header.Get("X-Forwarded-For"); f != "" {
		return f
	}
	f := req.RemoteAddr
	ip, _, err := net.SplitHostPort(f)
	if err != nil {
		return f
	}
	return ip
}
