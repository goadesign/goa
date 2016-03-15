package apidsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// Security defines an authentication requirements to access a goa
// Action.  When defined on a Resource, it applies to all Actions,
// unless overriden by individual actions.  When defined at the API
// level, it will apply to all resources by default, following the
// same logic.
//
// The method refers to previous definitions of either
// OAuth2Security(), BasicAuthSecurity(), APIKeySecurity() or
// JWTSecurity().  It can be a string, corresponding to the first
// parameter of those definitions, or a SecurityMethodDefinition,
// returned by those same functions.
func Security(method interface{}, dsl ...func()) {
	var def *design.SecurityDefinition
	switch val := method.(type) {
	case string:
		def = &design.SecurityDefinition{Method: val}
	case *design.SecurityMethodDefinition:
		def = &design.SecurityDefinition{Method: val.Method}
	default:
		dslengine.ReportError("invalid value for 'method' parameter, specify a string or a *SecurityMethodDefinition")
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

// NoSecurity resets the authentication methods for an Action or a
// Resource. It also prevents fallback to Resource or API-defined
// Security().
func NoSecurity() {
	def := &design.SecurityDefinition{NoSecurity: true}

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

// BasicAuthSecurity defines a "basic" security method for the API.
//
// Example:
//
//     BasicAuthSecurity("password", func() {
//         Description("Use your own password!")
//     })
//
func BasicAuthSecurity(name string, dsl ...func()) *design.SecurityMethodDefinition {
	switch dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition, *dslengine.TopLevelDefinition:
	default:
		dslengine.IncompatibleDSL()
		return nil
	}

	if securityMethodRedefined(name) {
		return nil
	}

	def := &design.SecurityMethodDefinition{
		Kind:   design.BasicAuthSecurityKind,
		Method: name,
		Type:   "basic",
	}

	if len(dsl) != 0 {
		if !dslengine.Execute(dsl[0], def) {
			return nil
		}
	}

	design.Design.SecurityMethods = append(design.Design.SecurityMethods, def)

	return def
}

func securityMethodRedefined(name string) bool {
	for _, previousMethod := range design.Design.SecurityMethods {
		if previousMethod.Method == name {
			dslengine.ReportError("cannot redefine SecurityMethod with name %q", name)
			return true
		}
	}
	return false
}

// APIKeySecurity defines an "apiKey" security method available throughout the API.
//
// Example:
//
//     APIKeySecurity("jwt", func() {
//			Description("Use your own password!")
//    		Header("Authorization")
//     })
//
func APIKeySecurity(name string, dsl ...func()) *design.SecurityMethodDefinition {
	switch dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition, *dslengine.TopLevelDefinition:
	default:
		dslengine.IncompatibleDSL()
		return nil
	}

	if securityMethodRedefined(name) {
		return nil
	}

	def := &design.SecurityMethodDefinition{
		Kind:   design.APIKeySecurityKind,
		Method: name,
		Type:   "apiKey",
	}

	if len(dsl) != 0 {
		if !dslengine.Execute(dsl[0], def) {
			return nil
		}
	}

	design.Design.SecurityMethods = append(design.Design.SecurityMethods, def)

	return def
}

// OAuth2Security defines the different Security methods that are
// available throughout the API.
//
// Example:
//
//     OAuth2Security("googAuth", func() {
//	    	AccessCodeFlow(...)
//	    	// ImplicitFlow(...)
//	    	// PasswordFlow(...)
//	    	// ApplicationFlow(...)
//
//	    	Scope("my_system:write", "Write to the system")
//	    	Scope("my_system:read", "Read anything in there")
//     })
//
func OAuth2Security(name string, dsl ...func()) *design.SecurityMethodDefinition {
	switch dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition, *dslengine.TopLevelDefinition:
	default:
		dslengine.IncompatibleDSL()
		return nil
	}

	if securityMethodRedefined(name) {
		return nil
	}

	def := &design.SecurityMethodDefinition{
		Method: name,
		Kind:   design.OAuth2SecurityKind,
		Type:   "oauth2",
	}

	if len(dsl) != 0 {
		if !dslengine.Execute(dsl[0], def) {
			return nil
		}
	}

	design.Design.SecurityMethods = append(design.Design.SecurityMethods, def)

	return def
}

// JWTSecurity defines an APIKey security method, with support for
// Scopes and a TokenURL.
//
// Since Scopes and TokenURLs are not compatible with the Swagger specification,
// the swagger generator inserts comments in the description of the different
// elements on which they are defined.
//
// Example:
//
//     JWTSecurity("jwt", func() {
//          Header("Authorization")
//	    	TokenURL("http://example.com/token")
//	    	Scope("my_system:write", "Write to the system")
//	    	Scope("my_system:read", "Read anything in there")
//     })
//
func JWTSecurity(name string, dsl ...func()) *design.SecurityMethodDefinition {
	switch dslengine.CurrentDefinition().(type) {
	case *design.APIDefinition, *dslengine.TopLevelDefinition:
	default:
		dslengine.IncompatibleDSL()
		return nil
	}

	if securityMethodRedefined(name) {
		return nil
	}

	def := &design.SecurityMethodDefinition{
		Method: name,
		Kind:   design.JWTSecurityKind,
		Type:   "apiKey",
	}

	if len(dsl) != 0 {
		if !dslengine.Execute(dsl[0], def) {
			return nil
		}
	}

	design.Design.SecurityMethods = append(design.Design.SecurityMethods, def)

	return def
}

// Scope defines an authorization scope. Used within SecurityMethod,
// the description is required, explaining what the scope
// means. Within a Security block, only a scope is needed.
func Scope(name string, desc ...string) {
	switch parent := dslengine.CurrentDefinition().(type) {
	case *design.SecurityDefinition:
		if len(desc) == 1 {
			dslengine.ReportError("too many arguments")
			return
		}
		parent.Scopes = append(parent.Scopes, name)
	case *design.SecurityMethodDefinition:
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
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
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

// Query defines that an APIKeySecurity or JWTSecurity implementation
// must check in the query parameter named "parameterName" to get the
// api key.
func Query(parameterName string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
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

// AccessCodeFlow defines an "access code" OAuth2 flow.  Use
// within an OAuth2Security definition.
func AccessCodeFlow(authorizationURL, tokenURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind {
			parent.Flow = "accessCode"
			parent.AuthorizationURL = authorizationURL
			parent.TokenURL = tokenURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// ApplicationFlow defines an "application" OAuth2 flow.  Use
// within an OAuth2Security definition.
func ApplicationFlow(tokenURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind {
			parent.Flow = "application"
			parent.TokenURL = tokenURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}


// PasswordFlow defines a "password" OAuth2 flow.  Use within an
// OAuth2Security definition.
func PasswordFlow(tokenURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind {
			parent.Flow = "password"
			parent.TokenURL = tokenURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// ImplicitFlow defines an "implicit" OAuth2 flow.  Use within an
// OAuth2Security definition.
func ImplicitFlow(authorizationURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind {
			parent.Flow = "implicit"
			parent.AuthorizationURL = authorizationURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// TokenURL defines a URL to get an access token.  If you are defining
// OAuth2 flows, please use `ImplicitFlow`, `PasswordFlow`,
// `AccessCodeFlow` or `ApplicationFlow` instead. This will set an
// endpoint where you can obtain a JWT with the JWTSecurity method.
func TokenURL(tokenURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
		if parent.Kind == design.JWTSecurityKind {
			parent.TokenURL = tokenURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}
