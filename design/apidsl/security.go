package apidsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
)

// Security defines an authentication method required to access a goa
// Action.  When defined on a Resource, it applies to all Actions,
// unless overriden by individual actions.  When defined at the API
// level, it will apply to all resources by default, following the
// same logic.
//
// The method name refers to what you have passed to the different
// OAuth2Security(), BasicAuthSecurity() and APIKeySecurity()
// declarations.
func Security(method string, dsl ...func()) {
	def := &design.SecurityDefinition{Method: method}

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
func BasicAuthSecurity(name string, dsl ...func()) {
	if _, ok := dslengine.CurrentDefinition().(*design.APIDefinition); !ok {
		dslengine.IncompatibleDSL()
		return
	}

	if securityMethodRedefined(name) {
		return
	}

	def := &design.SecurityMethodDefinition{
		Kind:   design.BasicAuthSecurityKind,
		Method: name,
		Type:   "basic",
	}

	if len(dsl) != 0 {
		if !dslengine.Execute(dsl[0], def) {
			return
		}
	}

	design.Design.SecurityMethods = append(design.Design.SecurityMethods, def)
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
//    		InHeader("Authorization")
//     })
//
func APIKeySecurity(name string, dsl ...func()) {
	if _, ok := dslengine.CurrentDefinition().(*design.APIDefinition); !ok {
		dslengine.IncompatibleDSL()
		return
	}

	if securityMethodRedefined(name) {
		return
	}

	def := &design.SecurityMethodDefinition{
		Kind:   design.APIKeySecurityKind,
		Method: name,
		Type:   "apiKey",
	}

	if len(dsl) != 0 {
		if !dslengine.Execute(dsl[0], def) {
			return
		}
	}

	design.Design.SecurityMethods = append(design.Design.SecurityMethods, def)
}

// OAuth2Security defines the different Security methods that are
// available throughout the API.
//
// Example:
//
//     OAuth2Security("googAuth", func() {
//	    	OAuth2Flow("accessCode")
//	    	AuthorizationURL("http://example.com/authorization")
//	    	TokenURL("http://example.com/token")
//	    	Scope("my_system:write", "Write to the system")
//	    	Scope("my_system:read", "Read anything in there")
//     })
//
func OAuth2Security(name string, dsl ...func()) {
	if _, ok := dslengine.CurrentDefinition().(*design.APIDefinition); !ok {
		dslengine.IncompatibleDSL()
		return
	}

	if securityMethodRedefined(name) {
		return
	}

	def := &design.SecurityMethodDefinition{
		Method: name,
		Kind:   design.OAuth2SecurityKind,
		Type:   "oauth2",
	}

	if len(dsl) != 0 {
		if !dslengine.Execute(dsl[0], def) {
			return
		}
	}

	design.Design.SecurityMethods = append(design.Design.SecurityMethods, def)
}

// OtherSecurity defines a free-form security method, going above and
// beyond the Swagger specs. If you want to benefit from Swagger UIs,
// make sure the "securityType" is one of "apiKey", "oauth2" or
// "basic".
//
// Example:
//
//     OtherSecurity("jwt", "apiKey", func() {
//	    	InHeader("Authorization")
//	    	AuthorizationURL("http://example.com/authorization")
//	    	Scope("my_system:write", "Write to the system")
//	    	Scope("my_system:read", "Read anything in there")
//     })
//
func OtherSecurity(name string, securityType string, dsl ...func()) {
	if _, ok := dslengine.CurrentDefinition().(*design.APIDefinition); !ok {
		dslengine.IncompatibleDSL()
		return
	}

	if securityMethodRedefined(name) {
		return
	}

	def := &design.SecurityMethodDefinition{
		Kind:   design.OtherSecurityKind,
		Method: name,
		Type:   securityType,
	}

	if len(dsl) != 0 {
		if !dslengine.Execute(dsl[0], def) {
			return
		}
	}

	design.Design.SecurityMethods = append(design.Design.SecurityMethods, def)
}

// Scope defines an authorization scope. Used within SecurityMethod,
// the description is required, explaining what the scope
// means. Within a Security block, only a scope is needed.
func Scope(name string, desc ...string) {
	switch parent := dslengine.CurrentDefinition().(type) {
	case *design.SecurityDefinition:
		if len(desc) == 1 {
			dslengine.ReportError("must not pass description to Scope()")
			return
		}
		parent.Scopes = append(parent.Scopes, name)
	case *design.SecurityMethodDefinition:
		if len(desc) == 0 {
			dslengine.ReportError("no description provided for Scope()")
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

// InHeader defines both `in` and `name` of a SecurityMethod.  See
// http://swagger.io/specification/#securitySchemeObject for details.
func InHeader(headerName string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
		if parent.Kind == design.APIKeySecurityKind || parent.Kind == design.OtherSecurityKind {
			if parent.In != "" {
				dslengine.ReportError("'In' previously defined through InHeader or InQuery")
				return
			}
			parent.In = "header"
			parent.Name = headerName
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// InQuery defines both `in` and `name` of a SecurityMethod. See
// http://swagger.io/specification/#securitySchemeObject for details.
func InQuery(parameterName string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
		if parent.Kind == design.APIKeySecurityKind || parent.Kind == design.OtherSecurityKind {
			if parent.In != "" {
				dslengine.ReportError("'In' previously defined through InHeader or InQuery")
				return
			}
			parent.In = "query"
			parent.Name = parameterName
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// OAuth2Flow defines the `flow` for a SecurityMethod.  See
// http://swagger.io/specification/#securitySchemeObject for details.
func OAuth2Flow(flow string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind || parent.Kind == design.OtherSecurityKind {
			parent.Flow = flow
			return
		}
	}
	dslengine.IncompatibleDSL()
}

// TokenURL defines the `tokenUrl` for a SecurityMethod.  See
// http://swagger.io/specification/#securitySchemeObject for details.
func TokenURL(tokenURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind || parent.Kind == design.OtherSecurityKind {
			parent.TokenURL = tokenURL
			return
		}
	}

	// TODO: we'll need to check consistency of Flow/AuthorizationURL/TargetURL in
	// a later validation step.. probablyin `Validate()` then.
	dslengine.IncompatibleDSL()
}

// AuthorizationURL defines the `authorizationUrl` for a SecurityMethod.  See
// http://swagger.io/specification/#securitySchemeObject for details.
func AuthorizationURL(authorizationURL string) {
	if parent, ok := dslengine.CurrentDefinition().(*design.SecurityMethodDefinition); ok {
		if parent.Kind == design.OAuth2SecurityKind || parent.Kind == design.OtherSecurityKind {
			parent.AuthorizationURL = authorizationURL
			return
		}
	}
	dslengine.IncompatibleDSL()
}
