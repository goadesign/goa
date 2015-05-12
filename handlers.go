package goa

import (
	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa/design"
)

var (
	router = httprouter.New()
)

type handlerFunc func(*Context) *Response

type handler struct {
	ResourceName string
	ActionName   string
	Handler      handlerFunc
}

func registerHandlers(handlers []*handler) {
	if design.Definition == nil {
		fatalf("missing API definition")
	}
	for _, handler := range handlers {
		resource, ok := design.Definition.Resources[handler.ResourceName]
		if !ok {
			fatalf("unknown resource %s", handler.ResourceName)
		}
		action, ok := resource.Actions[handler.ActionName]
		if !ok {
			fatalf("unknown %s action '%s'", handler.ResourceName, handler.ActionName)
		}
		for _, route := range action.Routes {
		}
	}
}
