package design

import (
	"fmt"

	"goa.design/goa/eval"
)

type (
	// ServiceExpr describes a set of related methods.
	ServiceExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Name of method group.
		Name string
		// Description of method group for consumption by humans.
		Description string
		// Docs points to external documentation
		Docs *DocsExpr
		// Servers list the API hosts
		Servers []*ServerExpr
		// Methods is the list of service methods.
		Methods []*MethodExpr
		// Errors list the errors common to all the service methods.
		Errors []*ErrorExpr
		// Metadata is a set of key/value pairs with semantic that is
		// specific to each generator.
		Metadata MetadataExpr
	}

	// ErrorExpr defines an error response. It consists of a named
	// attribute.
	ErrorExpr struct {
		// AttributeExpr is the underlying attribute.
		*AttributeExpr
		// Name is the unique name of the error.
		Name string
	}
)

// EvalName returns the generic expression name used in error messages.
func (s *ServiceExpr) EvalName() string {
	if s.Name == "" {
		return "unnamed service"
	}
	return fmt.Sprintf("service %#v", s.Name)
}

// Error returns the error with the given name if any.
func (s *ServiceExpr) Error(name string) *ErrorExpr {
	for _, erro := range s.Errors {
		if erro.Name == name {
			return erro
		}
	}
	return Root.Error(name)
}

// Hash returns a unique hash value for s.
func (s *ServiceExpr) Hash() string {
	return "_service_+" + s.Name
}

// Validate validates the service method payloads and results.
func (s *ServiceExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	for _, m := range s.Methods {
		if m.Payload != nil {
			verr.Merge(m.Payload.Validate("payload", m))
		}
		if m.Result != nil {
			verr.Merge(m.Result.Validate("result", m))
		}
	}
	return verr
}

// Finalize finalizes all the service methods.
func (s *ServiceExpr) Finalize() {
	for _, ep := range s.Methods {
		ep.Finalize()
	}
	for _, e := range s.Errors {
		e.Finalize()
	}
}

// Finalize makes sure the error type is a user type since it has to generate a
// Go error.
func (e *ErrorExpr) Finalize() {
	if _, ok := e.AttributeExpr.Type.(UserType); !ok {
		att := e.AttributeExpr
		if !IsObject(att.Type) {
			att = &AttributeExpr{
				Type:       &Object{{"value", att}},
				Validation: &ValidationExpr{Required: []string{"value"}},
			}
		}
		ut := &UserTypeExpr{
			AttributeExpr: att,
			TypeName:      e.Name,
		}
		e.AttributeExpr = &AttributeExpr{Type: ut}
	}
}
