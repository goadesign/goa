package goa

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
)

type (
	// Middleware represents the canonical goa middleware signature.
	Middleware func(Handler) Handler
)

// NewMiddleware creates a middleware from the given argument. The allowed types for the
// argument are:
//
// - a goa middleware: goa.Middleware or func(goa.Handler) goa.Handler
//
// - a goa handler: goa.Handler or func(*goa.Context) error
//
// - an http middleware: func(http.Handler) http.Handler
//
// - or an http handler: http.Handler or func(http.ResponseWriter, *http.Request)
//
// An error is returned if the given argument is not one of the types above.
func NewMiddleware(m interface{}) (mw Middleware, err error) {
	switch m := m.(type) {
	case Middleware:
		mw = m
	case func(Handler) Handler:
		mw = m
	case Handler:
		mw = handlerToMiddleware(m)
	case func(*Context) error:
		mw = handlerToMiddleware(m)
	case func(http.Handler) http.Handler:
		mw = func(h Handler) Handler {
			return func(ctx *Context) (err error) {
				rw := ctx.Value(respKey).(http.ResponseWriter)
				m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					err = h(ctx)
				})).ServeHTTP(rw, ctx.Request())
				return
			}
		}
	case http.Handler:
		mw = httpHandlerToMiddleware(m.ServeHTTP)
	case func(http.ResponseWriter, *http.Request):
		mw = httpHandlerToMiddleware(m)
	default:
		err = fmt.Errorf("invalid middleware %#v", m)
	}
	return
}

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
func LogRequest() Middleware {
	return func(h Handler) Handler {
		return func(ctx *Context) error {
			reqID := ctx.Value(ReqIDKey)
			if reqID == nil {
				reqID = shortID()
			}
			ctx.Logger = ctx.Logger.New("id", reqID)
			startedAt := time.Now()
			r := ctx.Value(reqKey).(*http.Request)
			ctx.Info("started", r.Method, r.URL.String())
			params := ctx.Value(paramsKey).(url.Values)
			if len(params) > 0 {
				logCtx := make(log.Ctx, len(params))
				for k, v := range params {
					logCtx[k] = interface{}(v)
				}
				ctx.Debug("params", logCtx)
			}
			payload := ctx.Value(payloadKey)
			if r.ContentLength > 0 {
				if mp, ok := payload.(map[string]interface{}); ok {
					ctx.Debug("payload", log.Ctx(mp))
				} else {
					ctx.Debug("payload", "raw", payload)
				}
			}
			err := h(ctx)
			ctx.Info("completed", "status", ctx.ResponseStatus(),
				"bytes", ctx.ResponseLength(), "time", time.Since(startedAt).String())
			return err
		}
	}
}

// RequestID is a middleware that injects a request ID into the context of each request.
// Retrieve it using ctx.Value(ReqIDKey). If the incoming request has a RequestIDHeader header then
// that value is used else a random value is generated.
func RequestID() Middleware {
	return func(h Handler) Handler {
		return func(ctx *Context) error {
			id := ctx.Request().Header.Get(RequestIDHeader)
			if id == "" {
				id = fmt.Sprintf("%s-%d", reqPrefix, atomic.AddInt64(&reqID, 1))
			}
			ctx.SetValue(ReqIDKey, id)

			return h(ctx)
		}
	}
}

