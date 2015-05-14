package goa

import "github.com/raphael/goa/design"

// Registered handlers
var handlers map[string]*handler = make(map[string]*handler)

type handlerFunc func(*Context) *Response

type handler struct {
	ResourceName string
	ActionName   string
	Handler      handlerFunc
}

func registerHandlers(hs []*handler) {
	for _, handler := range hs {
		resource, ok := design.Definition.Resources[handler.ResourceName]
		if !ok {
			fatalf("unknown resource %s", handler.ResourceName)
		}
		action, ok := resource.Actions[handler.ActionName]
		if !ok {
			fatalf("unknown %s action '%s'", handler.ResourceName, handler.ActionName)
		}
		handlers[handler.ResourceName+"#"+handler.ActionName] = handler
	}
}
