package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// BasicAuthSecurity defines a basic authentication security scheme.
//
// BasicAuthSecurity is a top level DSL.
//
// BasicAuthSecurity takes a name as first argument and an optional DSL as
// second argument.
//
// Example:
//
//	var Basic = BasicAuthSecurity("basicauth", func() {
//	    Description("Use your own password!")
//	})
func BasicAuthSecurity(name string, fn ...func()) *expr.SchemeExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	if securitySchemeRedefined(name) {
		return nil
	}

	e := &expr.SchemeExpr{
		Kind:       expr.BasicAuthKind,
		SchemeName: name,
	}

	if len(fn) != 0 {
		if !eval.Execute(fn[0], e) {
			return nil
		}
	}

	expr.Root.Schemes = append(expr.Root.Schemes, e)

	return e
}

// APIKeySecurity defines an API key security scheme where a key must be
// provided by the client to perform authorization.
//
// APIKeySecurity is a top level DSL.
//
// APIKeySecurity takes a name as first argument and an optional DSL as
// second argument.
//
// Example:
//
//	var APIKey = APIKeySecurity("key", func() {
//	      Description("Shared secret")
//	})
func APIKeySecurity(name string, fn ...func()) *expr.SchemeExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	if securitySchemeRedefined(name) {
		return nil
	}

	e := &expr.SchemeExpr{
		Kind:       expr.APIKeyKind,
		SchemeName: name,
	}

	if len(fn) != 0 {
		if !eval.Execute(fn[0], e) {
			return nil
		}
	}

	expr.Root.Schemes = append(expr.Root.Schemes, e)

	return e
}

// OAuth2Security defines an OAuth2 security scheme. The DSL provided as second
// argument defines the specific flows supported by the scheme. The supported
// flow types are ImplicitFlow, PasswordFlow, ClientCredentialsFlow, and
// AuthorizationCodeFlow. The DSL also defines the scopes that may be
// associated with the incoming request tokens.
//
// OAuth2Security is a top level DSL.
//
// OAuth2Security takes a name as first argument and a DSL as second argument.
//
// Example:
//
//	var OAuth2 = OAuth2Security("googauth", func() {
//	    ImplicitFlow("/authorization")
//
//	    Scope("api:write", "Write acess")
//	    Scope("api:read", "Read access")
//	})
func OAuth2Security(name string, fn ...func()) *expr.SchemeExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	if securitySchemeRedefined(name) {
		return nil
	}

	e := &expr.SchemeExpr{
		SchemeName: name,
		Kind:       expr.OAuth2Kind,
	}

	if len(fn) != 0 {
		if !eval.Execute(fn[0], e) {
			return nil
		}
	}

	expr.Root.Schemes = append(expr.Root.Schemes, e)

	return e
}

// BearerSecurity defines an HTTP Bearer security scheme where a token is passed
// in the request Authorization header using the "Bearer" authentication scheme.
// It models the standard OpenAPI v3 HTTP bearer scheme:
//
//	type: http
//	scheme: bearer
//
// Use BearerSecurity when the token format is generic, opaque, or otherwise
// not part of the Goa service contract. This includes APIs that accept opaque
// session tokens, JWT access tokens, or both under the same Bearer header.
//
// BearerSecurity is distinct from JWTSecurity. Both schemes use the same
// default HTTP wire format, "Authorization: Bearer <token>", but they generate
// different DSL/runtime names. BearerSecurity generates BearerAuth,
// security.BearerScheme and uses BearerToken payload attributes. JWTSecurity is
// the JWT-specific variant kept for designs that intentionally want JWT names.
//
// This scheme supports defining scopes that an endpoint may require to
// authorize the request. The scheme also supports specifying a bearer format
// hint for OpenAPI v3 with BearerFormat. The bearer format is documentation
// only; Goa still exposes the raw token string to the generated auth function.
//
// BearerSecurity is a top level DSL.
//
// BearerSecurity takes a name as first argument and an optional DSL as second
// argument.
//
// Example:
//
//	var Bearer = BearerSecurity("bearer", func() {
//	    Description("Opaque session token or trusted OIDC access token.")
//	    Scope("system:write", "Write to the system")
//	    Scope("system:read", "Read anything in there")
//	})
func BearerSecurity(name string, fn ...func()) *expr.SchemeExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	if securitySchemeRedefined(name) {
		return nil
	}

	e := &expr.SchemeExpr{
		SchemeName: name,
		Kind:       expr.BearerKind,
		In:         "header",
		Name:       "Authorization",
	}

	if len(fn) != 0 {
		if !eval.Execute(fn[0], e) {
			return nil
		}
	}

	expr.Root.Schemes = append(expr.Root.Schemes, e)

	return e
}

