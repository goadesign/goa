package goa

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa/design"
)

// Controllers implement a resource actions.
// Use Application.NewController to create controller objects.
type Controller struct {
	Application  *Application               // Parent application
	Resource     *design.ResourceDefinition // Corresponding resource definition
	Actions      map[string]*Action         // Registered actions indexed by name
	ErrorHandler ErrorHandler               // Controller specific error handler
	router       *httprouter.Router         // Controller router
}

// Action provides the implementation and definition of an API endpoint.
// The implementation is provided via a handler function whose signature is given by the
// definition.
type Action struct {
	Handler    interface{}              // Endpoint implementation: user action handler
	Definition *design.ActionDefinition // Endpoint definition
}

// SetActionHandler registers a handler for the action with given name.
// The handler is a function whose signature depends on the action
func (c *Controller) SetActionHandler(name string, handler interface{}) {
	if _, ok := c.Actions[name]; ok {
		fatalf("multiple handlers for %s of %s.", name, c.Resource.Name)
	}
	action, ok := c.Resource.Actions[name]
	if !ok {
		fatalf("%s does not have an action with name '%s'", c.Resource.Name, name)
	}
	c.Actions[name] = &Action{Handler: handler, Definition: action}
	h, ok := handlers[handlerId(c.Resource, action)]
	if !ok {
		fatalf("handler for action %s of %s not defined, you may need to run 'goa'.")
	}
	// We're set, now hook up the handler with httprouter.
	handle := c.actionHandle(action, h.HandlerF)
	for _, r := range action.Routes {
		c.router.Handle(r.Verb, r.Path, handle)
	}
}

// SetErrorHandler defines an application wide error handler.
// The default error handler returns a 500 status code with the error message in the response body.
// Controllers may override the application error handler.
func (c *Controller) SetErrorHandler(handler ErrorHandler) {
	c.ErrorHandler = handler
}

// actionHandle returns a httprouter handle for the given action definition.
func (c *Controller) actionHandle(a *design.ActionDefinition, h HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		//pathParans := load(params)
		context := Context{
			Action: a,
			// Initialize context with params
		}
		err := h(&context)
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
