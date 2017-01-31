package logging

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"goa.design/goa.v2"
	"goa.design/goa.v2/rest"
	"goa.design/goa.v2/rest/middleware/tracing"
)

// New returns a middleware that logs incoming requests and outgoing responses.
// If dump is true the middleware logs the request and response bodies.
func New(logger goa.Logger, dump bool) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				ctx       = r.Context()
				traceID   = tracing.ContextTraceID(ctx)
				startedAt = time.Now()
			)
			if traceID == "" {
				traceID = shortID()
			}
			logger.Log(
				"trace", traceID,
				r.Method, r.URL.String(),
				"from", from(r),
				"service", goa.ContextService(ctx),
				"endpoint", goa.ContextEndpoint(ctx),
			)
			if dump {
				if len(r.Header) > 0 {
					logCtx := make([]interface{}, 2+2*len(r.Header))
					logCtx[0] = "trace"
					logCtx[1] = traceID
					i := 2
					for k, v := range r.Header {
						logCtx[i] = k
						logCtx[i+1] = interface{}(strings.Join(v, ", "))
						i = i + 2
					}
					logger.Log(logCtx...)
				}
				req := rest.ContextRequest(ctx)
				if req != nil {
					if params := req.Params; len(params) > 0 {
						logCtx := make([]interface{}, 2+2*len(params))
						logCtx[0] = "trace"
						logCtx[1] = traceID
						i := 2
						for k, v := range params {
							logCtx[i] = k
							logCtx[i+1] = interface{}(v)
							i = i + 2
						}
						logger.Log(logCtx...)
					}
					if req.ContentLength > 0 {
						if mp, ok := req.Payload.(map[string]interface{}); ok {

							logCtx := make([]interface{}, 2+2*len(mp))
							logCtx[0] = "trace"
							logCtx[1] = traceID
							i := 0
							for k, v := range mp {
								logCtx[i] = k
								logCtx[i+1] = v
								i = i + 2
							}
							logger.Log(logCtx...)
						} else {
							// Not the most efficient but this is used for debugging
							js, err := json.Marshal(req.Payload)
							if err != nil {
								js = []byte("<invalid JSON>")
							}
							logger.Log("trace", traceID, "payload", string(js))
						}
					}
				}
			}

			h.ServeHTTP(w, r)

			resp := rest.ContextResponse(r.Context())
			if resp == nil {
				return
			}
			if code := resp.ErrorCode; code != "" {
				logger.Log("trace", traceID, "status", resp.Status, "error", code,
					"bytes", resp.Length, "time", time.Since(startedAt).String(),
					"service", goa.ContextService(ctx), "endpoint", goa.ContextEndpoint(ctx))
			} else {
				logger.Log("trace", traceID, "status", resp.Status,
					"bytes", resp.Length, "time", time.Since(startedAt).String(),
					"service", goa.ContextService(ctx), "endpoint", goa.ContextEndpoint(ctx))
			}
		})
	}
}

// shortID produces a "unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.StdEncoding.EncodeToString(b)
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
