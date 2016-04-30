package apidsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// Security defines an authentication requirements to access a goa Action.  When defined on a
// Resource, it applies to all Actions, unless overriden by individual actions.  When defined at the
// API level, it will apply to all resources by default, following the same logic.
//
// The scheme refers to previous definitions of either OAuth2Security, BasicAuthSecurity,
// APIKeySecurity or JWTSecurity.  It can be a string, corresponding to the first parameter of
// those definitions, or a SecuritySchemeDefinition, returned by those same functions.
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
	case *design.ResourceDefinition:
		parent.Security = def
	case *design.APIDefinition:
		parent.Security = def
	default:
		dslengine.IncompatibleDSL()
		return
	}
}

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
	case *design.ResourceDefinition:
		parent.Security = def
	default:
		dslengine.IncompatibleDSL()
		return
	}
}

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

// OAuth2Security defines the different Security schemes that are available throughout the API.
//
// Example:
//
//    OAuth2Security("googAuth", func() {
//        AccessCodeFlow(...)
//     // ImplicitFlow(...)
//     // PasswordFlow(...)
//     // ApplicationFlow(...)
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
//        TokenURL("http://example.com/token")
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

// Scope defines an authorization scope. Used within SecurityScheme, the description is required,
// explaining what the scope means. Within a Security block, only a scope is needed.
func Scope(name string, desc ...string) {
	switch parent := dslengine.CurrentDefinition().(type) {
	case *design.SecurityDefinition:
		if len(desc) == 1 {
			dslengine.ReportError("too many arguments")
			return
		}
		parent.Scopes = append(parent.Scopes, name)
	case *design.SecuritySchemeDefinition:
		if len(desc) == 0 {
			dslengine.ReportError("missing description")
			return
		}
		if parent.Scopes == nil {
			parent.Scopes = make(map[string]string)
		}
		parent.Scopes[name] = desc[0]
	default:
		dslengine.IncompatibleDSL()
	}
}

// inHeader is called by `Header()`, see documentation there.
func inHeader(headerName string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if parent.Kind == design.APIKeySecurityKind || parent.Kind == design.JWTSecurityKind {
			if parent.In != "" {
				dslengine.ReportError("'In' previously defined through Header or Query")
				return
			}
			parent.In = "header"
			parent.Name = headerName
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// Query defines that an APIKeySecurity or JWTSecurity implementation must check in the query
// parameter named "parameterName" to get the api key.
func Query(parameterName string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if parent.Kind == design.APIKeySecurityKind || parent.Kind == design.JWTSecurityKind {
			if parent.In != "" {
				dslengine.ReportError("'In' previously defined through Header or Query")
				return
			}
			parent.In = "query"
			parent.Name = parameterName
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// AccessCodeFlow defines an "access code" OAuth2 flow.  Use within an OAuth2Security definition.
func AccessCodeFlow(authorizationURL, tokenURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind {
			parent.Flow = "accessCode"
			parent.AuthorizationURL = authorizationURL
			parent.TokenURL = tokenURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}

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

// TokenURL defines a URL to get an access token.  If you are defining OAuth2 flows, use
// `ImplicitFlow`, `PasswordFlow`, `AccessCodeFlow` or `ApplicationFlow` instead. This will set an
// endpoint where you can obtain a JWT with the JWTSecurity scheme.
func TokenURL(tokenURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecuritySchemeDefinition); ok {
		if parent.Kind == design.JWTSecurityKind {
			parent.TokenURL = tokenURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}
