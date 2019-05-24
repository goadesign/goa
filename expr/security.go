package expr

import (
	"fmt"
	"net/url"

	"goa.design/goa/v3/eval"
)

// SchemeKind is a type of security scheme.
type SchemeKind int

const (
	// OAuth2Kind identifies a "OAuth2" security scheme.
	OAuth2Kind SchemeKind = iota + 1
	// BasicAuthKind means "basic" security scheme.
	BasicAuthKind
	// APIKeyKind means "apiKey" security scheme.
	APIKeyKind
	// JWTKind means an "apiKey" security scheme, with support for
	// TokenPath and Scopes.
	JWTKind
	// NoKind means to have no security for this endpoint.
	NoKind
)

// FlowKind is a type of OAuth2 flow.
type FlowKind int

const (
	// AuthorizationCodeFlowKind identifies a OAuth2 authorization code
	// flow.
	AuthorizationCodeFlowKind FlowKind = iota + 1
	// ImplicitFlowKind identifiers a OAuth2 implicit flow.
	ImplicitFlowKind
	// PasswordFlowKind identifies a Resource Owner Password flow.
	PasswordFlowKind
	// ClientCredentialsFlowKind identifies a OAuth Client Credentials flow.
	ClientCredentialsFlowKind
)

type (
	// SecurityExpr defines a security requirement.
	SecurityExpr struct {
		// Schemes is the list of security schemes used for this
		// requirement.
		Schemes []*SchemeExpr
		// Scopes list the required scopes if any.
		Scopes []string
	}

	// SchemeExpr defines a security scheme used to authenticate against the
	// method being designed.
	SchemeExpr struct {
		// Kind is the sort of security scheme this object represents.
		Kind SchemeKind
		// SchemeName is the name of the security scheme, e.g. "googAuth",
		// "my_big_token", "jwt".
		SchemeName string
		// Description describes the security scheme e.g. "Google OAuth2"
		Description string
		// In determines the location of the API key, one of "header" or
		// "query".
		In string
		// Name refers to a header or parameter name, based on In's
		// value.
		Name string
		// Scopes lists the Basic, APIKey, JWT or OAuth2 scopes.
		Scopes []*ScopeExpr
		// Flows determine the oauth2 flows supported by this scheme.
		Flows []*FlowExpr
		// Meta is a list of key/value pairs
		Meta MetaExpr
	}

	// FlowExpr describes a specific OAuth2 flow.
	FlowExpr struct {
		// Kind is the kind of flow.
		Kind FlowKind
		// AuthorizationURL to be used for implicit or authorizationCode
		// flows.
		AuthorizationURL string
		// TokenURL to be used for password, clientCredentials or
		// authorizationCode flows.
		TokenURL string
		// RefreshURL to be used for obtaining refresh token.
		RefreshURL string
	}

	// ScopeExpr defines a security scope.
	ScopeExpr struct {
		// Name of the scope.
		Name string
		// Description is the description of the scope.
		Description string
	}
)

// EvalName returns the generic definition name used in error messages.
func (s *SecurityExpr) EvalName() string {
	var suffix string
	if len(s.Schemes) > 0 && len(s.Schemes[0].SchemeName) > 0 {
		suffix = "scheme " + s.Schemes[0].SchemeName
	}
	return "Security" + suffix
}

// DupRequirement creates a copy of the given security requirement.
func DupRequirement(req *SecurityExpr) *SecurityExpr {
	dup := &SecurityExpr{
		Scopes:  req.Scopes,
		Schemes: make([]*SchemeExpr, 0, len(req.Schemes)),
	}
	for _, s := range req.Schemes {
		dup.Schemes = append(dup.Schemes, DupScheme(s))
	}
	return dup
}

// DupScheme creates a copy of the given scheme expression.
func DupScheme(sch *SchemeExpr) *SchemeExpr {
	dup := SchemeExpr{
		Kind:        sch.Kind,
		SchemeName:  sch.SchemeName,
		Description: sch.Description,
		In:          sch.In,
		Scopes:      sch.Scopes,
		Flows:       sch.Flows,
		Meta:        sch.Meta,
	}
	return &dup
}

// Type returns the type of the scheme.
func (s *SchemeExpr) Type() string {
	switch s.Kind {
	case OAuth2Kind:
		return "OAuth2"
	case BasicAuthKind:
		return "BasicAuth"
	case APIKeyKind:
		return "APIKey"
	case JWTKind:
		return "JWT"
	default:
		panic(fmt.Sprintf("unknown scheme kind: %#v", s.Kind)) // bug
	}
}

// EvalName returns the generic definition name used in error messages.
func (s *SchemeExpr) EvalName() string {
	return s.Type() + "Security"
}

// Hash returns a unique hash value for s.
func (s *SchemeExpr) Hash() string {
	return fmt.Sprintf("%s_%s_%s", s.SchemeName, s.In, s.Name)
}

// Validate ensures that the method payload contains attributes required
// by the scheme.
func (s *SchemeExpr) Validate() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	for _, f := range s.Flows {
		if err := f.Validate(); err != nil {
			verr.Merge(err)
		}
	}
	return verr
}

// EvalName returns the name of the expression used in error messages.
func (f *FlowExpr) EvalName() string {
	return "flow " + f.Type()
}

// Validate ensures that TokenURL and AuthorizationURL are valid URLs.
func (f *FlowExpr) Validate() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	if _, err := url.Parse(f.TokenURL); err != nil {
		verr.Add(f, "invalid token URL %q: %s", f.TokenURL, err)
	}
	if _, err := url.Parse(f.AuthorizationURL); err != nil {
		verr.Add(f, "invalid authorization URL %q: %s", f.AuthorizationURL, err)
	}
	if _, err := url.Parse(f.RefreshURL); err != nil {
		verr.Add(f, "invalid refresh URL %q: %s", f.RefreshURL, err)
	}
	return verr
}

// Type returns the grant type of the OAuth2 grant.
func (f *FlowExpr) Type() string {
	switch f.Kind {
	case AuthorizationCodeFlowKind:
		return "authorization_code"
	case ImplicitFlowKind:
		return "implicit"
	case PasswordFlowKind:
		return "password"
	case ClientCredentialsFlowKind:
		return "client_credentials"
	default:
		panic(fmt.Sprintf("unknown flow kind: %#v", f.Kind)) // bug
	}
}

func (k SchemeKind) String() string {
	switch k {
	case BasicAuthKind:
		return "Basic"
	case APIKeyKind:
		return "APIKey"
	case JWTKind:
		return "JWT"
	case OAuth2Kind:
		return "OAuth2"
	case NoKind:
		return "None"
	default:
		panic("unknown kind") // bug
	}
}
