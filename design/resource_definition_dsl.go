package design

import "fmt"

// MediaType sets the resource media type
func MediaType(val *MediaTypeDefinition) {
	switch c := ctxStack.Current().(type) {
	case *Resource:
		c.MediaType = val
	default:
		dslError = fmt.Errorf("Only resource definitions have a MediaType field")
	}
}

func Status(val int) {
	switch c := ctxStack.Current().(type) {
	case *Response:
		c.Status = val
	default:
		dslError = fmt.Errorf("Only response definitions have a Status field")
	}
}
