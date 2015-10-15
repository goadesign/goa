package goa

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/julienschmidt/httprouter"
	log "gopkg.in/inconshreveable/log15.v2"
)

type (
	// Application consists of a set of controllers.
	// A controller implements a resource action.
	Application struct {
		log.Logger                      // Application logger
		Name         string             // Application name
		ErrorHandler ErrorHandler       // Application error handler
		Middleware   []Middleware       // Middleware chain
		Router       *httprouter.Router // Application router
	}

	// Handler defines the goa action handler signatures.
	// Handlers accept a context and return an error.
	// Errors returned by handlers cause the application error handler to write the response.
	Handler func(*Context) error

	// ErrorHandler handles errors returned by action handlers and middleware.
	ErrorHandler func(*Context, error)
)

var (
	// Log is the global logger from which other loggers (e.g. request specific loggers) are
	// derived. Configure it by setting its handler prior to calling New.
	// See https://godoc.org/github.com/inconshreveable/log15
	Log log.Logger
)

// Log to STDOUT by default
func init() {
	Log = log.New()
	Log.SetHandler(log.StdoutHandler)
}

// New instantiates an application with the given name.
func New(name string) *Application {
	return &Application{
		Logger:       Log.New("app", name),
		Name:         name,
		ErrorHandler: DefaultErrorHandler,
		Router:       httprouter.New(),
	}
}

// Use adds a middleware to the middleware chain.
// See NewMiddleware for the list of possible types for middleware.
// middleware.go provides a set of common middleware.
func (a *Application) Use(middleware interface{}) {
	m, err := NewMiddleware(middleware)
	if err != nil {
		Fatal("invalid middleware", "middleware", middleware, "err", err)
	}
	a.Middleware = append(a.Middleware, m)
}

// SetErrorHandler defines an application wide error handler.
// The default error handler returns a 500 status code with the error message in the response body.
// TerseErrorHandler provides an alternative implementation that does not send the error message
// in the response body for internal errors (e.g. for production).
// Set it with SetErrorHandler(TerseErrorHandler).
func (a *Application) SetErrorHandler(handler ErrorHandler) {
	a.ErrorHandler = handler
}

// Run starts the application loop and sets up a listener on the given host/port.
// It logs an error and exits the process with status 1 if the HTTP server fails to start (e.g.
// listen port busy).
func (a *Application) Run(addr string) {
	a.Info("listen", "addr", addr)
	if err := http.ListenAndServe(addr, a.Router); err != nil {
		Fatal("startup failed", "err", err)
	}
}

// DefaultErrorHandler returns a 400 response for request validation errors (instances of
// BadRequestError) and a 500 response for other errors. It writes the error message to the
// response body in both cases.
func DefaultErrorHandler(c *Context, e error) {
	status := 500
	if _, ok := e.(*BadRequestError); ok {
		c.ResponseHeader().Set("Content-Type", "application/json")
		status = 400
	}
	if err := c.Respond(status, []byte(e.Error())); err != nil {
		Log.Error("failed to send default error handler response", "err", err)
	}
}

// TerseErrorHandler behaves like DefaultErrorHandler except that it does not set the response
// body for internal errors.
func TerseErrorHandler(c *Context, e error) {
	status := 500
	var body []byte
	if _, ok := e.(*BadRequestError); ok {
		c.ResponseHeader().Set("Content-Type", "application/json")
		status = 400
		body = []byte(e.Error())
	}
	if err := c.Respond(status, body); err != nil {
		Log.Error("failed to send terse error handler response", "err", err)
	}
}

// NewHTTPRouterHandle returns a httprouter handle which initializes a new context using the HTTP
// request state and calls the given handler with it.
func NewHTTPRouterHandle(app *Application, resName string, h Handler) httprouter.Handle {
	logger := app.Logger.New("ctrl", resName)
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Log started event
		startedAt := time.Now()
		id := ShortID()
		logger.Info("started", "id", id, r.Method, r.URL.String())

		// Collect URL and query string parameters
		params := make(map[string]string, len(p))
		for _, param := range p {
			params[param.Key] = param.Value
		}
		q := r.URL.Query()
		query := make(map[string][]string, len(q))
		for name, value := range q {
			query[name] = value
		}

		// Load body if any
		var payload interface{}
		var err error
		if r.ContentLength > 0 {
			decoder := json.NewDecoder(r.Body)
			err = decoder.Decode(&payload)
		}

		// Build context
		gctx, cancel := context.WithCancel(context.Background())
		defer cancel() // Signal completion of request to any child goroutine
		gctx = context.WithValue(context.Background(), reqKey, r)
		gctx = context.WithValue(gctx, respKey, w)
		gctx = context.WithValue(gctx, paramKey, params)
		gctx = context.WithValue(gctx, queryKey, query)
		gctx = context.WithValue(gctx, payloadKey, payload)
		ctx := &Context{
			Context: gctx,
			Logger:  logger.New("id", id),
		}

		// Setup middleware
		middleware := app.Middleware
		ml := len(middleware)
		for i := range middleware {
			h = middleware[ml-i-1](h)
		}

		// Log request details if needed
		if len(params) > 0 {
			ctx.Debug("params", ToLogCtx(params))
		}
		if len(query) > 0 {
			ctx.Debug("query", ToLogCtxA(query))
		}
		if err != nil {
			ctx.Respond(400, []byte(fmt.Sprintf(`{"kind":"invalid request","msg":"invalid JSON: %s"}`, err)))
			goto end
		}
		if r.ContentLength > 0 {
			if mp, ok := payload.(map[string]interface{}); ok {
				ctx.Debug("payload", log.Ctx(mp))
			} else {
				ctx.Debug("payload", "raw", payload)
			}
		}

		// Call middleware and user controller handler
		if err := h(ctx); err != nil {
			app.ErrorHandler(ctx, err)
		}

		// We're done
	end:
		log.Info("completed", "id", id, "status", ctx.ResponseStatus(),
			"bytes", ctx.ResponseLength(), "time", time.Since(startedAt).String())
	}
}

// ShortID produces a "unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func ShortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.StdEncoding.EncodeToString(b)
}

// Fatal logs a critical message and exits the process with status code 1.
// This function is meant to be used by initialization code to prevent the application from even
// starting up when something is obviously wrong.
// In particular this function should probably not be used when serving requests.
func Fatal(msg string, ctx ...interface{}) {
	log.Crit(msg, ctx...)
	os.Exit(1)
}

// ToLogCtx converts the given map into a log context.
func ToLogCtx(m map[string]string) log.Ctx {
	res := make(log.Ctx, len(m))
	for k, v := range m {
		res[k] = interface{}(v)
	}
	return res
}

// ToLogCtxA converts the given map into a log context.
func ToLogCtxA(m map[string][]string) log.Ctx {
	res := make(log.Ctx, len(m))
	for k, v := range m {
		res[k] = interface{}(v)
	}
	return res
}
