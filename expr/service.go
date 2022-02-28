package expr

import (
	"fmt"

	"goa.design/goa/v3/eval"
)

type (
	// ServiceExpr describes a set of related methods.
	ServiceExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Name of service.
		Name string
		// Description of service used in documentation.
		Description string
		// Docs points to external documentation
		Docs *DocsExpr
		// Methods is the list of service methods.
		Methods []*MethodExpr
		// Errors list the errors common to all the service methods.
		Errors []*ErrorExpr
		// Requirements contains the security requirements that apply to
		// all the service methods. One requirement is composed of
		// potentially multiple schemes. Incoming requests must validate
		// at least one requirement to be authorized.
		Requirements []*SecurityExpr
		// Meta is a set of key/value pairs with semantic that is
		// specific to each generator.
		Meta MetaExpr
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

// Method returns the method expression with the given name, nil if there isn't
// one.
func (s *ServiceExpr) Method(n string) *MethodExpr {
	for _, m := range s.Methods {
		if m.Name == n {
			return m
		}
	}
	return nil
}

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

// Validate validates the service methods and errors.
func (s *ServiceExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	for _, e := range s.Errors {
		if err := e.Validate(); err != nil {
			if verrs, ok := err.(*eval.ValidationErrors); ok {
				verr.Merge(verrs)
			}
		}
	}
	return verr
}

// Finalize finalizes all the service methods and errors.
func (s *ServiceExpr) Finalize() {
	for _, e := range s.Errors {
		e.Finalize()
	}
}

// Validate checks that the error name is found in the result meta for
// custom error types.
func (e *ErrorExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	var errField string
	walkAttribute(e.AttributeExpr, func(name string, att *AttributeExpr) error {
		if _, ok := att.Meta["struct:error:name"]; ok {
			if errField != "" {
				verr.Add(e, "duplicate error names in type %q", e.AttributeExpr.Type.Name())
			}
			errField = name
			if att.Type != String {
				verr.Add(e, "error name %q must be a string in type %q", name, e.AttributeExpr.Type.Name())
			}
			if !e.AttributeExpr.IsRequired(name) {
				verr.Add(e, "error name %q must be required in type %q", name, e.AttributeExpr.Type.Name())
			}
		}
		return nil
	})
	if len(verr.Errors) > 0 {
		return verr
	}
	return nil
}

// Finalize makes sure the error type is a user type since it has to generate a
// Go error.
// Note: this may produce a user type with an attribute that is not an object!
func (e *ErrorExpr) Finalize() {
	att := e.AttributeExpr
	switch dt := att.Type.(type) {
	case UserType:
		if dt != ErrorResult {
			// If this type contains an attribute with "struct:error:name" meta
			// then no need to do anything.
			if IsObject(dt) {
				for _, nat := range *AsObject(dt) {
					if _, ok := nat.Attribute.Meta["struct:error:name"]; ok {
						return
					}
				}
			}

			// This type does not have an attribute with "struct:error:name" meta.
			// It means the type is used by at most one error (otherwise validations
			// would have failed).
			dt.Attribute().AddMeta("struct:error:name", e.Name)
		}
	default:
		ut := &UserTypeExpr{
			AttributeExpr: att,
			TypeName:      e.Name,
		}
		e.AttributeExpr = &AttributeExpr{Type: ut}
	}
}
