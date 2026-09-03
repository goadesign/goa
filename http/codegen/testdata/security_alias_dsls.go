// This file defines HTTP services whose security credentials use named string
// types. The generated server must keep those types while HTTP carries strings.
package testdata

import . "goa.design/goa/v3/dsl"

// NamedSecurityTypesDSL exercises every security scheme with named string
// credentials and both required and optional payload fields.
var NamedSecurityTypesDSL = func() {
	credential := Type("Credential", String)
	basic := BasicAuthSecurity("basic")
	apiKey := APIKeySecurity("api_key")
	bearer := BearerSecurity("bearer")
	jwt := JWTSecurity("jwt")
	oauth := OAuth2Security("oauth")

	Service("NamedSecurityTypes", func() {
		Method("Basic", func() {
			Security(basic)
			Payload(func() {
				Username("username", credential)
				Password("password", credential)
				Required("username")
			})
			HTTP(func() {
				GET("/basic")
			})
		})
		Method("BasicOptional", func() {
			Security(basic)
			Payload(func() {
				Username("username", credential)
				Password("password", credential)
			})
			HTTP(func() {
				GET("/basic-optional")
			})
		})
		Method("BasicPasswordRequired", func() {
			Security(basic)
			Payload(func() {
				Username("username", credential)
				Password("password", credential)
				Required("password")
			})
			HTTP(func() {
				GET("/basic-password-required")
			})
		})
		Method("APIKey", func() {
			Security(apiKey)
			Payload(func() {
				APIKey("api_key", "key", credential)
				Required("key")
			})
			HTTP(func() {
				GET("/api-key")
				Header("key:X-API-Key")
			})
		})
		Method("Bearer", func() {
			Security(bearer)
			Payload(func() {
				BearerToken("token", credential)
			})
			HTTP(func() {
				GET("/bearer")
				Header("token:X-Bearer")
			})
		})
		Method("JWT", func() {
			Security(jwt)
			Payload(func() {
				Token("token", credential)
				Required("token")
			})
			HTTP(func() {
				GET("/jwt")
				Header("token:X-JWT")
			})
		})
		Method("OAuth2", func() {
			Security(oauth)
			Payload(func() {
				AccessToken("token", credential)
			})
			HTTP(func() {
				GET("/oauth")
				Header("token:X-OAuth")
			})
		})
	})
}
