package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type (
	// Logger is the logging interface used by the middleware to produce
	// log entries.
	Logger interface {
		// Log creates a log entry using a sequence of alternating keys
		// and values.
		Log(keyvals ...interface{}) error
	}

	// adapter is a thin wrapper around the stdlib logger that adapts it to
	// the Logger interface.
	adapter struct {
		*log.Logger
	}
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
func Log(l Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Context().Value(RequestIDKey)
			if reqID == nil {
				reqID = shortID()
			}
			started := time.Now()

			l.Log("id", reqID,
				"req", r.Method+" "+r.URL.String(),
				"from", from(r))

			rw := CaptureResponse(w)
			h.ServeHTTP(rw, r)

			l.Log("id", reqID,
				"status", rw.StatusCode,
				"bytes", rw.ContentLength,
				"time", time.Since(started).String())
		})
	}
}

// NewLogger creates a Logger backed by a stdlib logger.
func NewLogger(l *log.Logger) Logger {
	return &adapter{l}
}

func (a *adapter) Log(keyvals ...interface{}) error {
	n := (len(keyvals) + 1) / 2
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "MISSING")
	}
	var fm bytes.Buffer
	vals := make([]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		vals[i/2] = v
		fm.WriteString(fmt.Sprintf(" %s=%%+v", k))
	}
	a.Logger.Printf(fm.String(), vals...)
	return nil
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
