// This file defines services and their errors. It also distinguishes errors
// named in the design from errors created for one method.
package expr

import (
	"errors"
	"fmt"
	"strings"

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
		// ClientInterceptors is the list of client interceptors.
		ClientInterceptors []*InterceptorExpr
		// ServerInterceptors is the list of server interceptors.
		ServerInterceptors []*InterceptorExpr
		// Meta is a set of key/value pairs with semantic that is
		// specific to each generator.
		Meta MetaExpr
		// design points to the root containing this service.
		design *RootExpr
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
	return s.design.Error(name)
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
			var verrs *eval.ValidationErrors
			if errors.As(err, &verrs) {
				verr.Merge(verrs)
			}
		}
	}
	verr.Merge(s.validateInlineMethodErrors())
	return verr
}

// validateInlineMethodErrors rejects two inline errors that request one public
// Go error name but define different values.
func (s *ServiceExpr) validateInlineMethodErrors() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	seen := make(map[string]*ErrorExpr)
	for _, serviceError := range s.Errors {
		if standardErrorUsesGeneratedConstructor(serviceError) {
			seen[serviceError.Name] = serviceError
		}
	}
	for _, method := range s.Methods {
		for _, methodError := range method.Errors {
			if !standardErrorUsesGeneratedConstructor(methodError) {
				continue
			}
			if previous := seen[methodError.Name]; previous != nil {
				if !equivalentErrorAttributes(previous.AttributeExpr, methodError.AttributeExpr) {
					if settings := differingErrorQualifierSettings(previous.AttributeExpr, methodError.AttributeExpr); len(settings) > 0 {
						verr.Add(
							methodError,
							"error %q cannot use one generated constructor because its %s setting differs in service %q",
							methodError.Name,
							strings.Join(settings, ", "),
							s.Name,
						)
					} else {
						verr.Add(
							methodError,
							"inline error %q must define the same value contract in every method of service %q",
							methodError.Name,
							s.Name,
						)
					}
				}
				continue
			}
			seen[methodError.Name] = methodError
		}
	}
	return verr
}

// standardErrorUsesGeneratedConstructor reports whether Goa generates the
// shared Make<Name> function whose behavior repeated declarations could change.
func standardErrorUsesGeneratedConstructor(errorExpression *ErrorExpr) bool {
	userType, authored := errorExpression.Type.(UserType)
	return !authored || IsErrorResult(userType)
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
	walkAttribute(e.AttributeExpr, func(name string, att *AttributeExpr) error { // nolint: errcheck
		if _, ok := att.Meta["struct:error:name"]; ok {
			if errField != "" {
				verr.Add(e, "duplicate error names in type %q", e.Type.Name())
			}
			errField = name
			if att.Type != String {
				verr.Add(e, "error name %q must be a string in type %q", name, e.Type.Name())
			}
			if !e.IsRequired(name) {
				verr.Add(e, "error name %q must be required in type %q", name, e.Type.Name())
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
		if !IsErrorResult(dt) {
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

// finalizeMethodType wraps an inline method error and assigns the repeatable
// example key used by service and transport generators for that method error.
func (e *ErrorExpr) finalizeMethodType(method *MethodExpr) {
	e.AttributeExpr = &AttributeExpr{Type: newGeneratedUserType(
		e.Name,
		e.AttributeExpr,
		MethodErrorExampleIdentity(method, e),
		previousInlineMethodErrorOrigin(method, e.Name),
	)}
}

// previousInlineMethodErrorOrigin returns the declaration already created for
// the same inline error by an earlier method in this service.
func previousInlineMethodErrorOrigin(method *MethodExpr, name string) UserType {
	for _, previousMethod := range method.Service.Methods {
		if previousMethod == method {
			return nil
		}
		for _, previousError := range previousMethod.Errors {
			if previousError.Name != name {
				continue
			}
			userType, ok := previousError.Type.(UserType)
			if !ok {
				continue
			}
			if _, generated := GeneratedUserTypeExampleIdentity(userType); generated {
				return userType.Origin()
			}
		}
	}
	return nil
}
