package design

import (
	. "goa.design/goa/dsl"
)

var _ = API("multi_auth", func() {
	Title("Security Example API")
	Description("This API demonstrates the use of the goa security DSL")
	Docs(func() { // Documentation links
		Description("Security example README")
		URL("https://github.com/goadesign/goa/tree/master/example/security/README.md")
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

	GRPC(func() {
		Response("unauthorized", CodeUnauthenticated)
	})

	Method("signin", func() {
		Description("Creates a valid JWT")

		// The signin endpoint is secured via basic auth
		Security(BasicAuth)

		Payload(func() {
			Description("Credentials used to authenticate to retrieve JWT token")
			UsernameField(1, "username", String, "Username used to perform signin", func() {
				Example("user")
			})
			PasswordField(2, "password", String, "Password used to perform signin", func() {
				Example("password")
			})
			Required("username", "password")
		})

		Result(Creds)

		HTTP(func() {
			POST("/signin")
			// Use Authorization header to provide basic auth value.
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("secure", func() {
		Description("This action is secured with the jwt scheme")

		Security(JWTAuth, func() { // Use JWT to auth requests to this endpoint.
			Scope("api:read") // Enforce presence of "api:read" scope in JWT claims.
		})

		Payload(func() {
			Field(1, "fail", Boolean, func() {
				Description("Whether to force auth failure even with a valid JWT")
			})
			TokenField(2, "token", String, func() {
				Description("JWT used for authentication")
			})
			Required("token")
		})

		Result(String)

		Error("invalid-scopes", String, "Token scopes are invalid")

		HTTP(func() {
			GET("/secure")
			Param("fail")
			Response(StatusOK)
			Response("invalid-scopes", StatusForbidden)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("invalid-scopes", CodeUnauthenticated)
		})
	})

	Method("doubly_secure", func() {
		Description("This action is secured with the jwt scheme and also requires an API key query string.")

		Security(JWTAuth, APIKeyAuth, func() { // Use JWT and an API key to secure this endpoint.
			Scope("api:read")  // Enforce presence of both "api:read"
			Scope("api:write") // and "api:write" scopes in JWT claims.
		})

		Payload(func() {
			APIKeyField(1, "api_key", "key", String, func() {
				Description("API key")
				Example("abcdef12345")
			})
			TokenField(2, "token", String, func() {
				Description("JWT used for authentication")
				Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
			})
			Required("key", "token")
		})

		Result(String)

		Error("invalid-scopes", String, "Token scopes are invalid")

		HTTP(func() {
			PUT("/secure")
			Param("key:k") // API key "key" sent in query parameter "k"
			Response(StatusOK)
			Response("invalid-scopes", StatusForbidden)
		})

		GRPC(func() {
			Message(func() {
				Attribute("key") // API key "key" sent in request message
			})
			Response(CodeOK)
			Response("invalid-scopes", CodeUnauthenticated)
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
			UsernameField(1, "username", String, "Username used to perform signin", func() {
				Example("user")
			})
			PasswordField(2, "password", String, "Password used to perform signin", func() {
				Example("password")
			})
			APIKeyField(3, "api_key", "key", String, func() {
				Description("API key")
				Example("abcdef12345")
			})
			TokenField(4, "token", String, func() {
				Description("JWT used for authentication")
				Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
			})
			AccessTokenField(5, "oauth_token", String)
		})

		Result(String)

		Error("invalid-scopes", String, "Token scopes are invalid")

		HTTP(func() {
			POST("/secure")
			Header("token:Authorization") // JWT token passed in "Authorization" header
			Param("key:k")                // API key "key" sent in query parameter "k"
			Param("oauth_token:oauth")    // OAuth token sent in query parameter "oauth"
			Response(StatusOK)
			Response("invalid-scopes", StatusForbidden)
		})

		GRPC(func() {
			Message(func() {
				Attribute("username") // "username" sent in request message
				Attribute("password") // "password" sent in request message
				Attribute("key")      // API key "key" sent in request message
			})
			Metadata(func() {
				Attribute("oauth_token:oauth") // OAuth token sent in request metadata key "oauth"
			})
			Response(CodeOK)
			Response("invalid-scopes", CodeUnauthenticated)
		})
	})
})

// Creds defines the credentials to use for authenticating to service methods.
var Creds = Type("Creds", func() {
	Field(1, "jwt", String, "JWT token", func() {
		Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
	})
	Field(2, "api_key", String, "API Key", func() {
		Example("abcdef12345")
	})
	Field(3, "oauth_token", String, "OAuth2 token", func() {
		Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
	})
	Required("jwt", "api_key", "oauth_token")
})
