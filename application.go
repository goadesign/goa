package goa

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	log "gopkg.in/inconshreveable/log15.v2"
)

type (
	// Application consists of a set of controllers.
	// A controller implements a resource action.
	Application struct {
		log.Logger                      // Application logger
		Name         string             // Application name
		ErrorHandler ErrorHandler       // Application global error handler
		Middleware   []Middleware       // Middleware chain
		Router       *httprouter.Router // Application router
	}

	// ErrorHandler handles errors returned by action handlers and middleware.
	ErrorHandler func(Context, error)
)

var (
	// Log is the global logger. Configure it by setting its handler prior to calling New.
	// See https://godoc.org/github.com/inconshreveable/log15
	Log log.Logger
)

// Log to STDOUT by default
func init() {
	Log = log.New()
	Log.SetHandler(log.StdoutHandler)
}

// New instantiates a new goa application with the given name.
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
func (a *Application) Use(middleware interface{}) {
	m, err := NewMiddleware(middleware)
	if err != nil {
		Fatalf("invalid middlware %#v", middleware)
	}
	a.Middleware = append(a.Middleware, m)
}

// Run starts the application loop and sets up a listener on the given host/port
func (a *Application) Run(addr string) {
	a.Info("listen", "addr", addr)
	http.ListenAndServe(addr, a.Router)
}

// SetErrorHandler defines an application wide error handler.
// The default error handler returns a 500 status code with the error message in the response body.
// Controllers may override the application error handler.
func (a *Application) SetErrorHandler(handler ErrorHandler) {
	a.ErrorHandler = handler
}

// DefaultErrorHandler returns a 400 response with the error message as body.
func DefaultErrorHandler(c Context, e error) {
	if err := c.Respond(500, []byte(e.Error())); err != nil {
		Log.Error("failed to send default error handler response", "error", err)
	}
}

// Fatalf displays an error message and exits the process with status code 1.
// This function is meant to be used by initialization code to prevent the application from even
// starting up when something is obviously wrong.
// In particular this function must not be used when serving requests.
func Fatalf(format string, val ...interface{}) {
	fmt.Fprintf(os.Stderr, format, val...)
	os.Exit(1)
}
