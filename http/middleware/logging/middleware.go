package logging

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"time"

	"goa.design/goa.v2"
	"goa.design/goa.v2/http/middleware/tracing"
)

// New returns a middleware that logs short messages for incoming requests and
// outgoing responses.
func New(logger goa.LogAdapter) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := tracing.ContextTraceID(r.Context())
			if reqID == "" {
				reqID = shortID()
			}
			started := time.Now()

			logger.Info(r.Context(),
				"id", reqID,
				r.Method, r.URL.String(),
				"from", from(r))

			rw := CaptureResponse(w)
			h.ServeHTTP(rw, r)

			logger.Info(r.Context(),
				"id", reqID,
				"status", rw.StatusCode,
				"bytes", rw.ContentLength,
				"time", time.Since(started).String())
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

// shortID produces a "unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}