// JWTSecurity defines an HTTP security scheme where a JWT is passed in the
// request Authorization header as a bearer token to perform auth. Use
// JWTSecurity when the token is specifically a JWT and the generated DSL and
// runtime names should say JWT.
//
// JWTSecurity is not a different HTTP authentication protocol from
// BearerSecurity. Both schemes default to "Authorization: Bearer <token>" and
// both expose the raw token string to the generated auth function. The
// distinction is semantic and affects generated names: JWTSecurity generates
// JWTAuth, security.JWTScheme and uses Token payload attributes, while
// BearerSecurity generates BearerAuth, security.BearerScheme and uses
// BearerToken payload attributes.
//
// This scheme supports defining scopes that an endpoint may require to
// authorize the request. OpenAPI v3 output uses "JWT" as the default bearer
// format hint for JWTSecurity. The bearer format is documentation only; Goa
// does not parse or validate JWT claims. Use BearerFormat to override the
// OpenAPI v3 hint.
//
// Since scopes are not compatible with the Swagger specification, the swagger
// generator inserts comments in the description of the different elements on
// which they are defined.
//
// JWTSecurity is a top level DSL.
//
// JWTSecurity takes a name as first argument and an optional DSL as second
// argument.
//
// Example:
//
//	var JWT = JWTSecurity("jwt", func() {
//	    Scope("system:write", "Write to the system")
//	    Scope("system:read", "Read anything in there")
//	})
func JWTSecurity(name string, fn ...func()) *expr.SchemeExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	if securitySchemeRedefined(name) {
		return nil
	}

	e := &expr.SchemeExpr{
		SchemeName: name,
		Kind:       expr.JWTKind,
		In:         "header",
		Name:       "Authorization",
	}

	if len(fn) != 0 {
		if !eval.Execute(fn[0], e) {
			return nil
		}
	}

	expr.Root.Schemes = append(expr.Root.Schemes, e)

	return e
}

