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
	Service("AnotherInvalidSecuritySchemesService", func() {
		Method("Method", func() {
			Payload(func() {
				Username("user", String)
				Password("pass", String)
				APIKey("key_key", "key", String)
				Token("token", String)
				AccessToken("access_token", String)
			})
			// invalid: missing security scheme
		})
	})
}

var UnionBranchSecurityAttributeDSL = func() {
	TokenPayload := Type("UnionBranchSecurityTokenPayload", func() {
		Token("token", String)
		Attribute("message", String)
	})
	JSONPayload := Type("UnionBranchSecurityJSONPayload", func() {
		Attribute("message", String)
	})
	Service("UnionBranchSecurityService", func() {
		Method("SecureMethod", func() {
			Security(JWTAuth)
			Payload(OneOf(TokenPayload, JSONPayload))
		})
	})
}

var ConstructorUnionResultViewDSL = func() {
	TinyA := ResultType("application/vnd.a", func() {
		TypeName("TinyA")
		Attributes(func() {
			Attribute("a", String)
		})
		View("default", func() {
			Attribute("a")
		})
		View("tiny", func() {
			Attribute("a")
		})
	})
	TinyB := ResultType("application/vnd.b", func() {
		TypeName("TinyB")
		Attributes(func() {
			Attribute("b", String)
		})
		View("default", func() {
			Attribute("b")
		})
	})
	Service("ConstructorUnionResultViewService", func() {
		Method("Show", func() {
			Result(OneOf(TinyA, TinyB), func() {
				View("tiny")
			})
		})
	})
}
