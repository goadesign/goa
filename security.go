package goa

///////////////////////////////////////////////////////////////////

// OAuth2Security represents the `oauth2` security scheme. It is
// automatically instantiated in your generated code when you use the
// different `*Security()` DSL functions and `Security()` in your
// design.
type OAuth2Security struct {
	// Description of the security scheme
	Description string
	// Flow defines the OAuth2 flow type. See http://swagger.io/specification/#securitySchemeObject
	Flow string
	// TokenURL defines the OAuth2 tokenUrl.  See http://swagger.io/specification/#securitySchemeObject
	TokenURL string
	// AuthenticationURL defines the OAuth2 authenticationUrl.  See http://swagger.io/specification/#securitySchemeObject
	AuthenticationURL string
	// Scopes defines a list of scopes for the security scheme, along with their description.
	Scopes map[string]string
}

///////////////////////////////////////////////////////////////////

// BasicAuthSecurity represents the `Basic` security scheme, which
// consists of a simple login/pass, accessible through
// Request.BasicAuth().
type BasicAuthSecurity struct {
	// Description of the security scheme
	Description string
}

///////////////////////////////////////////////////////////////////

// APIKeySecurity represents the `apiKey` security scheme. It handles
// a key that can be in the headers or in the query parameters, and
// does authentication based on that.  The Name field represents the
// key of either the query string parameter or the header, depending
// on the In field.
type APIKeySecurity struct {
	// Description of the security scheme
	Description string
	// In represents where to check for some data, `query` or `header`
	In string
	// Name is the name of the `header` or `query` parameter to check for data.
	Name string
}

///////////////////////////////////////////////////////////////////

// JWTSecurity represents an api key based scheme, with support for
// scopes and a token URL.
type JWTSecurity struct {
	// Description of the security scheme
	Description string
	// In represents where to check for the JWT, `query` or `header`
	In string
	// Name is the name of the `header` or `query` parameter to check for data.
	Name string
	// TokenURL defines the URL where you'd get the JWT tokens.
	TokenURL string
	// Scopes defines a list of scopes for the security scheme, along with their description.
	Scopes map[string]string
}
