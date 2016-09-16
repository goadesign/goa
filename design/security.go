package design

import (
	"fmt"
	"net/url"

	"github.com/goadesign/goa/dslengine"
)

// SecuritySchemeKind is a type of security scheme, according to the
// swagger specs.
type SecuritySchemeKind int

const (
	// OAuth2SecurityKind means "oauth2" security type.
	OAuth2SecurityKind SecuritySchemeKind = iota + 1
	// BasicAuthSecurityKind means "basic" security type.
	BasicAuthSecurityKind
	// APIKeySecurityKind means "apiKey" security type.
	APIKeySecurityKind
	// JWTSecurityKind means an "apiKey" security type, with support for TokenPath and Scopes.
	JWTSecurityKind
	// NoSecurityKind means to have no security for this endpoint.
	NoSecurityKind
)

// SecurityDefinition defines security requirements for an Action
type SecurityDefinition struct {
	// Scheme defines the Security Scheme used for this action.
	Scheme *SecuritySchemeDefinition

	// Scopes are scopes required for this action
	Scopes []string `json:"scopes,omitempty"`
}

// Context returns the generic definition name used in error messages.
func (s *SecurityDefinition) Context() string { return "Security" }

// SecuritySchemeDefinition defines a security scheme used to
// authenticate against the API being designed. See
// http://swagger.io/specification/#securityDefinitionsObject for more
// information.
type SecuritySchemeDefinition struct {
	// Kind is the sort of security scheme this object represents
	Kind SecuritySchemeKind
	// DSLFunc is an optional DSL function
	DSLFunc func()

	// Scheme is the name of the security scheme, referenced in
	// Security() declarations. Ex: "googAuth", "my_big_token", "jwt".
	SchemeName string `json:"scheme"`

	// Type is one of "apiKey", "oauth2" or "basic", according to the
	// Swagger specs. We also support "jwt".
	Type string `json:"type"`
	// Description describes the security scheme. Ex: "Google OAuth2"
	Description string `json:"description"`
	// In determines whether it is in the "header" or in the "query"
	// string that we will find an `apiKey`.
	In string `json:"in,omitempty"`
	// Name refers to a header or parameter name, based on In's value.
	Name string `json:"name,omitempty"`
	// Scopes is a list of available scopes for this scheme, along
	// with their textual description.
	Scopes map[string]string `json:"scopes,omitempty"`
	// Flow determines the oauth2 flow to use for this scheme.
	Flow string `json:"flow,omitempty"`
	// TokenURL holds the URL for refreshing tokens with oauth2 or JWT
	TokenURL string `json:"token_url,omitempty"`
	// AuthorizationURL holds URL for retrieving authorization codes with oauth2
	AuthorizationURL string `json:"authorization_url,omitempty"`
	// Metadata is a list of key/value pairs
	Metadata dslengine.MetadataDefinition
}

// DSL returns the DSL function
func (s *SecuritySchemeDefinition) DSL() func() {
	return s.DSLFunc
}

// Context returns the generic definition name used in error messages.
func (s *SecuritySchemeDefinition) Context() string {
	dslFunc := "[unknown]"
	switch s.Kind {
	case OAuth2SecurityKind:
		dslFunc = "OAuth2Security"
	case BasicAuthSecurityKind:
		dslFunc = "BasicAuthSecurity"
	case APIKeySecurityKind:
		dslFunc = "APIKeySecurity"
	case JWTSecurityKind:
		dslFunc = "JWTSecurity"
	}
	return dslFunc
}

// Validate ensures that TokenURL and AuthorizationURL are valid URLs.
func (s *SecuritySchemeDefinition) Validate() error {
	_, err := url.Parse(s.TokenURL)
	if err != nil {
		return fmt.Errorf("invalid token URL %#v: %s", s.TokenURL, err)
	}
	_, err = url.Parse(s.AuthorizationURL)
	if err != nil {
		return fmt.Errorf("invalid authorization URL %#v: %s", s.AuthorizationURL, err)
	}
	return nil
}

// Finalize makes the TokenURL and AuthorizationURL complete if needed.
func (s *SecuritySchemeDefinition) Finalize() {
	tu, _ := url.Parse(s.TokenURL)         // validated in Validate
	au, _ := url.Parse(s.AuthorizationURL) // validated in Validate
	tokenOK := s.TokenURL == "" || tu.IsAbs()
	authOK := s.AuthorizationURL == "" || au.IsAbs()
	if tokenOK && authOK {
		return
	}
	var scheme string
	if len(Design.Schemes) > 0 {
		scheme = Design.Schemes[0]
	}
	if !tokenOK {
		tu.Scheme = scheme
		tu.Host = Design.Host
		s.TokenURL = tu.String()
	}
	if !authOK {
		au.Scheme = scheme
		au.Host = Design.Host
		s.AuthorizationURL = au.String()
	}
}
