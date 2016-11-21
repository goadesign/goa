package design

import "goa.design/goa.v2/eval"

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

// Validate makes sure all endpoints have a request type defined.
func (s *ServiceExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	for _, ep := range s.Endpoints {
		if ep.Request == nil {
			verr.Add(ep, "request type is not defined")
		}
		if s.DefaultType() == nil && ep.Response == nil {
			verr.Add(ep, "response type is not defined and service does not define a default type")
		}
	}

	return verr
}

// Finalize sets the endpoint response types with the service default type if
// the response type isn't set. It also merges attributes from the service
// default type with the request attributes of the same name.
func (s *ServiceExpr) Finalize() {
	def := s.DefaultType()
	for _, ep := range s.Endpoints {
		if ep.Response == nil {
			ep.Response = def
		}
		if def == nil {
			continue
		}
		ep.Request.Attribute().Inherit(def.Attribute())
	}
}
