package design

import (
	"fmt"
	"net/url"

	"goa.design/goa/eval"
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
		// Scopes lists the JWT or OAuth2 scopes.
		Scopes []*ScopeExpr
		// Flows determine the oauth2 flows supported by this scheme.
		Flows []*FlowExpr
		// Metadata is a list of key/value pairs
		Metadata MetadataExpr
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

	// A ScopeExpr defines a scope name and description.
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

// DupScheme creates a copy of the given scheme expression.
func DupScheme(sch *SchemeExpr) *SchemeExpr {
	dup := SchemeExpr{
		Kind:        sch.Kind,
		SchemeName:  sch.SchemeName,
		Description: sch.Description,
		In:          sch.In,
		Scopes:      sch.Scopes,
		Flows:       sch.Flows,
		Metadata:    sch.Metadata,
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

// Finalize makes the TokenURL and AuthorizationURL complete if needed.
func (s *SchemeExpr) Finalize() {
	for _, f := range s.Flows {
		f.Finalize()
	}
}

// EvalName returns the name of the expression used in error messages.
func (s *FlowExpr) EvalName() string {
	if s.TokenURL != "" {
		return fmt.Sprintf("flow with token URL %q", s.TokenURL)
	}
	return fmt.Sprintf("flow with refresh URL %q", s.RefreshURL)
}

// Validate ensures that TokenURL and AuthorizationURL are valid URLs.
func (s *FlowExpr) Validate() *eval.ValidationErrors {
	verr := new(eval.ValidationErrors)
	_, err := url.Parse(s.TokenURL)
	if err != nil {
		verr.Add(s, "invalid token URL %q: %s", s.TokenURL, err)
	}
	_, err = url.Parse(s.AuthorizationURL)
	if err != nil {
		verr.Add(s, "invalid authorization URL %q: %s", s.AuthorizationURL, err)
	}
	_, err = url.Parse(s.RefreshURL)
	if err != nil {
		verr.Add(s, "invalid refresh URL %q: %s", s.RefreshURL, err)
	}
	return verr
}

// Finalize makes the TokenURL and AuthorizationURL complete if needed.
func (s *FlowExpr) Finalize() {
	tu, _ := url.Parse(s.TokenURL)         // validated in Validate
	au, _ := url.Parse(s.AuthorizationURL) // validated in Validate
	ru, _ := url.Parse(s.RefreshURL)       // validated in Validate
	tokenOK := s.TokenURL == "" || tu.IsAbs()
	authOK := s.AuthorizationURL == "" || au.IsAbs()
	refreshOK := s.RefreshURL == "" || ru.IsAbs()
	if tokenOK && authOK && refreshOK {
		return
	}
	var (
		scheme string
		host   string
	)
	if len(Root.API.Servers) > 0 {
		u, _ := url.Parse(Root.API.Servers[0].URL)
		scheme = u.Scheme
		host = u.Host
	}
	if !tokenOK {
		tu.Scheme = scheme
		tu.Host = host
		s.TokenURL = tu.String()
	}
	if !authOK {
		au.Scheme = scheme
		au.Host = host
		s.AuthorizationURL = au.String()
	}
	if !refreshOK {
		ru.Scheme = scheme
		ru.Host = host
		s.RefreshURL = ru.String()
	}
}

// Type returns the grant type of the OAuth2 grant.
func (s *FlowExpr) Type() string {
	switch s.Kind {
	case AuthorizationCodeFlowKind:
		return "authorization_code"
	case ImplicitFlowKind:
		return "implicit"
	case PasswordFlowKind:
		return "password"
	case ClientCredentialsFlowKind:
		return "client_credentials"
	default:
		panic(fmt.Sprintf("unknown flow kind: %#v", s.Kind)) // bug
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
