package goa

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
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
			return func(c *Context) (err error) {
				m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					nc := context.WithValue(c.Context, respKey, w)
					nc = context.WithValue(nc, reqKey, r)
					newC := Context{
						Context: nc,
						Logger:  c.Logger,
					}
					err = h(&newC)
				})).ServeHTTP(c.ResponseWriter(), c.Request())
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

// ReqIDKey is the RequestID middleware key used to store the request ID value in the context.
const ReqIDKey middlewareKey = 0

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
// This middleware is aware of the RequestID middleware and if registered after it leverates the
// request ID for logging.
func LogRequest() Middleware {
	return func(h Handler) Handler {
		return func(ctx *Context) error {
			reqID := ctx.Value(ReqIDKey)
			if reqID == nil {
				reqID = ShortID()
			}
			ctx.Logger = ctx.Logger.New("id", reqID)
			startedAt := time.Now()
			r := ctx.Value(reqKey).(*http.Request)
			ctx.Info("started", r.Method, r.URL.String())
			params := ctx.Value(paramKey).(map[string]string)
			if len(params) > 0 {
				logCtx := make(log.Ctx, len(params))
				for k, v := range params {
					logCtx[k] = interface{}(v)
				}
				ctx.Debug("params", logCtx)
			}
			query := ctx.Value(queryKey).(map[string][]string)
			if len(query) > 0 {
				logCtx := make(log.Ctx, len(query))
				for k, v := range query {
					logCtx[k] = interface{}(v)
				}
				ctx.Debug("query", logCtx)
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
			ctx.AddValue(ReqIDKey, id)

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
					Log.Error("panic", "err", err, "stack", stack)
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

// handlerToMiddleware creates a middleware from a raw handler.
// The middleware calls the handler and either breaks the middleware chain if the handler returns
// an error by also returning the error or calls the next handler in the chain otherwise.
func handlerToMiddleware(m Handler) Middleware {
	return func(h Handler) Handler {
		return func(c *Context) error {
			if err := m(c); err != nil {
				return err
			}
			return h(c)
		}
	}
}

// httpHandlerToMiddleware creates a middleware from a http.HandlerFunc.
// The middleware calls the ServerHTTP method exposed by the http handler and then calls the next
// middleware in the chain.
func httpHandlerToMiddleware(m http.HandlerFunc) Middleware {
	return func(h Handler) Handler {
		return func(c *Context) error {
			m.ServeHTTP(c.ResponseWriter(), c.Request())
			return h(c)
		}
	}
}