// BearerFormat sets the format hint for a bearer token security scheme. The
// value is emitted as the OpenAPI v3 bearerFormat field. JWTSecurity emits
// "JWT" by default when the bearer format is empty. The field is an OpenAPI
// documentation hint only and does not change generated token extraction or
// validation behavior.
//
// BearerFormat must appear in BearerSecurity or JWTSecurity.
func BearerFormat(format string) {
	current, ok := eval.Current().(*expr.SchemeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if current.Kind != expr.BearerKind && current.Kind != expr.JWTKind {
		eval.ReportError("cannot specify bearer format for non-bearer security scheme.")
		return
	}
	current.BearerFormat = format
}

// Security defines authentication requirements to access an entire API, service
// or individual service method.
//
// The requirement refers to one or more OAuth2Security, BasicAuthSecurity,
// APIKeySecurity, BearerSecurity or JWTSecurity security scheme. If the schemes
// include a BearerSecurity, OAuth2Security or JWTSecurity scheme then required
// scopes may be listed by name in the Security DSL. All the listed schemes must
// be validated by the client for the request to be authorized. Security may
// appear multiple times in the same scope in which case the client may validate
// any one of the requirements for the request to be authorized.
//
// Security must appear in an API, Service or Method expression.
//
// Security accepts an arbitrary number of security schemes as argument
// specified by name or by reference and an optional DSL function as last
// argument.
//
// Examples:
//
//	var _ = Service("calculator", func() {
//	    // Override default API security requirements. Accept either basic
//	    // auth or OAuth2 access token with "api:read" scope.
//	    Security(BasicAuth)
//	    Security("oauth2", func() {
//	        Scope("api:read")
//	    })
//
//	    Method("add", func() {
//	        Description("Add two operands")
//
//	        // Override default service security requirements. Require
//	        // both basic auth and OAuth2 access token with "api:write"
//	        // scope.
//	        Security(BasicAuth, "oauth2", func() {
//	            Scope("api:write")
//	        })
//
//	        Payload(Operands)
//	        Error(ErrBadRequest, ErrorResult)
//	    })
//
//	    Method("health-check", func() {
//	        Description("Check health")
//
//	        // Remove need for authorization for this endpoint.
//	        NoSecurity()
//
//	        Payload(Operands)
//	        Error(ErrBadRequest, ErrorResult)
//	    })
//	})
func Security(args ...any) {
	var dsl func()
	if d, ok := args[len(args)-1].(func()); ok {
		args = args[:len(args)-1]
		dsl = d
	}

	schemes := make([]*expr.SchemeExpr, len(args))
	for i, arg := range args {
		switch val := arg.(type) {
		case string:
			for _, s := range expr.Root.Schemes {
				if s.SchemeName == val {
					schemes[i] = expr.DupScheme(s)
					break
				}
			}
			if schemes[i] == nil {
				eval.ReportError("security scheme %q not found", val)
				return
			}
		case *expr.SchemeExpr:
			if val == nil {
				eval.InvalidArgError("security scheme", val)
				return
			}
			schemes[i] = expr.DupScheme(val)
		default:
			eval.InvalidArgError("security scheme or security scheme name", val)
			return
		}
	}

	security := &expr.SecurityExpr{Schemes: schemes}
	if dsl != nil {
		if !eval.Execute(dsl, security) {
			return
		}
	}

	current := eval.Current()
	switch actual := current.(type) {
	case *expr.MethodExpr:
		actual.Requirements = append(actual.Requirements, security)
	case *expr.ServiceExpr:
		actual.Requirements = append(actual.Requirements, security)
	case *expr.APIExpr:
		actual.Requirements = append(actual.Requirements, security)
	case expr.SecurityHolder:
		actual.AddSecurityRequirement(security)
	default:
		eval.IncompatibleDSL()
		return
	}
}

// NoSecurity removes the need for an endpoint to perform authorization.
//
// NoSecurity must appear in Method.
func NoSecurity() {
	security := &expr.SecurityExpr{
		Schemes: []*expr.SchemeExpr{{Kind: expr.NoKind}},
	}

	current := eval.Current()
	switch actual := current.(type) {
	case *expr.MethodExpr:
		actual.Requirements = append(actual.Requirements, security)
	default:
		eval.IncompatibleDSL()
		return
	}
}

// Username defines the attribute used to provide the username to an endpoint
// secured with basic authentication. The parameters and usage of Username are
// the same as the goa DSL Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to compute the basic authentication Authorization header value.
// The attribute must use String or a named String type and cannot have a default.
//
// Username must appear in Payload or Type.
//
// Example:
//
//	Method("login", func() {
//	    Security(Basic)
//	    Payload(func() {
//	        Username("user", String)
//	        Password("pass", String)
//	    })
//	    HTTP(func() {
//	        // The "Authorization" header is defined implicitly.
//	        POST("/login")
//	    })
//	})
func Username(name string, args ...any) {
	args = useDSL(args, func() { Meta("security:username") })
	Attribute(name, args...)
}

// UsernameField is syntactic sugar to define a username attribute with the
// "rpc:tag" meta set with the value of the first argument.
//
// UsernameField takes the same arguments as Username with the addition of the
// tag value as the first argument.
func UsernameField(tag any, name string, args ...any) {
	args = useDSL(args, func() { Meta("security:username") })
	Field(tag, name, args...)
}

// Password defines the attribute used to provide the password to an endpoint
// secured with basic authentication. The parameters and usage of Password are
// the same as the goa DSL Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to compute the basic authentication Authorization header value.
// The attribute must use String or a named String type and cannot have a default.
//
// Password must appear in Payload or Type.
//
// Example:
//
//	Method("login", func() {
//	    Security(Basic)
//	    Payload(func() {
//	        Username("user", String)
//	        Password("pass", String)
//	    })
//	    HTTP(func() {
//	        // The "Authorization" header is defined implicitly.
//	        POST("/login")
//	    })
//	})
func Password(name string, args ...any) {
	args = useDSL(args, func() { Meta("security:password") })
	Attribute(name, args...)
}

// PasswordField is syntactic sugar to define a password attribute with the
// "rpc:tag" meta set with the value of the first argument.
//
// PasswordField takes the same arguments as Password with the addition of the
// tag value as the first argument.
func PasswordField(tag any, name string, args ...any) {
	args = useDSL(args, func() { Meta("security:password") })
	Field(tag, name, args...)
}

// APIKey defines the attribute used to provide the API key to an endpoint
// secured with API keys. The parameters and usage of APIKey are the same as the
// Attribute function except that it accepts an extra first argument
// corresponding to the name of the API key security scheme.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to set the API key value.
// The attribute must use String or a named String type and cannot have a default.
//
// APIKey must appear in Payload or Type.
//
// Example:
//
//	Method("secured_read", func() {
//	    Security(APIKeyAuth)
//	    Payload(func() {
//	        APIKey("api_key", "key", String, "API key used to perform authorization")
//	        Required("key")
//	    })
//	    Result(String)
//	    HTTP(func() {
//	        GET("/")
//	        Param("key:k") // Provide the key as a query string param "k"
//	    })
//	})
//
//	Method("secured_write", func() {
//	    Security(APIKeyAuth)
//	    Payload(func() {
//	        APIKey("api_key", "key", String, "API key used to perform authorization")
//	        Attribute("data", String, "Data to be written")
//	        Required("key", "data")
//	    })
//	    HTTP(func() {
//	        POST("/")
//	        Header("key:Authorization") // Provide the key in Authorization header (default)
//	    })
//	})
func APIKey(scheme, name string, args ...any) {
	args = useDSL(args, func() { Meta("security:apikey:"+scheme, scheme) })
	Attribute(name, args...)
}

// APIKeyField is syntactic sugar to define an API key attribute with the
// "rpc:tag" meta set with the value of the first argument.
//
// APIKeyField takes the same arguments as APIKey with the addition of the
// tag value as the first argument.
func APIKeyField(tag any, scheme, name string, args ...any) {
	args = useDSL(args, func() { Meta("security:apikey:"+scheme, scheme) })
	Field(tag, name, args...)
}

// AccessToken defines the attribute used to provide the access token to an
// endpoint secured with OAuth2. The parameters and usage of AccessToken are the
// same as the goa DSL Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to initialize the Authorization header.
// The attribute must use String or a named String type and cannot have a default.
//
// AccessToken must appear in Payload or Type.
//
// Example:
//
//	Method("secured", func() {
//	    Security(OAuth2)
//	    Payload(func() {
//	        AccessToken("token", String, "OAuth2 access token used to perform authorization")
//	        Required("token")
//	    })
//	    Result(String)
//	    HTTP(func() {
//	        // The "Authorization" header is defined implicitly.
//	        GET("/")
//	    })
//	})
func AccessToken(name string, args ...any) {
	args = useDSL(args, func() { Meta("security:accesstoken") })
	Attribute(name, args...)
}

// AccessTokenField is syntactic sugar to define an access token attribute with the
// "rpc:tag" meta set with the value of the first argument.
//
// AccessTokenField takes the same arguments as AccessToken with the addition of the
// tag value as the first argument.
func AccessTokenField(tag any, name string, args ...any) {
	args = useDSL(args, func() { Meta("security:accesstoken") })
	Field(tag, name, args...)
}

// BearerToken defines the attribute used to provide the raw token to an
// endpoint secured via BearerSecurity. The parameters and usage of BearerToken
// are the same as the goa DSL Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to initialize the Authorization header.
// The attribute must use String or a named String type and cannot have a default.
//
// Example:
//
//	Method("secured", func() {
//	    Security(Bearer)
//	    Payload(func() {
//	        BearerToken("token", String, "Bearer token used to perform authorization")
//	        Required("token")
//	    })
//	    Result(String)
//	    HTTP(func() {
//	        // The "Authorization" header is defined implicitly.
//	        GET("/")
//	    })
//	})
func BearerToken(name string, args ...any) {
	args = useDSL(args, func() { Meta("security:bearer") })
	Attribute(name, args...)
}

// BearerTokenField is syntactic sugar to define a bearer token attribute with
// the "rpc:tag" meta set with the value of the first argument.
//
// BearerTokenField takes the same arguments as BearerToken with the addition of
// the tag value as the first argument.
func BearerTokenField(tag any, name string, args ...any) {
	args = useDSL(args, func() { Meta("security:bearer") })
	Field(tag, name, args...)
}

// Token defines the attribute used to provide the raw JWT bearer token to an
// endpoint secured via JWTSecurity. The parameters and usage of Token are the
// same as the goa DSL Attribute function.
//
// Use BearerToken instead when the endpoint is secured via BearerSecurity and
// the generated names should not be JWT-specific.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to initialize the Authorization header.
// The attribute must use String or a named String type and cannot have a default.
//
// Example:
//
//	Method("secured", func() {
//	    Security(JWT)
//	    Payload(func() {
//	        Token("token", String, "JWT token used to perform authorization")
//	        Required("token")
//	    })
//	    Result(String)
//	    HTTP(func() {
//	        // The "Authorization" header is defined implicitly.
//	        GET("/")
//	    })
//	})
func Token(name string, args ...any) {
	args = useDSL(args, func() { Meta("security:token") })
	Attribute(name, args...)
}

// TokenField is syntactic sugar to define a JWT token attribute with the
// "rpc:tag" meta set with the value of the first argument.
//
// TokenField takes the same arguments as Token with the addition of the
// tag value as the first argument.
func TokenField(tag any, name string, args ...any) {
	args = useDSL(args, func() { Meta("security:token") })
	Field(tag, name, args...)
}

// Scope has two uses: in BearerSecurity, JWTSecurity or OAuth2Security it
// defines a scope supported by the scheme. In Security it lists required scopes.
// Scopes on BearerSecurity and JWTSecurity are Goa authorization metadata passed
// to the generated auth function; they are not OpenAPI OAuth2 scopes.
//
// Scope must appear in Security, BasicSecurity, APIKeySecurity,
// BearerSecurity, JWTSecurity or OAuth2Security.
//
// Scope accepts one or two arguments: the first argument is the scope name and
// when used in BearerSecurity, JWTSecurity or OAuth2Security the second
// argument is a description.
//
// Example:
//
//	var JWT = JWTSecurity("JWT", func() {
//	    Scope("api:read", "Read access") // Defines a scope
//	    Scope("api:write", "Write access")
//	})
//
//	Method("secured", func() {
//	    Security(JWT, func() {
//	        Scope("api:read") // Required scope for auth
//	    })
//	})
func Scope(name string, desc ...string) {
	switch current := eval.Current().(type) {
	case *expr.SecurityExpr:
		if len(desc) >= 1 {
			eval.TooManyArgError()
			return
		}
		current.Scopes = append(current.Scopes, name)
	case *expr.SchemeExpr:
		if len(desc) > 1 {
			eval.TooManyArgError()
			return
		}
		d := "no description"
		if len(desc) == 1 {
			d = desc[0]
		}
		current.Scopes = append(current.Scopes,
			&expr.ScopeExpr{Name: name, Description: d})
	default:
		eval.IncompatibleDSL()
	}
}

// AuthorizationCodeFlow defines an authorizationCode OAuth2 flow as described
// in section 1.3.1 of RFC 6749.
//
// AuthorizationCodeFlow must be used in OAuth2Security.
//
// AuthorizationCodeFlow accepts three arguments: the authorization, token and
// refresh URLs.
func AuthorizationCodeFlow(authorizationURL, tokenURL, refreshURL string) {
	current, ok := eval.Current().(*expr.SchemeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if current.Kind != expr.OAuth2Kind {
		eval.ReportError("cannot specify flow for non-oauth2 security scheme.")
		return
	}
	current.Flows = append(current.Flows, &expr.FlowExpr{
		Kind:             expr.AuthorizationCodeFlowKind,
		AuthorizationURL: authorizationURL,
		TokenURL:         tokenURL,
		RefreshURL:       refreshURL,
	})
}

// ImplicitFlow defines an implicit OAuth2 flow as described in section 1.3.2
// of RFC 6749.
//
// ImplicitFlow must be used in OAuth2Security.
//
// ImplicitFlow accepts two arguments: the authorization and refresh URLs.
func ImplicitFlow(authorizationURL, refreshURL string) {
	current, ok := eval.Current().(*expr.SchemeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if current.Kind != expr.OAuth2Kind {
		eval.ReportError("cannot specify flow for non-oauth2 security scheme.")
		return
	}
	current.Flows = append(current.Flows, &expr.FlowExpr{
		Kind:             expr.ImplicitFlowKind,
		AuthorizationURL: authorizationURL,
		RefreshURL:       refreshURL,
	})
}

// PasswordFlow defines an Resource Owner Password Credentials OAuth2 flow as
// described in section 1.3.3 of RFC 6749.
//
// PasswordFlow must be used in OAuth2Security.
//
// PasswordFlow accepts two arguments: the token and refresh URLs.
func PasswordFlow(tokenURL, refreshURL string) {
	current, ok := eval.Current().(*expr.SchemeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if current.Kind != expr.OAuth2Kind {
		eval.ReportError("cannot specify flow for non-oauth2 security scheme.")
		return
	}
	current.Flows = append(current.Flows, &expr.FlowExpr{
		Kind:       expr.PasswordFlowKind,
		TokenURL:   tokenURL,
		RefreshURL: refreshURL,
	})
}

// ClientCredentialsFlow defines an clientCredentials OAuth2 flow as described
// in section 1.3.4 of RFC 6749.
//
// ClientCredentialsFlow must be used in OAuth2Security.
//
// ClientCredentialsFlow accepts two arguments: the token and refresh URLs.
func ClientCredentialsFlow(tokenURL, refreshURL string) {
	current, ok := eval.Current().(*expr.SchemeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if current.Kind != expr.OAuth2Kind {
		eval.ReportError("cannot specify flow for non-oauth2 security scheme.")
		return
	}
	current.Flows = append(current.Flows, &expr.FlowExpr{
		Kind:       expr.ClientCredentialsFlowKind,
		TokenURL:   tokenURL,
		RefreshURL: refreshURL,
	})
}

func securitySchemeRedefined(name string) bool {
	for _, s := range expr.Root.Schemes {
		if s.SchemeName == name {
			eval.ReportError("cannot redefine security scheme with name %q", name)
			return true
		}
	}
	return false
}

// useDSL modifies the Attribute function to use the given function as DSL,
// merging it with any pre-existing DSL.
func useDSL(args []any, d func()) []any {
	if len(args) == 0 {
		return []any{d}
	}
	ds, ok := args[len(args)-1].(func())
	if ok {
		newdsl := func() { ds(); d() }
		args = append(args[:len(args)-1], newdsl)
	} else {
		args = append(args, d)
	}
	return args
}