// Recover is a middleware that recovers panics and returns an internal error response.
func Recover() Middleware {
	return func(h Handler) Handler {
		return func(ctx *Context) (err error) {
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
						if ctx.Logger != nil {
							reqID := ctx.Value(ReqIDKey)
							if reqID != nil {
								message = fmt.Sprintf(
									"%s\nRefer to the following token when contacting support: %s",
									http.StatusText(status),
									reqID)
							}
							ctx.Logger.Error("panic", "err", err, "stack", stack)
						}

						// note we must respond or else a 500 with "unhandled request" is the
						// default response.
						if message == "" {
							// without the logger and/or request id (from middleware) we can
							// only return the full error message for reference purposes. it
							// is unlikely to make sense to the caller unless they understand
							// the source code.
							message = err.Error()
						}
						ctx.Respond(status, []byte(message))
					}
				}
			}()
			return h(ctx)
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
func Timeout(timeout time.Duration) Middleware {
	return func(h Handler) Handler {
		return func(ctx *Context) (err error) {
			// We discard the cancel function because the goa handler already takes
			// care of canceling on completion.
			ctx.Context, _ = context.WithTimeout(ctx.Context, timeout)
			return h(ctx)
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
	failureStatus int) Middleware {

	return func(h Handler) Handler {
		return func(ctx *Context) (err error) {
			if pathPattern == nil || pathPattern.MatchString(ctx.Request().URL.Path) {
				matched := false
				header := ctx.Request().Header
				headerValue := header.Get(requiredHeaderName)
				if len(headerValue) > 0 {
					if requiredHeaderValue == nil {
						matched = true
					} else {
						matched = requiredHeaderValue.MatchString(headerValue)
					}
				}
				if matched {
					err = h(ctx)
				} else {
					err = ctx.Respond(failureStatus, []byte(http.StatusText(failureStatus)))
				}
			} else {
				err = h(ctx)
			}
			return
		}
	}
}

// These compression constants are copied from the compress/gzip package.
const (
	encodingGzip = "gzip"

	headerAcceptEncoding  = "Accept-Encoding"
	headerContentEncoding = "Content-Encoding"
	headerContentLength   = "Content-Length"
	headerContentType     = "Content-Type"
	headerVary            = "Vary"
	headerSecWebSocketKey = "Sec-WebSocket-Key"
)

// gzipResponseWriter wraps the http.ResponseWriter to provide gzip
// capabilities.
type gzipResponseWriter struct {
	http.ResponseWriter
	gzw *gzip.Writer
}

// Write writes bytes to the gzip.Writer. It will also set the Content-Type
// header using the net/http library content type detection if the Content-Type
// header was not set yet.
func (grw gzipResponseWriter) Write(b []byte) (int, error) {
	if len(grw.ResponseWriter.Header().Get(headerContentType)) == 0 {
		grw.ResponseWriter.Header().Set(headerContentType, http.DetectContentType(b))
	}
	return grw.gzw.Write(b)
}

// handler struct contains the ServeHTTP method
type handler struct {
	pool sync.Pool
}

// Gzip encodes the response using Gzip encoding and sets all the appropriate
// headers. If the Content-Type is not set, it will be set by calling
// http.DetectContentType on the data being written.
func Gzip(level int) Middleware {
	gzipPool := sync.Pool{
		New: func() interface{} {
			gz, err := gzip.NewWriterLevel(ioutil.Discard, level)
			if err != nil {
				panic(err)
			}
			return gz
		},
	}
	return func(h Handler) Handler {
		return func(ctx *Context) (err error) {
			w := ctx.Value(respKey).(http.ResponseWriter)
			r := ctx.Request()
			// Skip compression if the client doesn't accept gzip encoding, is
			// requesting a WebSocket or the data is already compressed.
			if !strings.Contains(r.Header.Get(headerAcceptEncoding), encodingGzip) ||
				len(r.Header.Get(headerSecWebSocketKey)) > 0 ||
				w.Header().Get(headerContentEncoding) == encodingGzip {
				return h(ctx)
			}

			// Retrieve gzip writer from the pool. Reset it to use the ResponseWriter.
			// This allows us to re-use an already allocated buffer rather than
			// allocating a new buffer for every request.
			gz := gzipPool.Get().(*gzip.Writer)
			gz.Reset(w)

			// Set the appropriate gzip headers.
			headers := w.Header()
			headers.Set(headerContentEncoding, encodingGzip)
			headers.Set(headerVary, headerAcceptEncoding)

			// Wrap the original http.ResponseWriter with our gzipResponseWriter
			grw := gzipResponseWriter{
				ResponseWriter: w,
				gzw:            gz,
			}
			ctx.SetValue(respKey, grw)

			// Call the next handler supplying the gzipResponseWriter instead of
			// the original.
			err = h(ctx)
			if err != nil {
				return
			}

			// Delete the content length after we know we have been written to.
			grw.ResponseWriter.Header().Del(headerContentLength)
			gz.Close()
			gzipPool.Put(gz)
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

// handlerToMiddleware creates a middleware from a raw handler.
// The middleware calls the handler and either breaks the middleware chain if the handler returns
// an error by also returning the error or calls the next handler in the chain otherwise.
func handlerToMiddleware(m Handler) Middleware {
	return func(h Handler) Handler {
		return func(ctx *Context) error {
			if err := m(ctx); err != nil {
				return err
			}
			return h(ctx)
		}
	}
}

// httpHandlerToMiddleware creates a middleware from a http.HandlerFunc.
// The middleware calls the ServerHTTP method exposed by the http handler and then calls the next
// middleware in the chain.
func httpHandlerToMiddleware(m http.HandlerFunc) Middleware {
	return func(h Handler) Handler {
		return func(ctx *Context) error {
			m.ServeHTTP(ctx, ctx.Request())
			return h(ctx)
		}
	}
}
