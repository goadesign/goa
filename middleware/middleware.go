package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/goadesign/goa"

	"golang.org/x/net/context"
)

// ReqIDKey is the context key used by the RequestID middleware to store the request ID value.
const ReqIDKey middlewareKey = 1

// RequestIDHeader is the name of the header used to transmit the request ID.
const RequestIDHeader = "X-Request-Id"

// Counter used to create new request ids.
var reqID int64

// Common prefix to all newly created request ids for this process.
var reqPrefix string

// Initialize common prefix on process startup.
func init() {
	// algorithm taken from https://github.com/zenazn/goji/blob/master/web/middleware/request_id.go#L44-L50
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}
	reqPrefix = string(b64[0:10])
}

// middlewareKey is the private type used for goa middlewares to store values in the context.
// It is private to avoid possible collisions with keys used by other packages.
type middlewareKey int

// LogRequest creates a request logger middleware.
// This middleware is aware of the RequestID middleware and if registered after it leverages the
// request ID for logging.
// If verbose is true then the middlware logs the request and response bodies.
func LogRequest(verbose bool) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			reqID := ctx.Value(ReqIDKey)
			if reqID == nil {
				reqID = shortID()
			}
			ctx = goa.LogWith(ctx, "id", reqID)
			startedAt := time.Now()
			r := goa.ContextRequest(ctx)
			goa.LogInfo(ctx, "started", r.Method, r.URL.String())
			if verbose {
				if len(r.Params) > 0 {
					logCtx := make([]interface{}, 2*len(r.Params))
					i := 0
					for k, v := range r.Params {
						logCtx[i] = k
						logCtx[i+1] = interface{}(strings.Join(v, ", "))
						i = i + 2
					}
					goa.LogInfo(ctx, "params", logCtx...)
				}
				if r.ContentLength > 0 {
					if mp, ok := r.Payload.(map[string]interface{}); ok {
						logCtx := make([]interface{}, 2*len(mp))
						i := 0
						for k, v := range mp {
							logCtx[i] = k
							logCtx[i+1] = interface{}(v)
							i = i + 2
						}
						goa.LogInfo(ctx, "payload", logCtx...)
					} else {
						// Not the most efficient but this is used for debugging
						js, err := json.Marshal(r.Payload)
						if err == nil {
							js = []byte("<invalid JSON>")
						}
						goa.LogInfo(ctx, "payload", "raw", string(js))
					}
				}
			}
			err := h(ctx, rw, req)
			resp := goa.ContextResponse(ctx)
			goa.LogInfo(ctx, "completed", "status", resp.Status,
				"bytes", resp.Length, "time", time.Since(startedAt).String())
			return err
		}
	}
}

// loggingResponseWriter wraps an http.ResponseWriter and writes only raw
// response data (as text) to the context logger. assumes status and duration
// are logged elsewhere (i.e. by the LogRequest middleware).
type loggingResponseWriter struct {
	http.ResponseWriter
	ctx context.Context
}

// Write will write raw data to logger and response writer.
func (lrw *loggingResponseWriter) Write(buf []byte) (int, error) {
	goa.LogInfo(lrw.ctx, "response", "body", string(buf))
	return lrw.ResponseWriter.Write(buf)
}

// LogResponse creates a response logger middleware.
// Only Logs the raw response data without accumulating any statistics.
func LogResponse() goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			// chain a new logging writer to the current response writer.
			resp := goa.ContextResponse(ctx)
			resp.SwitchWriter(
				&loggingResponseWriter{
					ResponseWriter: resp.SwitchWriter(nil),
					ctx:            ctx,
				})

			// next
			return h(ctx, rw, req)
		}
	}
}

// RequestID is a middleware that injects a request ID into the context of each request.
// Retrieve it using ctx.Value(ReqIDKey). If the incoming request has a RequestIDHeader header then
// that value is used else a random value is generated.
func RequestID() goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			id := req.Header.Get(RequestIDHeader)
			if id == "" {
				id = fmt.Sprintf("%s-%d", reqPrefix, atomic.AddInt64(&reqID, 1))
			}
			ctx = context.WithValue(ctx, ReqIDKey, id)

			return h(ctx, rw, req)
		}
	}
}

