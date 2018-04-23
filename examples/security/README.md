# goa v2 Security Example

This example illustrates how to secure microservice endpoints. The service
endpoints showcase the various security schemes supported in goa. It exposes
endpoints secured via different security requirements, the `doubly_secure` and
`also_doubly_secure` endpoints illustrate how to secure a single endpoint using
multiple requirements.

## Design

The key design sections for the `multi_auth` service define the various security
requirements. The most interesting ones are the `doubly_secure` and
`also_doubly_secure` requirements:

```go
Security(JWTAuth, APIKeyAuth, func() { // Use JWT and an API key to secure this endpoint.
	Scope("api:read")  // Enforce presence of both "api:read"
	Scope("api:write") // and "api:write" scopes in JWT claims.
})
```

The payload DSL defines two attributes `key` and `token` that hold the API key
and JWT token respectively:

```go
Payload(func() {
	APIKey("api_key", "key", String, func() {
		Description("API key")
	})
	Token("token", String, func() {
		Description("JWT used for authentication")
	})
})
```
The design requires the client to provide both an API key and a JWT token.
`doubly_secure` loads the value of the API key from the request query string
while `also_doubly_secure` loads it from the request headers.

`doubly_secure`

```go
HTTP(func() {
	GET("/secure")

	Param("key:k")
          ...
```

`also_doubly_secure`

```go
HTTP(func() {
	POST("/secure")

	Header("key:Authorization")
```

