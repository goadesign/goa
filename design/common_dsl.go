package design

// Description sets the description field of the calling context.
func Description(val string) {
	switch c := ctxStack.Current().(type) {
	case *APIDefinition:
		c.Description = val
	case *ResponseTemplate:
		c.Description = val
	case *Attribute:
		c.Description = val
	case *MediaType:
		c.Description = val
	case *Action:
		c.Description = val
	default:
		func() { incompatibleDsl() }()
	}
}

func Header(name string, dsl func()) {
	h := HeaderDefinition{Name: name}


// Headers sets the headers field of the calling context.
func Headers(val ...HeaderDefinition) {
	switch c := ctxStack.Current().(type) {
	case *ActionDefinition:
		c.Headers = append(a.Headers, val)
	case *ResponseTemplate:
		c.Headers = append(r.Headers, val)
	default:
		func() { incompatibleDsl() }()
	}
}
