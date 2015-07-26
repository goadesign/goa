package dsl

import (
	"fmt"

	"github.com/raphael/goa/design"
)

// Resource defines a resource
func Resource(name string, dsl func()) {
	if a, ok := apiDefinition(); ok {
		if _, ok := a.Resources[name]; ok {
			appendError(fmt.Errorf("multiple definitions for resource %s", name))
			return nil
		}
		resource := &ResourceDefinition{Name: name}
		if ok := executeDSL(dsl, resource); ok {
			a.Resources[name] = resource
		}
	}
}

// MediaType sets the resource media type
func MediaType(val interface{}) {
	if r, ok := ctxStack.current().(*design.ResourceDefinition); ok {
		r.MediaType = val
	} else if r, ok := responseDefinition(); ok {
		r.MediaType = val
	}
}

// Prefix sets the resource path prefix
func Prefix(p string) {
	if r, ok := resourceDefinition(); ok {
		r.Prefix = p
	}
}

// Prefix sets the resource path prefix
func CanonicalAction(a string) {
	if r, ok := resourceDefinition(); ok {
		r.CanonicalAction = a
	}
}
