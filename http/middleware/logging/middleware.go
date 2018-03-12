/*
Package logging implements a middleware that logs incoming HTTP requests and
outgoing responses. The middleware creates a short unique request ID for each
incoming request and logs it with the request and corresponding response
details.

The middleware logs the incoming requests HTTP method and path as well as the
originator of the request. The originator is computed by looking at the
X-Forwarded-For HTTP header or - absent of that - the originating IP. The
middleware also logs the response HTTP status code, body length (in bytes) and
timing information.

The package also defines the Logger interface it uses internally so that
different logger backends may be used. The default logger used by the middleware
is the Go stdlib logger.
*/
package logging

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"time"
)

// New returns a middleware that logs short messages for incoming requests and
// outgoing responses. See the package documentation for details on what is
// logged.
func New(l Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := shortID()
			started := time.Now()

			l.Log(r.Context(),
				"id", reqID,
				"req", r.Method+" "+r.URL.String(),
				"from", from(r))

			rw := CaptureResponse(w)
			h.ServeHTTP(rw, r)

			l.Log(r.Context(),
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

// shortID does a best effort to produce a "unique" 6 bytes long string
// efficiently. Do not use as a reliable way to get unique IDs.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}
