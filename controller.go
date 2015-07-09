package goa

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type (
	// Controllers implement the actions of a single resource.
	// Use NewController to create controller objects.
	Controller struct {
		Application  *Application // Parent application
		Handlers     Handlers     // Registered action handlers indexed by name
		ErrorHandler ErrorHandler // Controller specific error handler
		Resource     string       // Name of resource controller implements
	}

	// UserHandlers associates action names with their handler.
	Handlers map[string]UserHandler

	// UserHandlers are functions that contain the implementation for controller actions.
	// The function signatures match the corresponding design action definition.
	UserHandler interface{}
)

// NewController instantiates a new goa controller for the resource with given name.
func NewController(resource string) *Controller {
	return &Controller{Resource: resource}
}

// SetHandlers sets the controller action handlers.
func (c *Controller) SetHandlers(handlers Handlers) {
	c.Handlers = handlers
}

// SetErrorHandler defines an application wide error handler.
// The default error handler returns a 500 status code with the error message in the response body.
// Controllers may override the application error handler.
func (c *Controller) SetErrorHandler(handler ErrorHandler) {
	c.ErrorHandler = handler
}

// actionHandle returns a httprouter handle for the given action definition.
// The generated handler builds an action specific context and calls the user handler.
func (c *Controller) actionHandle(generated HandlerFunc, user interface{}) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		//pathParans := load(params)
		context := Context{
		// Initialize context with params
		}
		err := generated(user, &context)
		if err != nil {
			c.handleError(&context, err)
		}
	}
}

// handleError looks up the error handler (first in the controller then in the application) and
// invokes it.
func (c *Controller) handleError(ctx *Context, actionErr error) {
	handler := c.ErrorHandler
	if handler == nil {
		handler = c.Application.ErrorHandler
	}
	handler(ctx, actionErr)
}
