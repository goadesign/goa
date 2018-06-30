package design

import (
	"fmt"
	"net/url"
	"sort"

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
		// Requirements contains the security requirements that apply to
		// all the service methods. One requirement is composed of
		// potentially multiple schemes. Incoming requests must validate
		// at least one requirement to be authorized.
		Requirements []*SecurityExpr
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
	for _, m := range s.Methods {
		if err := m.Validate(); err != nil {
			if verrs, ok := err.(*eval.ValidationErrors); ok {
				verr.Merge(verrs)
			}
		}
	}
	for _, e := range s.Errors {
		if err := e.Validate(); err != nil {
			if verrs, ok := err.(*eval.ValidationErrors); ok {
				verr.Merge(verrs)
			}
		}
	}
	return verr
}

// Finalize finalizes all the service errors and security requirements.
func (s *ServiceExpr) Finalize() {
	for _, e := range s.Errors {
		e.Finalize()
	}
	for _, r := range s.Requirements {
		for _, s := range r.Schemes {
			s.Finalize()
		}
	}
}

// Schemes returns the list of schemes used by the service.
func (s *ServiceExpr) Schemes() []string {
	schemes := make(map[string]bool)
	for _, srv := range s.Servers {
		if u, err := url.Parse(srv.URL); err == nil && u.Scheme != "" {
			schemes[u.Scheme] = true
		}
	}
	if len(schemes) == 0 {
		return Root.API.Schemes()
	}
	ss := make([]string, len(schemes))
	i := 0
	for s := range schemes {
		ss[i] = s
		i++
	}
	sort.Strings(ss)
	return ss
}

// Validate checks that the error name is found in the result metadata for
// custom error types.
func (e *ErrorExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	rt, ok := e.AttributeExpr.Type.(*ResultTypeExpr)
	if !ok {
		return verr
	}
	if o := AsObject(rt); o != nil {
		var errField string
		for _, n := range *o {
			if _, ok := n.Attribute.Metadata["struct:error:name"]; ok {
				if errField != "" {
					verr.Add(e, "metadata 'struct:error:name' already set for attribute %q of result type %q", errField, rt.Identifier)
					continue
				}
				errField = n.Name
			}
		}
		if errField == "" {
			verr.Add(e, "metadata 'struct:error:name' is missing in result type %q", rt.Identifier)
		}
	}
	return verr
}

// Finalize makes sure the error type is a user type since it has to generate a
// Go error.
// Note: this may produce a user type with an attribute that is not an object!
func (e *ErrorExpr) Finalize() {
	att := e.AttributeExpr
	if _, ok := att.Type.(UserType); !ok {
		ut := &UserTypeExpr{
			AttributeExpr: att,
			TypeName:      e.Name,
		}
		e.AttributeExpr = &AttributeExpr{Type: ut}
	}
}
