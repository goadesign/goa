package design

import "fmt"

var (
	ctxStack contextStack // Global DSL evaluation stack
	dslError error        // Last evaluation error if any
)

// DSL evaluation contexts stack
type contextStack []interface{}

// Current evaluation context, i.e. object being currently built by DSL
func (s contextStack) Current() interface{} {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// Run DSL in given evaluation context
func executeDSL(dsl func(), ctx interface{}) error {
	ctxStack = append(ctxStack, ctx)
	dsl()
	ctxStack = ctxStack[:len(ctxStack)-1]
	return dslError
}

// Action defines an action definition DSL
func Action(name string, dsl func()) {
	action := &ActionDefinition{Name: name}
	err := executeDSL(dsl, action)
	if err != nil {
		return
	}
	switch c := ctxStack.Current().(type) {
	case *Resource:
		c.Actions = append(c.Actions, action)
	default:
		dslError = fmt.Errorf("Only resources have a Action field")
	}
}

// Define API base params
func BaseParams(attributes ...*Attribute) {
	switch c := ctxStack.Current().(type) {
	case *APIDefinition:
		c.BaseParams = attributes
	default:
		dslError = fmt.Errorf("Only API definitions have a BaseParams field")
	}
}

// Define API base path
func BasePath(val string) {
	switch c := ctxStack.Current().(type) {
	case *APIDefinition:
		c.BasePath = val
	default:
		dslError = fmt.Errorf("Only API definitions have a BasePath field")
	}
}

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
		dslError = fmt.Errorf("Only API definitions, response templates, attributes, media types and actions have a Description field")
	}
}

func Headers(val ...Header) {
	switch c := ctxStack.Current().(type) {
	case *ActionDefinition:
		c.Headers = val
	case *Response:
		c.Headers = val
	default:
		dslError = fmt.Errorf("Only Action and response definitions have a Header field")
	}
}

func MediaType(val *MediaType) {
	switch c := ctxStack.Current().(type) {
	case *Resource:
		c.MediaType = val
	default:
		dslError = fmt.Errorf("Only resource definitions have a MediaType field")
	}
}

func ResponseTemplate(name string, dsl func()) {
	template := &ResponseTemplate{Name: name}
	err := executeDSL(dsl, template)
	if err != nil {
		return err
	}
	switch c := ctxStack.Current().(type) {
	case *APIDefinition:
		c.ResponseTemplates = append(c.ResponseTemplates, template)
	default:
		dslError = fmt.Errorf("Only API definitions have a ResourceTemplate field")
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

func Title(val string) {
	switch c := ctxStack.Current().(type) {
	case *APIDefinition:
		c.Title = val
	default:
		dslError = fmt.Errorf("Only API definitions have a Title field")
	}
}

func Trait(name string, val func()) {
	trait := &TraitDefinition{Name: name, Definition: val}
	switch c := ctxStack.Current().(type) {
	case *APIDefinition:
		c.Traits = append(c.Traits, trait)
	default:
		dslError = fmt.Errorf("Only API definitions have a Trait field")
	}

}
