package endpoint

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

type (
	// endpointData contains the data necessary to render the endpoint template.
	endpointData struct {
		// Name is the endpoint interface name
		Name string
		// Description is the endpoint description.
		Description string
		// Methods list the interface methods.
		Methods []*endpointMethod
	}

	// endpointMethod describes a single endpoint method.
	endpointMethod struct {
		// Name is the method name.
		Name string
		// Description is the method description.
		Description string
		// Payload is the payload type.
		Payload *design.UserTypeExpr
		// Result is the result type.
		Result *design.UserTypeExpr
	}
)

// Writer returns the codegen.FileWriter for the given endpoint.
func Writer(api *design.APIExpr, endpoint *design.EndpointExpr) codegen.FileWriter {

}

// endpointT is the template used to write a endpoint definition.
const endpointT = `
type (
	{{ comment .Description 1 }}
	{{ .Name }} interface {{{ range .Methods }}
		{{ comment .Description 2 }}
		{{ .Name }}(context.Context, {{ gotypename .Payload }}) ({{ gotypename .Result }}, error)
	}
)
`
