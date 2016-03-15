package goa

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"
)

// SecurityMethod extracts the security method (OAuth2Security,
// BasicAuthSecurity, APIKeySecurity or JWTSecurity) from the
// request context.
//
// This is to be used from within a security middleware
// implementation. If you do get into your middleware code, it is
// guaranteed that the request will hold a valid SecurityMethod.  You
// can then typecast it and validate the proper credentials.
func SecurityMethod(ctx context.Context) interface{} {
	return ctx.Value(securityMethodKey)
}

// Scopes extracts from a request the scopes relevant to the action
// being executed.
//
// Call this from within your Security middleware implementation.
//
// Scopes can be empty
func Scopes(ctx context.Context) (out []string) {
	scopes, ok := ctx.Value(securityScopesKey).([]string)
	if !ok {
		return
	}
	return scopes
}

// securityMiddleware represents a security method middleware, used
// for deduplicating the different security method's methods!
type securityMiddleware struct {
	middleware Middleware
	method     interface{}

	// Description of the security method
	Description string

	// Metadata is some data passed on from the DSL.
	Metadata map[string][]string
}

// Dispatch returns a wrapped Handler, configured to handle a certain
// action's credentials validation.
//
// It is called by `app`-generated code. You shouldn't need to use
// this directly.
func (sec *securityMiddleware) Dispatch(h Handler, scopes ...string) Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
		if sec.middleware == nil {
			RequestService(ctx).Error("security method not implemented")
			return errors.New("security method not implemented")
		}

		ctx = context.WithValue(ctx, securityMethodKey, sec.method)
		if len(scopes) != 0 {
			ctx = context.WithValue(ctx, securityScopesKey, scopes)
		}

		return sec.middleware(h)(ctx, rw, req)
	}
}

///////////////////////////////////////////////////////////////////

// OAuth2Security represents the `oauth2` security method. It is
// automatically instantiated in your generated code when you use the
// different `*Security()` DSL functions and `Security()` in your
// design.
type OAuth2Security struct {
	securityMiddleware

	// Flow defines the OAuth2 flow type. See http://swagger.io/specification/#securitySchemeObject
	Flow string
	// TokenURL defines the OAuth2 tokenUrl.  See http://swagger.io/specification/#securitySchemeObject
	TokenURL string
	// AuthenticationURL defines the OAuth2 authenticationUrl.  See http://swagger.io/specification/#securitySchemeObject
	AuthenticationURL string
	// Scopes defines a list of scopes for the security method.
	Scopes []string
}

// Use sets the middleware that will implement the actual security
// mechanisms, most probably in user code or in some shared packages.
func (sec *OAuth2Security) Use(m Middleware) {
	sec.middleware = m
	sec.method = sec
}

///////////////////////////////////////////////////////////////////

// BasicAuthSecurity represents the `Basic` security method, which
// consists of a simple login/pass, accessible through
// Request.BasicAuth().
type BasicAuthSecurity struct {
	securityMiddleware
}

// Use sets the middleware that will implement the actual security
// mechanisms, most probably in user code or in some shared packages.
func (sec *BasicAuthSecurity) Use(m Middleware) {
	sec.middleware = m
	sec.method = sec
}

///////////////////////////////////////////////////////////////////

// APIKeySecurity represents the `apiKey` security method. It handles
// a key that can be in the headers or in the query parameters, and
// does authentication based on that.  The Name field represents the
// key of either the query string parameter or the header, depending
// on the In field.
type APIKeySecurity struct {
	securityMiddleware

	// In represents where to check for some data, `query` or `header`
	In string
	// Name is the name of the `header` or `query` parameter to check for data.
	Name string
}

// Use sets the middleware that will implement the actual security
// mechanisms, most probably in user code or in some shared packages.
func (sec *APIKeySecurity) Use(m Middleware) {
	sec.middleware = m
	sec.method = sec
}

///////////////////////////////////////////////////////////////////

// JWTSecurity represents an api key based method, with support for
// scopes and a token URL.
type JWTSecurity struct {
	securityMiddleware

	// In represents where to check for the JWT, `query` or `header`
	In string
	// Name is the name of the `header` or `query` parameter to check for data.
	Name string
	// TokenURL defines the URL where you'd get the JWT tokens.
	TokenURL string
	// Scopes defines a list of scopes for the security method.
	Scopes []string
}

// Use sets the middleware that will implement the actual security
// mechanisms, most probably in user code or in some shared packages.
func (sec *JWTSecurity) Use(m Middleware) {
	sec.middleware = m
	sec.method = sec
}
