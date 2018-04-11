package design

import (
	"fmt"

	"goa.design/goa/eval"
)

type (
	// MethodExpr defines a single method.
	MethodExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Name of method.
		Name string
		// Description of method for consumption by humans.
		Description string
		// Docs points to the method external documentation if any.
		Docs *DocsExpr
		// Payload attribute
		Payload *AttributeExpr
		// Result attribute
		Result *AttributeExpr
		// Errors lists the error responses.
		Errors []*ErrorExpr
		// Requirements contains the security requirements for the
		// method. One requirement is composed of potentially multiple
		// schemes. Incoming requests must validate at least one
		// requirement to be authorized.
		Requirements []*SecurityExpr
		// Service that owns method.
		Service *ServiceExpr
		// Metadata is an arbitrary set of key/value pairs, see dsl.Metadata
		Metadata MetadataExpr
	}
)

// Error returns the error with the given name. It looks up recursively in the
// endpoint then the service and finally the root expression.
func (m *MethodExpr) Error(name string) *ErrorExpr {
	for _, err := range m.Errors {
		if err.Name == name {
			return err
		}
	}
	return m.Service.Error(name)
}

// EvalName returns the generic expression name used in error messages.
func (m *MethodExpr) EvalName() string {
	var prefix, suffix string
	if m.Name != "" {
		suffix = fmt.Sprintf("method %#v", m.Name)
	} else {
		suffix = "unnamed method"
	}
	if m.Service != nil {
		prefix = m.Service.EvalName() + " "
	}
	return prefix + suffix
}

// Validate validates the method payloads, results, and errors (if any).
func (m *MethodExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if m.Payload != nil {
		verr.Merge(m.Payload.Validate("payload", m))
	}
	if m.Result != nil {
		verr.Merge(m.Result.Validate("result", m))
	}
	for _, e := range m.Errors {
		if err := e.Validate(); err != nil {
			if verrs, ok := err.(*eval.ValidationErrors); ok {
				verr.Merge(verrs)
			}
		}
	}
	for _, r := range m.Requirements {
		for _, s := range r.Schemes {
			verr.Merge(s.Validate())
			switch s.Kind {
			case BasicAuthKind:
				if !m.Payload.HasTag("security:username") {
					verr.Add(m, "payload of method %q of service %q does not define a username attribute, use Username to define one.", m.Name, m.Service.Name)
				}
				if !m.Payload.HasTag("security:password") {
					verr.Add(m, "payload of method %q of service %q does not define a password attribute, use Password to define one.", m.Name, m.Service.Name)
				}
			case APIKeyKind:
				if !m.Payload.HasTag("security:apikey:" + s.SchemeName) {
					verr.Add(m, "payload of method %q of service %q does not define an API key attribute, use APIKey to define one.", m.Name, m.Service.Name)
				}
			case JWTKind:
				if !m.Payload.HasTag("security:token") {
					verr.Add(m, "payload of method %q of service %q does not define a JWT attribute, use Token to define one.", m.Name, m.Service.Name)
				}
			case OAuth2Kind:
				if !m.Payload.HasTag("security:accesstoken") {
					verr.Add(m, "payload of method %q of service %q does not define a OAuth2 access token attribute, use AccessToken to define one.", m.Name, m.Service.Name)
				}
			}
		}
	}
	return verr
}

// Finalize makes sure the method payload and result types are set.
func (m *MethodExpr) Finalize() {
	if m.Payload == nil {
		m.Payload = &AttributeExpr{Type: Empty}
	}
	if m.Result == nil {
		m.Result = &AttributeExpr{Type: Empty}
	}
	for _, e := range m.Errors {
		e.Finalize()
	}
}
