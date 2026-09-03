package testdata

import . "goa.design/goa/v3/dsl"

var BasicAuth = BasicAuthSecurity("basic")

var JWTAuth = JWTSecurity("jwt", func() {
	Scope("api:read", "Read-only access")
	Scope("api:write", "Read and write access")
	Scope("api:admin", "Admin access")
})

var APIKeyAuth = APIKeySecurity("api_key")

var BearerAuth = BearerSecurity("bearer", func() {
	Scope("api:read", "Read-only access")
})

var OAuth2 = OAuth2Security("authCode", func() {
	AuthorizationCodeFlow("http://^authorization", "^example:/token<>", "http://refresh^") // invalid URLs
	Scope("api:write", "Write acess")
	Scope("api:read", "Read access")
})

var ValidSecuritySchemesExtendDSL = func() {
	var CommonAttr = Type("Common", func() {
		Attribute("version", String)
	})
	var SecurityAttr = Type("Security", func() {
		Username("user", String)
		Password("pass", String)
	})
	Service("ValidSecuritySchemesExtendService", func() {
		Method("SecureMethod", func() {
			Security(BasicAuth)
			Payload(func() {
				Extend(CommonAttr)
				Extend(SecurityAttr)
			})
		})
	})
}

var InvalidSecuritySchemesDSL = func() {
	Service("InvalidSecuritySchemesService", func() {
		Security(OAuth2, APIKeyAuth, func() {
			Scope("not:found") // invalid security scope
		})
		Method("SecureMethod", func() {
			Security(BasicAuth, BearerAuth, JWTAuth, func() {
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
	Service("AnotherInvalidSecuritySchemesService", func() {
		Method("Method", func() {
			Payload(func() {
				Username("user", String)
				Password("pass", String)
				APIKey("key_key", "key", String)
				BearerToken("bearer_token", String)
				Token("token", String)
				AccessToken("access_token", String)
			})
			// invalid: missing security scheme
		})
	})
}

var InvalidSecurityFieldTypeDSL = func() {
	Service("InvalidSecurityFieldType", func() {
		Method("Authenticate", func() {
			Security(BasicAuth)
			Payload(func() {
				Username("username", String, func() {
					Meta("struct:field:type", "CustomUsername")
				})
				Password("password", String)
			})
		})
	})
}

var InvalidSecurityFieldDataTypesDSL = func() {
	basic := BasicAuthSecurity("basic")
	apiKey := APIKeySecurity("api_key")
	bearer := BearerSecurity("bearer")
	jwt := JWTSecurity("jwt")
	oauth := OAuth2Security("oauth")

	Service("InvalidSecurityFieldDataTypes", func() {
		Method("Authenticate", func() {
			Security(basic, apiKey, bearer, jwt, oauth)
			Payload(func() {
				Username("username", Int)
				Password("password", Boolean)
				APIKey("api_key", "api_key", Bytes)
				BearerToken("bearer", ArrayOf(String))
				Token("jwt", MapOf(String, String))
				AccessToken("oauth", func() {
					Attribute("value", String)
				})
			})
		})
	})
}

var InvalidSecurityFieldDefaultDSL = func() {
	jwt := JWTSecurity("jwt")

	Service("InvalidSecurityFieldDefault", func() {
		Method("Authenticate", func() {
			Security(jwt)
			Payload(func() {
				Token("token", String, func() {
					Default("generated credential")
				})
			})
		})
	})
}

var ValidNamedSecurityFieldTypesDSL = func() {
	credential := Type("Credential", String)
	basic := BasicAuthSecurity("basic")
	apiKey := APIKeySecurity("api_key")
	bearer := BearerSecurity("bearer")
	jwt := JWTSecurity("jwt")
	oauth := OAuth2Security("oauth")

	Service("ValidNamedSecurityFieldTypes", func() {
		Method("Authenticate", func() {
			Security(basic, apiKey, bearer, jwt, oauth)
			Payload(func() {
				Username("username", credential)
				Password("password", credential)
				APIKey("api_key", "api_key", credential)
				BearerToken("bearer", credential)
				Token("jwt", credential)
				AccessToken("oauth", credential)
			})
		})
	})
}
