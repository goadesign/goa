package design

// SecurityMethodKind is a type of security method, according to the
// swagger specs.
type SecurityMethodKind int

const (
	// OAuth2SecurityKind means "oauth2" security type.
	OAuth2SecurityKind SecurityMethodKind = iota + 1
	// BasicAuthSecurityKind means "basic" security type.
	BasicAuthSecurityKind
	// APIKeySecurityKind means "apiKey" security type.
	APIKeySecurityKind
	// JWTSecurityKind means an "apiKey" security type, with support for TokenURL and Scopes.
	JWTSecurityKind
	// NoSecurityKind means to have no security for this endpoint.
	NoSecurityKind
)

// SecurityDefinition defines security requirements for an Action
type SecurityDefinition struct {
	// Method defines the Security Method used for this action.
	Method *SecurityMethodDefinition

	// Scopes are scopes required for this action
	Scopes []string `json:"scopes,omitempty"`
}

// Context returns the generic definition name used in error messages.
func (s *SecurityDefinition) Context() string { return "Security" }

// SecurityMethodDefinition defines a security method used to
// authenticate against the API being designed. See
// http://swagger.io/specification/#securityDefinitionsObject for more
// information.
type SecurityMethodDefinition struct {
	// Kind is the sort of security method this object represents
	Kind SecurityMethodKind
	// DSLFunc is an optional DSL function
	DSLFunc func()

	// Method is the name of the security method, referenced in
	// Security() declarations. Ex: "googAuth", "my_big_token", "jwt".
	Method string `json:"method"`

	// Type is one of "apiKey", "oauth2" or "basic", according to the
	// Swagger specs.
	Type string `json:"type"`
	// Description describes the security method. Ex: "Google OAuth2"
	Description string `json:"description"`
	// In determines whether it is in the "header" or in the "query"
	// string that we will find an `apiKey`.
	In string `json:"in,omitempty"`
	// Name refers to a header or parameter name, based on In's value.
	Name string `json:"name,omitempty"`
	// Scopes is a list of available scopes for this method, along
	// with their textual description.
	Scopes map[string]string `json:"scopes,omitempty"`
	// Flow determines the oauth2 flow to use for this method.
	Flow string `json:"flow,omitempty"`
	// TokenURL holds the tokenUrl for the oauth2 flow
	TokenURL string `json:"token_url,omitempty"`
	// AuthorizationURL holds the authorizationUrl for the oauth2 flow
	AuthorizationURL string `json:"authorization_url,omitempty"`
}

// DSL returns the DSL function
func (s *SecurityMethodDefinition) DSL() func() {
	return s.DSLFunc
}

// Context returns the generic definition name used in error messages.
func (s *SecurityMethodDefinition) Context() string {
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
