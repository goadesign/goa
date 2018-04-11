package design

import (
	. "goa.design/goa/http/design"
	. "goa.design/plugins/security/dsl"
)

var _ = API("multi_auth", func() {
	Title("Security Example API")
	Description("This API demonstrates the use of the goa security DSL")
	Docs(func() { // Documentation links
		Description("Security example README")
		URL("https://github.com/goadesign/plugins/security/tree/master/example/multi_auth/README.md")
	})
})

// JWTAuth defines a security scheme that uses JWT tokens.
var JWTAuth = JWTSecurity("jwt", func() {
	Description(`Secures endpoint by requiring a valid JWT token retrieved via the signin endpoint. Supports scopes "api:read" and "api:write".`)
	Scope("api:read", "Read-only access")
	Scope("api:write", "Read and write access")
})

// APIKeyAuth defines a security scheme that uses API keys.
var APIKeyAuth = APIKeySecurity("api_key", func() {
	Description("Secures endpoint by requiring an API key.")
})

// BasicAuth defines a security scheme using basic authentication. The scheme
// protects the "signin" action used to create JWTs.
var BasicAuth = BasicAuthSecurity("basic", func() {
	Description("Basic authentication used to authenticate security principal during signin")
})

// OAuth2Auth defines a security scheme that uses OAuth2 tokens.
var OAuth2Auth = OAuth2Security("oauth2", func() {
	AuthorizationCodeFlow("/authorization", "/token", "/refresh")
	Description(`Secures endpoint by requiring a valid OAuth2 token retrieved via the signin endpoint. Supports scopes "api:read" and "api:write".`)
	Scope("api:read", "Read-only access")
	Scope("api:write", "Read and write access")
})

var _ = Service("secured_service", func() {
	Description("The secured service exposes endpoints that require valid authorization credentials.")

	Error("unauthorized", String, "Credentials are invalid")
	HTTP(func() {
		Response("unauthorized", StatusUnauthorized)
	})

	Method("signin", func() {
		Description("Creates a valid JWT")

		// The signin endpoint is secured via basic auth
		Security(BasicAuth)

		Payload(func() {
			Description("Credentials used to authenticate to retrieve JWT token")
			Username("username", String, "Username used to perform signin", func() {
				Example("user")
			})
			Password("password", String, "Password used to perform signin", func() {
				Example("password")
			})
		})

		HTTP(func() {
			POST("/signin")
			// Use Authorization header to provide basic auth value.
			Response(StatusNoContent)
		})
	})

	Method("secure", func() {
		Description("This action is secured with the jwt scheme")
		Security(JWTAuth, func() { // Use JWT to auth requests to this endpoint.
			Scope("api:read") // Enforce presence of "api:read" scope in JWT claims.
		})
		Payload(func() {
			Attribute("fail", Boolean, func() {
				Description("Whether to force auth failure even with a valid JWT")
			})
			Token("token", String, func() {
				Description("JWT used for authentication")
			})
		})
		Result(String, func() {
			Example("JWT secured data")
		})
		HTTP(func() {
			GET("/secure")
			Param("fail")
			Response(StatusOK)
		})
	})

	Method("doubly_secure", func() {
		Description("This action is secured with the jwt scheme and also requires an API key query string.")
		Security(JWTAuth, APIKeyAuth, func() { // Use JWT and an API key to secure this endpoint.
			Scope("api:read")  // Enforce presence of both "api:read"
			Scope("api:write") // and "api:write" scopes in JWT claims.
		})
		Payload(func() {
			APIKey("api_key", "key", String, func() {
				Description("API key")
				Example("abcdef12345")
			})
			Token("token", String, func() {
				Description("JWT used for authentication")
				Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
			})
		})
		Result(String, func() {
			Example("JWT secured data")
		})
		HTTP(func() {
			PUT("/secure")
			Param("key:k")
			Response(StatusOK)
		})
	})

	Method("also_doubly_secure", func() {
		Description("This action is secured with the jwt scheme and also requires an API key header.")
		Security(JWTAuth, APIKeyAuth, func() { // Use JWT and an API key to secure this endpoint.
			Scope("api:read")  // Enforce presence of both "api:read"
			Scope("api:write") // and "api:write" scopes in JWT claims.
		})
		Security(OAuth2Auth, BasicAuth, func() {
			Scope("api:read")  // Enforce presence of both "api:read"
			Scope("api:write") // and "api:write" scopes in OAuth2 claims.
		})
		Payload(func() {
			Username("username", String, "Username used to perform signin", func() {
				Example("user")
			})
			Password("password", String, "Password used to perform signin", func() {
				Example("password")
			})
			APIKey("api_key", "key", String, func() {
				Description("API key")
				Example("abcdef12345")
			})
			Token("token", String, func() {
				Description("JWT used for authentication")
				Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
			})
			AccessToken("oauth_token", String)
		})
		Result(String, func() {
			Example("JWT secured data")
		})
		HTTP(func() {
			POST("/secure")
			Header("token:Authorization")
			Param("key:k")
			Param("oauth_token:oauth")
			Response(StatusOK)
		})
	})
})
