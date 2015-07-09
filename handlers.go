package goa

import "fmt"

// Registered handlers
var handlers map[string]*Handler

// HandlerFunc defines the generic handler function.
type HandlerFunc func(interface{}, *Context) error

// Handler associates a generated handler function with the corresponding resource and action.
// Generated handler functions all use the same generic signature so they can be called from non
// generated code. They wrap the user handler which uses a concrete context type argument that is
// specific to the action so user code does not have to use type assertions or other form of
// dynamic casting.
type Handler struct {
	ResourceName string
	ActionName   string
	HandlerF     HandlerFunc
	Verb         string
	Path         string
}

// RegisterHandlers stores the given handlers and indexes them for later lookup.
// This function is meant to be called by code generated with the goa tool.
func RegisterHandlers(hs ...*Handler) {
	handlers := make(map[string]*Handler)
	for _, handler := range hs {
		handlers[handlerId(handler.ResourceName, handler.ActionName)] = handler
	}
}

// handlerId is an internal function that returns a unique id for a resource and an action.
// The id must always be the same given the same resource and action as it is used as a lookup key.
func handlerId(resName, actionName string) string {
	return fmt.Sprintf("%s#%s", resName, actionName)
}
