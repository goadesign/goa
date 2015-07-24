package goa

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	log "gopkg.in/inconshreveable/log15.v2"
)

type (
	// Applications consist of a set of controllers.
	// A controller implements a resource action.
	Application struct {
		log.Logger                          // Application logger
		Name         string                 // Application name
		Controllers  map[string]*Controller // Controllers indexed by resource name
		ErrorHandler ErrorHandler           // Application global error handler
		router       *httprouter.Router     // Application router
	}

	// ErrorHandlers handle errors returned by action handlers and middleware.
	ErrorHandler func(*Context, error)
)

var (
	Log log.Logger
)

// Log nothing by default
func init() {
	Log = log.New()
	Log.SetHandler(log.DiscardHandler())
}

// New instantiates a new goa application with the given name.
func New(name string) *Application {
	return &Application{
		Logger:       Log.New("app", name),
		Name:         name,
		Controllers:  make(map[string]*Controller),
		ErrorHandler: DefaultErrorHandler,
		router:       httprouter.New(),
	}
}

// Mount adds the given controller to the application.
// It panics if a controller for a resource with the same name was already added.
func (a *Application) Mount(c *Controller) {
	c.Logger = a.Logger.New("ctl", c.Resource)
	c.Info("mouting")
	if c.Handlers == nil {
		Fatalf("controller has no handlers, use SetHandlers to register them")
	}
	for k, u := range c.Handlers {
		id := handlerId(c.Resource, k)
		h, ok := handlers[id]
		if !ok {
			Fatalf("unknown %s action %s", c.Resource, k)
		}
		handler, err := h.HandlerF(u)
		if err != nil {
			Fatalf(err.Error())
		}
		a.router.Handle(h.Verb, h.Path, c.actionHandle(handler))
		c.Info("handler", "action", k, h.Verb, h.Path)
	}
	c.Info("mounted")
	c.Application = a
}

// Run starts the application loop and sets up a listener on the given host/port
func (a *Application) Run(addr string) {
	a.Info("listen", "addr", addr)
	http.ListenAndServe(addr, a.router)
}

// SetErrorHandler defines an application wide error handler.
// The default error handler returns a 500 status code with the error message in the response body.
// Controllers may override the application error handler.
func (a *Application) SetErrorHandler(handler ErrorHandler) {
	a.ErrorHandler = handler
}

// DefaultErrorHandler returns a 400 response with the error message as body.
func DefaultErrorHandler(c *Context, e error) {
	if err := c.Respond(400, []byte(e.Error())); err != nil {
		Log.Error("failed to send default error handler response", "error", err)
		c.W.WriteHeader(500)
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
