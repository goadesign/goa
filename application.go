package goa

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

// Applications consist of a set of controllers.
// A controller implements a resource action.
type Application struct {
	Name         string                 // Application name
	Controllers  map[string]*Controller // Controllers indexed by resource name
	ErrorHandler ErrorHandler           // Application global error handler
	router       *httprouter.Router     // Application router
}

// ErrorHandlers handle errors returned by action handlers and middleware.
type ErrorHandler func(*Context, error)

// New instantiates a new goa application with the given name.
func New(name string) *Application {
	return &Application{
		Name:        name,
		Controllers: make(map[string]*Controller),
		router:      httprouter.New(),
	}
}

// Mount adds the given controller to the application.
// It panics if a controller for a resource with the same name was already added.
func (a *Application) Mount(c *Controller) {
	if c.Handlers == nil {
		Fatalf("controller has no handlers, use SetHandlers to register them")
	}
	for k, u := range c.Handlers {
		id := handlerId(c.Resource, k)
		h, ok := handlers[id]
		if !ok {
			Fatalf("unknown %s action %s", c.Resource, k)
		}
		a.router.Handle(h.Verb, h.Path, c.actionHandle(h.HandlerF, u))
	}
	c.Application = a
}

// Run starts the application loop and sets up a listener on the given host/port
func (a *Application) Run(addr string) {
	http.ListenAndServe(addr, a.router)
}

// SetErrorHandler defines an application wide error handler.
// The default error handler returns a 500 status code with the error message in the response body.
// Controllers may override the application error handler.
func (a *Application) SetErrorHandler(handler ErrorHandler) {
	a.ErrorHandler = handler
}

// Fatalf displays an error message and exits the process with status code 1.
// This function is meant to be used by initialization code to prevent the application from even
// starting up when something is obviously wrong.
// In particular this function must not be used when serving requests.
func Fatalf(format string, val ...interface{}) {
	fmt.Fprintf(os.Stderr, format, val...)
	os.Exit(1)
}
