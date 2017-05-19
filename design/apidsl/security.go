package apidsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// Security can be used in: API, Action, Files, Resource
//
// Security defines an authentication requirements to access a goa Action.  When defined on a
// Resource, it applies to all Actions, unless overriden by individual actions.  When defined at the
// API level, it will apply to all resources by default, following the same logic.
//
// The scheme refers to previous definitions of either OAuth2Security, BasicAuthSecurity,
// APIKeySecurity or JWTSecurity.  It can be a string, corresponding to the first parameter of
// those definitions, or a SecuritySchemeDefinition, returned by those same functions. Examples:
//
//    Security(BasicAuth)
//
//    Security("oauth2", func() {
//        Scope("api:read")  // Requires "api:read" oauth2 scope
//    })
//
func Security(scheme interface{}, dsl ...func()) {
	var def *design.SecurityDefinition
	switch val := scheme.(type) {
	case string:
		def = &design.SecurityDefinition{}
		for _, scheme := range design.Design.SecuritySchemes {
			if scheme.SchemeName == val {
				def.Scheme = scheme
			}
		}
		if def.Scheme == nil {
			dslengine.ReportError("security scheme %q not found", val)
			return
		}
	case *design.SecuritySchemeDefinition:
		def = &design.SecurityDefinition{Scheme: val}
	default:
		dslengine.ReportError("invalid value for 'scheme' parameter, specify a string or a *SecuritySchemeDefinition")
		return
	}

	if len(dsl) != 0 {
		if !dslengine.Execute(dsl[0], def) {
			return
		}
	}

	parentDef := dslengine.CurrentDefinition()
	switch parent := parentDef.(type) {
	case *design.ActionDefinition:
		parent.Security = def
	case *design.FileServerDefinition:
		parent.Security = def
	case *design.ResourceDefinition:
		parent.Security = def
	case *design.APIDefinition:
		parent.Security = def
	default:
		dslengine.IncompatibleDSL()
		return
	}
}

// NoSecurity can be used in: API, Action, Files, Resource
//
// NoSecurity resets the authentication schemes for an Action or a Resource. It also prevents
// fallback to Resource or API-defined Security.
func NoSecurity() {
	def := &design.SecurityDefinition{
		Scheme: &design.SecuritySchemeDefinition{Kind: design.NoSecurityKind},
	}

	parentDef := dslengine.CurrentDefinition()
	switch parent := parentDef.(type) {
	case *design.ActionDefinition:
		parent.Security = def
	case *design.FileServerDefinition:
		parent.Security = def
	case *design.ResourceDefinition:
		parent.Security = def
	default:
		dslengine.IncompatibleDSL()
		return
	}
}

// BasicAuthSecurity is a top level DSL.
// BasicAuthSecurity defines a "basic" security scheme for the API.
//
// Example:
//
//     BasicAuthSecurity("password", func() {
//         Description("Use your own password!")
//     })
//
func BasicAuthSecurity(name string, dsl ...func()) *design.SecuritySchemeDefinition {
	switch dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition, *dslengine.TopLevelDefinition:
	default:
		dslengine.IncompatibleDSL()
		return nil
	}

	if securitySchemeRedefined(name) {
		return nil
	}

	def := &design.SecuritySchemeDefinition{
		Kind:       design.BasicAuthSecurityKind,
		SchemeName: name,
		Type:       "basic",
	}

	if len(dsl) != 0 {
		def.DSLFunc = dsl[0]
	}

	design.Design.SecuritySchemes = append(design.Design.SecuritySchemes, def)

	return def
}

func securitySchemeRedefined(name string) bool {
	for _, previousScheme := range design.Design.SecuritySchemes {
		if previousScheme.SchemeName == name {
			dslengine.ReportError("cannot redefine SecurityScheme with name %q", name)
			return true
		}
	}
	return false
}

// APIKeySecurity is a top level DSL.
// APIKeySecurity defines an "apiKey" security scheme available throughout the API.
//
// Example:
//
//     APIKeySecurity("key", func() {
//          Description("Shared secret")
//          Header("Authorization")
//    })
//
func APIKeySecurity(name string, dsl ...func()) *design.SecuritySchemeDefinition {
	switch dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition, *dslengine.TopLevelDefinition:
	default:
		dslengine.IncompatibleDSL()
		return nil
	}

	if securitySchemeRedefined(name) {
		return nil
	}

	def := &design.SecuritySchemeDefinition{
		Kind:       design.APIKeySecurityKind,
		SchemeName: name,
		Type:       "apiKey",
	}

	if len(dsl) != 0 {
		def.DSLFunc = dsl[0]
	}

	design.Design.SecuritySchemes = append(design.Design.SecuritySchemes, def)

	return def
}

// OAuth2Security is a top level DSL.
// OAuth2Security defines an OAuth2 security scheme. The child DSL must define one and exactly one
// flow. One of AccessCodeFlow, ImplicitFlow, PasswordFlow or ApplicationFlow. Each flow defines
// endpoints for retrieving OAuth2 authorization codes and/or refresh and access tokens. The
// endpoint URLs may be complete or may be just a path in which case the API scheme and host are
// used to build the full URL. See for example [Aaron Parecki's
// writeup](https://aaronparecki.com/2012/07/29/2/oauth2-simplified) for additional details on
// OAuth2 flows.
//
// The OAuth2 DSL also allows for defining scopes that must be associated with the incoming request
// token for successful authorization.
//
// Example:
//
//    OAuth2Security("googAuth", func() {
//        AccessCodeFlow("/authorization", "/token")
//     // ImplicitFlow("/authorization")
//     // PasswordFlow("/token"...)
//     // ApplicationFlow("/token")
//
//        Scope("my_system:write", "Write to the system")
//        Scope("my_system:read", "Read anything in there")
//    })
//
func OAuth2Security(name string, dsl ...func()) *design.SecuritySchemeDefinition {
	switch dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition, *dslengine.TopLevelDefinition:
	default:
		dslengine.IncompatibleDSL()
		return nil
	}

	if securitySchemeRedefined(name) {
		return nil
	}

	def := &design.SecuritySchemeDefinition{
		SchemeName: name,
		Kind:       design.OAuth2SecurityKind,
		Type:       "oauth2",
	}

	if len(dsl) != 0 {
		def.DSLFunc = dsl[0]
	}

	design.Design.SecuritySchemes = append(design.Design.SecuritySchemes, def)

	return def
}

