package goa

import (
	"fmt"
)

// Registered handlers
var handlers map[string]*HandlerFactory

// Handler defines a goa controller action handler signature.
// Handlers accept a context and return an error.
// If the error returned is not nil then the controller error handler (if defined) or application
// error handler gets invoked.
type Handler func(*Context) error

// HandlerFunc defines the generic handler factory function.
type HandlerFunc func(interface{}) (Handler, error)

// Handler associates a generated handler function with the corresponding resource and action.
// Generated handler functions all use the same generic signature so they can be called from non
// generated code. They wrap the user handler which uses a concrete context type argument that is
// specific to the action so user code does not have to use type assertions or other form of
// dynamic casting.
type HandlerFactory struct {
	ResourceName string
	ActionName   string
	Verb         string
	Path         string
	HandlerF     HandlerFunc
}

func init() {
	handlers = make(map[string]*HandlerFactory)
}

// RegisterHandlers stores the given handlers and indexes them for later lookup.
// This function is meant to be called by code generated with the goa tool.
func RegisterHandlers(hs ...*HandlerFactory) {
	for _, handler := range hs {
		handlers[handlerId(handler.ResourceName, handler.ActionName)] = handler
	}
}

// handlerId is an internal function that returns a unique id for a resource and an action.
// The id must always be the same given the same resource and action as it is used as a lookup key.
func handlerId(resName, actionName string) string {
	return fmt.Sprintf("%s#%s", resName, actionName)
}
