package design

import "github.com/goadesign/goa/eval"

type (
	// ServiceExpr describes a set of related endpoints.
	ServiceExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Name of endpoint group.
		Name string
		// Description of endpoint group for consumption by humans.
		Description string
		// Docs points to external documentation
		Docs *DocsExpr
		// Endpoints is the list of service endpoints.
		Endpoints []*EndpointExpr
		// DefaultTypeName is the name of the service default response
		// type.  The default type attributes also define the default
		// properties for request attributes with identical names.
		DefaultTypeName string
		// Errors list the errors common to all the service endpoints.
		Errors []*ErrorExpr
		// Metadata is a set of key/value pairs with semantic that is
		// specific to each generator.
		Metadata MetadataExpr
	}

	// ErrorExpr defines an error response. It consists of a named
	// attribute.
	ErrorExpr struct {
		*AttributeExpr
		Name string
	}
)

// EvalName returns the generic expression name used in error messages.
func (s *ServiceExpr) EvalName() string {
	if s.Name == "" {
		return "unnamed service"
	}
	return "service " + s.Name
}

// DefaultType returns the service default type or nil if there isn't one.
func (s *ServiceExpr) DefaultType() UserType {
	return Root.UserType(s.DefaultTypeName)
}

// Error returns the error with the given name if any.
func (s *ServiceExpr) Error(name string) *ErrorExpr {
	for _, erro := range s.Errors {
		if erro.Name == name {
			return erro
		}
	}
	return nil
}

// Finalize makes sure all endpoints that must use the service default type do.
func (s *ServiceExpr) Finalize() {
	for _, ep := range s.Endpoints {
		if ep.Request == nil {
			ep.Request = s.DefaultType()
		}
	}
}
