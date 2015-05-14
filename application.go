package goa

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa/design"
)

// Current application
var app *Application

// Applications consist of a set of controllers.
// A controller implements a resource action.
type Application struct {
	Name        string                 // Application name
	Controllers map[string]*Controller // Controllers indexed by resource name
	router      *httprouter.Router     // Application router
}

// New instantiates a new goa application with the given name.
func New(name string) *Application {
	app = &Application{
		Name:        name,
		Controllers: make(map[string]*Controller),
	}
	return app
}

// NewController adds a controller for the resource with given name to the application.
func (a *Application) NewController(name string) *Controller {
	if _, ok := a.Controllers[name]; ok {
		fatalf("multiple controllers for %s", name)
	}
	resource, ok := design.Definition.Resources[name]
	if !ok {
		fatalf("unknown resource %s", name)
	}
	c := Controller{
		ResourceName: name,
		actions:      make(map[string]interface{}),
	}
	a.Controllers[name] = &c

	// Now register httprouter handles.
	for _, action := range resource.Actions {
		h, ok := handlers[resource.Name+"#"+action.Name]
		if !ok {
			fatalf("Handler for action %s of %s not defined, please run 'goa' if you think it should be.")
		}
		for _, r := range action.Routes {
			a.router.Handle(r.Verb, r.Path, a.actionHandle(&c, resource, action, h.Handler))
		}
	}

	return &c
}

// actionHandle returns a httprouter handle for the given action definition.
// It strives to put as much computation as possible outside of the handle and uses closures to
// transmit the state.
func (app *Application) actionHandle(c *Controller, r *design.ResourceDefinition, a *design.ActionDefinition, h handlerFunc) httprouter.Handle {
	actionName := a.Name
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		action, ok := c.actions[actionName]
		if !ok {
			respondErrorf(w, "unknown %s action %s", r.Name, actionName)
			return
		}

		//pathParans := load(params)
		context := Context{
			Action: action,
			// Initialize context with params
		}
		response := h(&context)
		validateResponse(response)
		writeResponse(w, response)
	}
}

func respondErrorf(w http.ResponseWriter, format string, val ...interface{}) {
	w.WriteHeader(500)
	w.Write([]byte(fmt.Sprintf(format, val...)))
}
