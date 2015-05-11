package design

// Description sets the description field of the calling context.
func Description(val string) {
	switch c := ctxStack.current().(type) {
	case *APIDefinition:
		c.Description = val
	case *ResponseTemplateDefinition:
		c.Description = val
	case *AttributeDefinition:
		c.Description = val
	case *MediaTypeDefinition:
		c.Description = val
	case *ActionDefinition:
		c.Description = val
	default:
		incompatibleDsl(caller())
	}
}

func Header(name string, dsl func()) {
	h := HeaderDefinition{Name: name}
	executeDSL(dsl, &h)
}

// Headers sets the headers field of the calling context.
func Headers(val ...*HeaderDefinition) {
	switch c := ctxStack.current().(type) {
	case *ActionDefinition:
		c.Headers = append(c.Headers, val...)
	case *ResponseTemplateDefinition:
		c.Headers = append(c.Headers, val...)
	default:
		incompatibleDsl(caller())
	}
}
