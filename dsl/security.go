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
//     var Basic = BasicAuthSecurity("basicauth", func() {
//         Description("Use your own password!")
//     })
//
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
//    var APIKey = APIKeySecurity("key", func() {
//          Description("Shared secret")
//    })
//
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
//    var OAuth2 = OAuth2Security("googauth", func() {
//        ImplicitFlow("/authorization")
//
//        Scope("api:write", "Write acess")
//        Scope("api:read", "Read access")
//    })
//
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

// JWTSecurity defines an HTTP security scheme where a JWT is passed in the
// request Authorization header as a bearer token to perform auth. This scheme
// supports defining scopes that endpoint may require to authorize the request.
// The scheme also supports specifying a token URL used to retrieve token
// values.
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
//    var JWT = JWTSecurity("jwt", func() {
//        Scope("system:write", "Write to the system")
//        Scope("system:read", "Read anything in there")
//    })
//
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

// Security defines authentication requirements to access a service or a service
// method.
//
// The requirement refers to one or more OAuth2Security, BasicAuthSecurity,
// APIKeySecurity or JWTSecurity security scheme. If the schemes include a
// OAuth2Security or JWTSecurity scheme then required scopes may be listed by
// name in the Security DSL. All the listed schemes must be validated by the
// client for the request to be authorized. Security may appear multiple times
// in the same scope in which case the client may validate any one of the
// requirements for the request to be authorized.
//
// Security must appear in a Service or Method expression.
//
// Security accepts an arbitrary number of security schemes as argument
// specified by name or by reference and an optional DSL function as last
// argument.
//
// Examples:
//
//    var _ = Service("calculator", func() {
//        // Override default API security requirements. Accept either basic
//        // auth or OAuth2 access token with "api:read" scope.
//        Security(BasicAuth)
//        Security("oauth2", func() {
//            Scope("api:read")
//        })
//
//        Method("add", func() {
//            Description("Add two operands")
//
//            // Override default service security requirements. Require
//            // both basic auth and OAuth2 access token with "api:write"
//            // scope.
//            Security(BasicAuth, "oauth2", func() {
//                Scope("api:write")
//            })
//
//            Payload(Operands)
//            Error(ErrBadRequest, ErrorResult)
//        })
//
//        Method("health-check", func() {
//            Description("Check health")
//
//            // Remove need for authorization for this endpoint.
//            NoSecurity()
//
//            Payload(Operands)
//            Error(ErrBadRequest, ErrorResult)
//        })
//    })
//
func Security(args ...interface{}) {
	var dsl func()
	{
		if d, ok := args[len(args)-1].(func()); ok {
			args = args[:len(args)-1]
			dsl = d
		}
	}

	var schemes []*expr.SchemeExpr
	{
		schemes = make([]*expr.SchemeExpr, len(args))
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
				schemes[i] = expr.DupScheme(val)
			default:
				eval.InvalidArgError("security scheme or security scheme name", val)
				return
			}
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
		Schemes: []*expr.SchemeExpr{
			&expr.SchemeExpr{Kind: expr.NoKind},
		},
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
//
// Username must appear in Payload or Type.
//
// Example:
//
//    Method("login", func() {
//        Security(Basic)
//        Payload(func() {
//            Username("user", String)
//            Password("pass", String)
//        })
//        HTTP(func() {
//            // The "Authorization" header is defined implicitly.
//            POST("/login")
//        })
//    })
//
func Username(name string, args ...interface{}) {
	args = useDSL(args, func() { Meta("security:username") })
	Attribute(name, args...)
}

// UsernameField is syntactic sugar to define a username attribute with the
// "rpc:tag" meta set with the value of the first argument.
//
// UsernameField takes the same arguments as Username with the addition of the
// tag value as the first argument.
//
func UsernameField(tag interface{}, name string, args ...interface{}) {
	args = useDSL(args, func() { Meta("security:username") })
	Field(tag, name, args...)
}

// Password defines the attribute used to provide the password to an endpoint
// secured with basic authentication. The parameters and usage of Password are
// the same as the goa DSL Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to compute the basic authentication Authorization header value.
//
// Password must appear in Payload or Type.
//
// Example:
//
//    Method("login", func() {
//        Security(Basic)
//        Payload(func() {
//            Username("user", String)
//            Password("pass", String)
//        })
//        HTTP(func() {
//            // The "Authorization" header is defined implicitly.
//            POST("/login")
//        })
//    })
//
func Password(name string, args ...interface{}) {
	args = useDSL(args, func() { Meta("security:password") })
	Attribute(name, args...)
}

// PasswordField is syntactic sugar to define a password attribute with the
// "rpc:tag" meta set with the value of the first argument.
//
// PasswordField takes the same arguments as Password with the addition of the
// tag value as the first argument.
//
func PasswordField(tag interface{}, name string, args ...interface{}) {
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
//
// APIKey must appear in Payload or Type.
//
// Example:
//
//    Method("secured_read", func() {
//        Security(APIKeyAuth)
//        Payload(func() {
//            APIKey("api_key", "key", String, "API key used to perform authorization")
//            Required("key")
//        })
//        Result(String)
//        HTTP(func() {
//            GET("/")
//            Param("key:k") // Provide the key as a query string param "k"
//        })
//    })
//
//    Method("secured_write", func() {
//        Security(APIKeyAuth)
//        Payload(func() {
//            APIKey("api_key", "key", String, "API key used to perform authorization")
//            Attribute("data", String, "Data to be written")
//            Required("key", "data")
//        })
//        HTTP(func() {
//            POST("/")
//            Header("key:Authorization") // Provide the key in Authorization header (default)
//        })
//    })
//
func APIKey(scheme, name string, args ...interface{}) {
	args = useDSL(args, func() { Meta("security:apikey:"+scheme, scheme) })
	Attribute(name, args...)
}

// APIKeyField is syntactic sugar to define an API key attribute with the
// "rpc:tag" meta set with the value of the first argument.
//
// APIKeyField takes the same arguments as APIKey with the addition of the
// tag value as the first argument.
//
func APIKeyField(tag interface{}, scheme, name string, args ...interface{}) {
	args = useDSL(args, func() { Meta("security:apikey:"+scheme, scheme) })
	Field(tag, name, args...)
}

// AccessToken defines the attribute used to provide the access token to an
// endpoint secured with OAuth2. The parameters and usage of AccessToken are the
// same as the goa DSL Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to initialize the Authorization header.
//
// AccessToken must appear in Payload or Type.
//
// Example:
//
//    Method("secured", func() {
//        Security(OAuth2)
//        Payload(func() {
//            AccessToken("token", String, "OAuth2 access token used to perform authorization")
//            Required("token")
//        })
//        Result(String)
//        HTTP(func() {
//            // The "Authorization" header is defined implicitly.
//            GET("/")
//        })
//    })
//
func AccessToken(name string, args ...interface{}) {
	args = useDSL(args, func() { Meta("security:accesstoken") })
	Attribute(name, args...)
}

// AccessTokenField is syntactic sugar to define an access token attribute with the
// "rpc:tag" meta set with the value of the first argument.
//
// AccessTokenField takes the same arguments as AccessToken with the addition of the
// tag value as the first argument.
//
func AccessTokenField(tag interface{}, name string, args ...interface{}) {
	args = useDSL(args, func() { Meta("security:accesstoken") })
	Field(tag, name, args...)
}

// Token defines the attribute used to provide the JWT to an endpoint secured
// via JWT. The parameters and usage of Token are the same as the goa DSL
// Attribute function.
//
// The generated code produced by goa uses the value of the corresponding
// payload field to initialize the Authorization header.
//
// Example:
//
//    Method("secured", func() {
//        Security(JWT)
//        Payload(func() {
//            Token("token", String, "JWT token used to perform authorization")
//            Required("token")
//        })
//        Result(String)
//        HTTP(func() {
//            // The "Authorization" header is defined implicitly.
//            GET("/")
//        })
//    })
//
func Token(name string, args ...interface{}) {
	args = useDSL(args, func() { Meta("security:token") })
	Attribute(name, args...)
}

// TokenField is syntactic sugar to define a JWT token attribute with the
// "rpc:tag" meta set with the value of the first argument.
//
// TokenField takes the same arguments as Token with the addition of the
// tag value as the first argument.
//
func TokenField(tag interface{}, name string, args ...interface{}) {
	args = useDSL(args, func() { Meta("security:token") })
	Field(tag, name, args...)
}

// Scope has two uses: in JWTSecurity or OAuth2Security it defines a scope
// supported by the scheme. In Security it lists required scopes.
//
// Scope must appear in Security, BasicSecurity, APIKeySecurity, JWTSecurity or OAuth2Security.
//
// Scope accepts one or two arguments: the first argument is the scope name and
// when used in JWTSecurity or OAuth2Security the second argument is a
// description.
//
// Example:
//
//    var JWT = JWTSecurity("JWT", func() {
//        Scope("api:read", "Read access") // Defines a scope
//        Scope("api:write", "Write access")
//    })
//
//    Method("secured", func() {
//        Security(JWT, func() {
//            Scope("api:read") // Required scope for auth
//        })
//    })
//
func Scope(name string, desc ...string) {
	switch current := eval.Current().(type) {
	case *expr.SecurityExpr:
		if len(desc) >= 1 {
			eval.ReportError("too many arguments")
			return
		}
		current.Scopes = append(current.Scopes, name)
	case *expr.SchemeExpr:
		if len(desc) > 1 {
			eval.ReportError("too many arguments")
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
func useDSL(args []interface{}, d func()) []interface{} {
	if len(args) == 0 {
		return []interface{}{d}
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
