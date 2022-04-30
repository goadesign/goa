package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	"goa.design/goa/v3/middleware"
)

// Log returns a middleware that logs incoming HTTP requests and outgoing
// responses. The middleware uses the request ID set by the RequestID middleware
// or creates a short unique request ID if missing for each incoming request and
// logs it with the request and corresponding response details.
//
// The middleware logs the incoming requests HTTP method and path as well as the
// originator of the request. The originator is computed by looking at the
// X-Forwarded-For HTTP header or - absent of that - the originating IP. The
// middleware also logs the response HTTP status code, body length (in bytes) and
// timing information.
func Log(l middleware.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log(l, r, w, h)
		})
	}
}

// LogContext returns a middleware that logs the incoming requests similarly to
// Log. LogContext calls the given function with the request context to extract
// the logger.
func LogContext(logFromCtx func(context.Context) middleware.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logFromCtx(r.Context())
			if l == nil {
				h.ServeHTTP(w, r)
				return
			}
			log(l, r, w, h)
		})
	}
}

// log does the actual logging given the logger.
func log(l middleware.Logger, r *http.Request, w http.ResponseWriter, next http.Handler) {
	reqID := r.Context().Value(middleware.RequestIDKey)
	if reqID == nil {
		reqID = shortID()
	}
	started := time.Now()

	l.Log("id", reqID,
		"req", r.Method+" "+r.URL.String(),
		"from", from(r))

	rw := CaptureResponse(w)
	next.ServeHTTP(rw, r)

	l.Log("id", reqID,
		"status", rw.StatusCode,
		"bytes", rw.ContentLength,
		"time", time.Since(started).String())
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
