package goa

import "fmt"

// A Handler associates a wrapper with a resource action name
type Handler struct {
	ResourceName string
	ActionName   string
	Wrapper      *ActionWrapper
}

// An ActionWrapper implements an API endpoint by calling the corresponding user
// defined action implementation with the action specific context.
type ActionWrapper func(c *Context) *Response

// Registered handlers
var handlers []*Handler

// RegisterHandlers initializes the list of action handlers
func RegisterHandlers(h ...*Handler) {
	handlers = h
}

// mountWrappers is called when the application starts to register all the API
// endpoints and their corresponding action wrappers.
func mountWrappers() error {
	for _, handler := range handlers {
		r, ok := Resources[handler.ResourceName]
		if !ok {
			return fmt.Errorf("Handler for unknown resource '%s'",
				handler.ResourceName)
		}
		a, ok := r.Actions[handler.ActionName]
		if !ok {
			return fmt.Errorf("Handler for unknown %s action '%s'",
				handler.ResourceName, handler.ActionName)
		}
		SetWrappers(a.Routes, handler.wrapper)
	}
	return nil
}