// Recover is a middleware that recovers panics and returns an internal error response.
func Recover() goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
			defer func() {
				if r := recover(); r != nil {
					if ctx != nil {
						switch x := r.(type) {
						case string:
							err = fmt.Errorf("panic: %s", x)
						case error:
							err = x
						default:
							err = errors.New("unknown panic")
						}
						const size = 64 << 10 // 64KB
						buf := make([]byte, size)
						buf = buf[:runtime.Stack(buf, false)]
						lines := strings.Split(string(buf), "\n")
						stack := lines[3:]
						status := http.StatusInternalServerError
						var message string
						reqID := ctx.Value(ReqIDKey)
						if reqID != nil {
							message = fmt.Sprintf(
								"%s\nRefer to the following token when contacting support: %s",
								http.StatusText(status),
								reqID)
						}
						goa.LogError(ctx, "PANIC", "error", err, "stack", strings.Join(stack, "\n"))

						// note we must respond or else a 500 with "unhandled request" is the
						// default response.
						if message == "" {
							// without the logger and/or request id (from middleware) we can
							// only return the full error message for reference purposes. it
							// is unlikely to make sense to the caller unless they understand
							// the source code.
							message = err.Error()
						}
						rw.WriteHeader(status)
						rw.Write([]byte(message))
					}
				}
			}()
			return h(ctx, rw, req)
		}
	}
}

// Timeout sets a global timeout for all controller actions.
// The timeout notification is made through the context, it is the responsability of the request
// handler to handle it. For example:
//
// 	func (ctrl *Controller) DoLongRunningAction(ctx *DoLongRunningActionContext) error {
// 		action := NewLongRunning()      // setup long running action
//		c := make(chan error, 1)        // create return channel
//		go func() { c <- action.Run() } // Launch long running action goroutine
//		select {
//		case <- ctx.Done():             // timeout triggered
//			action.Cancel()         // cancel long running action
//			<-c                     // wait for Run to return.
//			return ctx.Err()        // retrieve cancel reason
//		case err := <-c:   		// action finished on time
//			return err  		// forward its return value
//		}
//	}
//
// Package golang.org/x/net/context/ctxhttp contains an implementation of an HTTP client which is
// context-aware:
//
// 	func (ctrl *Controller) HttpAction(ctx *HttpActionContext) error {
//		req, err := http.NewRequest("GET", "http://iamaslowservice.com", nil)
//		// ...
//		resp, err := ctxhttp.Do(ctx, nil, req) // returns if timeout triggers
//		// ...
// 	}
//
// Controller actions can check if a timeout is set by calling the context Deadline method.
func Timeout(timeout time.Duration) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
			// We discard the cancel function because the goa handler already takes
			// care of canceling on completion.
			nctx, _ := context.WithTimeout(ctx, timeout)
			return h(nctx, rw, req)
		}
	}
}

// RequireHeader requires a request header to match a value pattern. If the
// header is missing or does not match then the failureStatus is the response
// (e.g. http.StatusUnauthorized). If pathPattern is nil then any path is
// included. If requiredHeaderValue is nil then any value is accepted so long as
// the header is non-empty.
func RequireHeader(
	pathPattern *regexp.Regexp,
	requiredHeaderName string,
	requiredHeaderValue *regexp.Regexp,
	failureStatus int) goa.Middleware {

	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
			if pathPattern == nil || pathPattern.MatchString(req.URL.Path) {
				matched := false
				headerValue := req.Header.Get(requiredHeaderName)
				if len(headerValue) > 0 {
					if requiredHeaderValue == nil {
						matched = true
					} else {
						matched = requiredHeaderValue.MatchString(headerValue)
					}
				}
				if matched {
					err = h(ctx, rw, req)
				} else {
					resp := goa.ContextResponse(ctx)
					err = resp.Send(ctx, failureStatus, http.StatusText(failureStatus))
				}
			} else {
				err = h(ctx, rw, req)
			}
			return
		}
	}
}

// shortID produces a "unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.StdEncoding.EncodeToString(b)
}
