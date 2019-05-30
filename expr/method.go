package expr

import (
	"fmt"

	"goa.design/goa/v3/eval"
)

type (
	// StreamKind is a type denoting the kind of stream.
	StreamKind int

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
		// Meta is an arbitrary set of key/value pairs, see dsl.Meta
		Meta MetaExpr
		// Stream is the kind of stream (none, payload, result, or both)
		// the method defines.
		Stream StreamKind
		// StreamingPayload is the payload sent across the stream.
		StreamingPayload *AttributeExpr
	}
)

const (
	// NoStreamKind represents no payload or result stream in method.
	NoStreamKind StreamKind = iota + 1
	// ClientStreamKind represents client sends a streaming payload to
	// method.
	ClientStreamKind
	// ServerStreamKind represents server sends a streaming result from
	// method.
	ServerStreamKind
	// BidirectionalStreamKind represents client and server sending payload
	// and result respectively via a stream.
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

// Prepare makes sure the payload and result types are initialized (to the Empty
// type if nil).
func (m *MethodExpr) Prepare() {
	if m.Payload == nil {
		m.Payload = &AttributeExpr{Type: Empty}
	}
	if m.StreamingPayload == nil {
		m.StreamingPayload = &AttributeExpr{Type: Empty}
	}
	if m.Result == nil {
		m.Result = &AttributeExpr{Type: Empty}
	}
}

// Validate validates the method payloads, results, and errors (if any).
func (m *MethodExpr) Validate() error {
	verr := new(eval.ValidationErrors)
	if m.Payload.Type != Empty {
		verr.Merge(m.Payload.Validate("payload", m))
		// validate security scheme requirements
		var requirements []*SecurityExpr
		if len(m.Requirements) > 0 {
			requirements = m.Requirements
		} else if len(m.Service.Requirements) > 0 {
			requirements = m.Service.Requirements
		}
		for _, r := range requirements {
			for _, s := range r.Schemes {
				verr.Merge(s.Validate())
				switch s.Kind {
				case BasicAuthKind:
					if !hasTag(m.Payload, "security:username") {
						verr.Add(m, "payload of method %q of service %q does not define a username attribute, use Username to define one", m.Name, m.Service.Name)
					}
					if !hasTag(m.Payload, "security:password") {
						verr.Add(m, "payload of method %q of service %q does not define a password attribute, use Password to define one", m.Name, m.Service.Name)
					}
				case APIKeyKind:
					if !hasTag(m.Payload, "security:apikey:"+s.SchemeName) {
						verr.Add(m, "payload of method %q of service %q does not define an API key attribute, use APIKey to define one", m.Name, m.Service.Name)
					}
				case JWTKind:
					if !hasTag(m.Payload, "security:token") {
						verr.Add(m, "payload of method %q of service %q does not define a JWT attribute, use Token to define one", m.Name, m.Service.Name)
					}
				case OAuth2Kind:
					if !hasTag(m.Payload, "security:accesstoken") {
						verr.Add(m, "payload of method %q of service %q does not define a OAuth2 access token attribute, use AccessToken to define one", m.Name, m.Service.Name)
					}
				}
			}
			for _, scope := range r.Scopes {
				found := false
				for _, s := range r.Schemes {
					if s.Kind == BasicAuthKind || s.Kind == APIKeyKind || s.Kind == OAuth2Kind || s.Kind == JWTKind {
						for _, se := range s.Scopes {
							if se.Name == scope {
								found = true
								break
							}
						}
					}
				}
				if !found {
					verr.Add(m, "security scope %q not found in any of the security schemes.", scope)
				}
			}
		}
	}
	if m.StreamingPayload.Type != Empty {
		verr.Merge(m.StreamingPayload.Validate("streaming_payload", m))
	}
	if m.Result.Type != Empty {
		verr.Merge(m.Result.Validate("result", m))
	}
	for i, e := range m.Errors {
		if err := e.Validate(); err != nil {
			if verrs, ok := err.(*eval.ValidationErrors); ok {
				verr.Merge(verrs)
			}
		}
		for j, e2 := range m.Errors {
			// If an object type is used to define more than one errors validate the
			// presence of struct:error:name meta in the object type.
			if i != j && e.Type == e2.Type && IsObject(e.Type) {
				var found bool
				walkAttribute(e.AttributeExpr, func(name string, att *AttributeExpr) error {
					if _, ok := att.Meta["struct:error:name"]; ok {
						found = true
						return fmt.Errorf("struct:error:name found: stop iteration")
					}
					return nil
				})
				if !found {
					verr.Add(e, "type %q is used to define multiple errors and must identify the attribute containing error name. Use Meta with the key 'struct:error:name' on the error name attribute", e.AttributeExpr.Type.Name())
					break
				}
			}
		}
	}
	return verr
}

// hasTag is a helper function that traverses the given attribute and all its
// bases recursively looking for an attribute with the given tag meta. This
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
	if m.StreamingPayload == nil {
		m.StreamingPayload = &AttributeExpr{Type: Empty}
	} else {
		m.StreamingPayload.Finalize()
	}
	if m.Result == nil {
		m.Result = &AttributeExpr{Type: Empty}
	} else {
		m.Result.Finalize()
		if rt, ok := m.Result.Type.(*ResultTypeExpr); ok {
			rt.Finalize()
		}
	}
	for _, e := range m.Errors {
		e.Finalize()
	}

	// Inherit security requirements
	noreq := false
	for _, r := range m.Requirements {
		// Handle special case of no security
		for _, s := range r.Schemes {
			if s.Kind == NoKind {
				noreq = true
				break
			}
		}
		if noreq {
			break
		}
	}
	if noreq {
		m.Requirements = nil
	} else if len(m.Requirements) == 0 && len(m.Service.Requirements) > 0 {
		m.Requirements = copyReqs(m.Service.Requirements)
	}

}

// IsStreaming determines whether the method streams payload or result.
func (m *MethodExpr) IsStreaming() bool {
	return m.Stream != 0 && m.Stream != NoStreamKind
}

// IsPayloadStreaming determines whether the method streams payload.
func (m *MethodExpr) IsPayloadStreaming() bool {
	return m.Stream == ClientStreamKind || m.Stream == BidirectionalStreamKind
}

// helper function that duplicates just enough of a security expression so that
// its scheme names can be overridden without affecting the original.
func copyReqs(reqs []*SecurityExpr) []*SecurityExpr {
	reqs2 := make([]*SecurityExpr, len(reqs))
	for i, req := range reqs {
		req2 := &SecurityExpr{Scopes: req.Scopes}
		schs := make([]*SchemeExpr, len(req.Schemes))
		for j, sch := range req.Schemes {
			schs[j] = &SchemeExpr{
				Kind:        sch.Kind,
				SchemeName:  sch.SchemeName,
				Description: sch.Description,
				In:          sch.In,
				Name:        sch.Name,
				Scopes:      sch.Scopes,
				Flows:       sch.Flows,
				Meta:        sch.Meta,
			}
		}
		req2.Schemes = schs
		reqs2[i] = req2
	}
	return reqs2
}
