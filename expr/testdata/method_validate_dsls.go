package testdata

import . "goa.design/goa/v3/dsl"

var BasicAuth = BasicAuthSecurity("basic")

var JWTAuth = JWTSecurity("jwt", func() {
	Scope("api:read", "Read-only access")
	Scope("api:write", "Read and write access")
	Scope("api:admin", "Admin access")
})

var APIKeyAuth = APIKeySecurity("api_key")

var OAuth2 = OAuth2Security("authCode", func() {
	AuthorizationCodeFlow("http://^authorization", "^example:/token<>", "http://refresh^") // invalid URLs
	Scope("api:write", "Write acess")
	Scope("api:read", "Read access")
})

var InvalidSecuritySchemesDSL = func() {
	Service("InvalidSecuritySchemesService", func() {
		Security(OAuth2, APIKeyAuth, func() {
			Scope("not:found") // invalid security scope
		})
		Method("SecureMethod", func() {
			Security(BasicAuth, JWTAuth, func() {
				Scope("not:found") // invalid security scope
			})
			Payload(func() {
				Attribute("a", String)
				// invalid: missing security attribute definitions
			})
		})
		Method("InheritedSecureMethod", func() {
			Payload(func() {
				Attribute("b", String)
				// invalid: missing security attribute definitions
			})
		})
	})
}
