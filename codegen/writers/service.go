package writers

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

type (
	// serviceData contains the data necessary to render the service template.
	serviceData struct {
		// Name is the service interface name
		Name string
		// Description is the service description.
		Description string
		// Methods list the interface methods.
		Methods []*serviceMethod
	}

	// serviceMethod describes a single service method.
	serviceMethod struct {
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

// Service returns the codegen.FileWriter for the given service.
func Service(api *design.APIExpr, service *design.ServiceExpr) codegen.FileWriter {
	return nil
}

// serviceT is the template used to write a service definition.
const serviceT = `
type (
	{{ comment .Description 1 }}
	{{ .Name }} interface {{{ range .Methods }}
		{{ comment .Description 2 }}
		{{ .Name }}(context.Context, {{ gotypename .Payload }}) ({{ gotypename .Result }}, error)
	}
)
`