// JWTSecurity is a top level DSL.
// JWTSecurity defines an APIKey security scheme, with support for Scopes and a TokenURL.
//
// Since Scopes and TokenURLs are not compatible with the Swagger specification, the swagger
// generator inserts comments in the description of the different elements on which they are
// defined.
//
// Example:
//
//    JWTSecurity("jwt", func() {
//        Header("Authorization")
//        TokenURL("https://example.com/token")
//        Scope("my_system:write", "Write to the system")
//        Scope("my_system:read", "Read anything in there")
//    })
//
func JWTSecurity(name string, dsl ...func()) *design.SecuritySchemeDefinition {
	switch dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition, *dslengine.TopLevelDefinition:
	default:
		dslengine.IncompatibleDSL()
		return nil
	}

	if securitySchemeRedefined(name) {
		return nil
	}

	def := &design.SecuritySchemeDefinition{
		SchemeName: name,
		Kind:       design.JWTSecurityKind,
		Type:       "apiKey",
	}

	if len(dsl) != 0 {
		def.DSLFunc = dsl[0]
	}

	design.Design.SecuritySchemes = append(design.Design.SecuritySchemes, def)

	return def
}

// Scope can be used in: Security, JWTSecurity, OAuth2Security
//
// Scope defines an authorization scope. Used within SecurityScheme, a description may be provided
// explaining what the scope means. Within a Security block, only a scope is needed.
func Scope(name string, desc ...string) {
	switch current := dslengine.CurrentDefinition().(type) {
	case *design.SecurityDefinition:
		if len(desc) >= 1 {
			dslengine.ReportError("too many arguments")
			return
		}
		current.Scopes = append(current.Scopes, name)
	case *design.SecuritySchemeDefinition:
		if len(desc) > 1 {
			dslengine.ReportError("too many arguments")
			return
		}
		if current.Scopes == nil {
			current.Scopes = make(map[string]string)
		}
		d := "no description"
		if len(desc) == 1 {
			d = desc[0]
		}
		current.Scopes[name] = d
	default:
		dslengine.IncompatibleDSL()
	}
}

// inHeader is called by `Header()`, see documentation there.
func inHeader(headerName string) {
	if current, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if current.Kind == design.APIKeySecurityKind || current.Kind == design.JWTSecurityKind {
			if current.In != "" {
				dslengine.ReportError("'In' previously defined through Header or Query")
				return
			}
			current.In = "header"
			current.Name = headerName
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// Query can be used in: APIKeySecurity, JWTSecurity
//
// Query defines that an APIKeySecurity or JWTSecurity implementation must check in the query
// parameter named "parameterName" to get the api key.
func Query(parameterName string) {
	if current, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if current.Kind == design.APIKeySecurityKind || current.Kind == design.JWTSecurityKind {
			if current.In != "" {
				dslengine.ReportError("'In' previously defined through Header or Query")
				return
			}
			current.In = "query"
			current.Name = parameterName
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// AccessCodeFlow can be used in: OAuth2Security
//
// AccessCodeFlow defines an "access code" OAuth2 flow.  Use within an OAuth2Security definition.
func AccessCodeFlow(authorizationURL, tokenURL string) {
	if current, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if current.Kind == design.OAuth2SecurityKind {
			current.Flow = "accessCode"
			current.AuthorizationURL = authorizationURL
			current.TokenURL = tokenURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// ApplicationFlow can be used in: OAuth2Security
//
// ApplicationFlow defines an "application" OAuth2 flow.  Use within an OAuth2Security definition.
func ApplicationFlow(tokenURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind {
			parent.Flow = "application"
			parent.TokenURL = tokenURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// PasswordFlow can be used in: OAuth2Security
//
// PasswordFlow defines a "password" OAuth2 flow.  Use within an OAuth2Security definition.
func PasswordFlow(tokenURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind {
			parent.Flow = "password"
			parent.TokenURL = tokenURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// ImplicitFlow can be used in: OAuth2Security
//
// ImplicitFlow defines an "implicit" OAuth2 flow.  Use within an OAuth2Security definition.
func ImplicitFlow(authorizationURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind {
			parent.Flow = "implicit"
			parent.AuthorizationURL = authorizationURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// TokenURL can be used in: JWTSecurity
//
// TokenURL defines a URL to get an access token.  If you are defining OAuth2 flows, use
// `ImplicitFlow`, `PasswordFlow`, `AccessCodeFlow` or `ApplicationFlow` instead. This will set an
// endpoint where you can obtain a JWT with the JWTSecurity scheme. The URL may be a complete URL
// or just a path in which case the API scheme and host are used to build the full URL.
func TokenURL(tokenURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if parent.Kind == design.JWTSecurityKind {
			parent.TokenURL = tokenURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}
