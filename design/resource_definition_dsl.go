package design

import "fmt"

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
