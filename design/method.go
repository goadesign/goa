package design

import (
	"fmt"

	"goa.design/goa/eval"
)

type (
	// streamKind is a type denoting the kind of stream.
	streamKind int

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
		// Stream is the kind of stream (none, payload, result, or both) the method
		// defines.
		Stream streamKind
	}
)

const (
	// NoStreamKind represents no payload or result stream in method.
	NoStreamKind streamKind = iota
	// ClientStreamKind represents client sends a streaming payload to method.
	ClientStreamKind
	// ServerStreamKind represents server sends a streaming result from method.
	ServerStreamKind
	// BidirectionalStreamKind represents both client and server streams payload
	// and result respectively.
	BidirectionalStreamKind
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
		// validate security scheme requirements
		for _, r := range m.Requirements {
			for _, s := range r.Schemes {
				verr.Merge(s.Validate())
				switch s.Kind {
				case BasicAuthKind:
					if !hasTag(m.Payload, "security:username") {
						verr.Add(m, "payload of method %q of service %q does not define a username attribute, use Username to define one.", m.Name, m.Service.Name)
					}
					if !hasTag(m.Payload, "security:password") {
						verr.Add(m, "payload of method %q of service %q does not define a password attribute, use Password to define one.", m.Name, m.Service.Name)
					}
				case APIKeyKind:
					if !hasTag(m.Payload, "security:apikey:"+s.SchemeName) {
						verr.Add(m, "payload of method %q of service %q does not define an API key attribute, use APIKey to define one.", m.Name, m.Service.Name)
					}
				case JWTKind:
					if !hasTag(m.Payload, "security:token") {
						verr.Add(m, "payload of method %q of service %q does not define a JWT attribute, use Token to define one.", m.Name, m.Service.Name)
					}
				case OAuth2Kind:
					if !hasTag(m.Payload, "security:accesstoken") {
						verr.Add(m, "payload of method %q of service %q does not define a OAuth2 access token attribute, use AccessToken to define one.", m.Name, m.Service.Name)
					}
				}
			}
		}
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
	return verr
}

// hasTag is a helper function that traverses the given attribute and all its
// bases recursively looking for an attribute with the given tag metadata. This
// recursion is only needed for attributes that have not been finalized yet.
func hasTag(p *AttributeExpr, tag string) bool {
	if p.HasTag(tag) {
		return true
	}
	for _, base := range p.Bases {
		ut, ok := base.(UserType)
		if !ok {
			continue
		}
		return hasTag(ut.Attribute(), tag)
	}
	if ut, ok := p.Type.(UserType); ok {
		return hasTag(ut.Attribute(), tag)
	}
	return false
}

// Finalize makes sure the method payload and result types are set. It also
// projects the result if it is a result type and a view is explicitly set in
// the design or a result type having at most one view.
func (m *MethodExpr) Finalize() {
	if m.Payload == nil {
		m.Payload = &AttributeExpr{Type: Empty}
	} else {
		m.Payload.Finalize()
	}
	if m.Result == nil {
		m.Result = &AttributeExpr{Type: Empty}
	} else {
		m.Result.Finalize()
	}
	for _, e := range m.Errors {
		e.Finalize()
	}
}

// IsResultStreaming determines whether the method result is streamed.
func (m *MethodExpr) IsResultStreaming() bool {
	return m.Stream == ServerStreamKind || m.Stream == BidirectionalStreamKind
}
