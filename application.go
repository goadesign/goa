package goa

import (
	"fmt"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa/design"
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

// NewController adds a controller for the resource with given name to the application.
func (a *Application) NewController(name string) *Controller {
	def := design.Definition
	if def == nil {
		fatalf("no API metadata, use design.Api to create it")
	}
	res, ok := def.Resources[name]
	if !ok {
		fatalf("unknown resource \"%s\"", name)
	}
	if _, ok := a.Controllers[name]; ok {
		fatalf("multiple controllers for %s", name)
	}
	c := &Controller{
		Application: a,
		Resource:    res,
		Actions:     make(map[string]*Action),
		router:      a.router,
	}
	a.Controllers[name] = c

	return c
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
