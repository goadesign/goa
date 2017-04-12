package goa

import "context"

// Location is the enum defining where the value of key based security schemes should be read:
// either a HTTP request header or a URL querystring value
type Location string

// LocHeader indicates the secret value should be loaded from the request headers.
const LocHeader Location = "header"

// LocQuery indicates the secret value should be loaded from the request URL querystring.
const LocQuery Location = "query"

// ContextRequiredScopes extracts the security scopes from the given context.
// This should be used in auth handlers to validate that the required scopes are present in the
// JWT or OAuth2 token.
func ContextRequiredScopes(ctx context.Context) []string {
	if s := ctx.Value(securityScopesKey); s != nil {
		return s.([]string)
	}
	return nil
}

// WithRequiredScopes builds a context containing the given required scopes.
func WithRequiredScopes(ctx context.Context, scopes []string) context.Context {
	return context.WithValue(ctx, securityScopesKey, scopes)
}

// OAuth2Security represents the `oauth2` security scheme. It is instantiated by the generated code
// accordingly to the use of the different `*Security()` DSL functions and `Security()` in the
// design.
type OAuth2Security struct {
	// Description of the security scheme
	Description string
	// Flow defines the OAuth2 flow type. See http://swagger.io/specification/#securitySchemeObject
	Flow string
	// TokenURL defines the OAuth2 tokenUrl.  See http://swagger.io/specification/#securitySchemeObject
	TokenURL string
	// AuthorizationURL defines the OAuth2 authorizationUrl.  See http://swagger.io/specification/#securitySchemeObject
	AuthorizationURL string
	// Scopes defines a list of scopes for the security scheme, along with their description.
	Scopes map[string]string
}

// BasicAuthSecurity represents the `Basic` security scheme, which consists of a simple login/pass,
// accessible through Request.BasicAuth().
type BasicAuthSecurity struct {
	// Description of the security scheme
	Description string
}

// APIKeySecurity represents the `apiKey` security scheme. It handles a key that can be in the
// headers or in the query parameters, and does authentication based on that.  The Name field
// represents the key of either the query string parameter or the header, depending on the In field.
type APIKeySecurity struct {
	// Description of the security scheme
	Description string
	// In represents where to check for some data, `query` or `header`
	In Location
	// Name is the name of the `header` or `query` parameter to check for data.
	Name string
}

// JWTSecurity represents an api key based scheme, with support for scopes and a token URL.
type JWTSecurity struct {
	// Description of the security scheme
	Description string
	// In represents where to check for the JWT, `query` or `header`
	In Location
	// Name is the name of the `header` or `query` parameter to check for data.
	Name string
	// TokenURL defines the URL where you'd get the JWT tokens.
	TokenURL string
	// Scopes defines a list of scopes for the security scheme, along with their description.
	Scopes map[string]string
}
