package design

import "fmt"

// Resource defines a resource
func Resource(name string, dsl func()) {
	if a, ok := apiDefinition(); ok {
		if _, ok := a.Resources[name]; ok {
			appendError(fmt.Errorf("multiple definitions for resource %s", name))
			return
		}
		resource := &ResourceDefinition{Name: name}
		if ok := executeDSL(dsl, resource); ok {
			a.Resources[name] = resource
		}
	}
}

// MediaType sets the resource media type
func MediaType(val *MediaTypeDefinition) {
	switch c := ctxStack.current().(type) {
	case *ResourceDefinition:
		c.MediaType = val
	default:
		appendError(fmt.Errorf("Only resource definitions have a MediaType field"))
	}
}

func Status(val int) {
	switch c := ctxStack.current().(type) {
	case *ResponseDefinition:
		c.Status = val
	default:
		appendError(fmt.Errorf("Only response definitions have a Status field"))
	}
}
