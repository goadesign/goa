// These tests compile service endpoints and native transports together. They
// verify exact callback errors, credentials, scopes and contexts. Dispatch counts
// show when authentication fails, while native transport responses stay unchanged.
package generator

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

func TestGeneratedAuthenticationRejection(t *testing.T) {
	registry := testRegistry("gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
	)
	codegen.RunDSL(t, authenticationRuntimeDSL)
	directory := t.TempDir()
	generated := filepath.Join(directory, codegen.Gendir)
	writeGeneratedModule(t, generated, "generated.local/gen")
	_, err := generate(directory, "gen", false, registry)
	require.NoError(t, err)
	writeGeneratedTest(t, filepath.Join(generated, "authentication_test.go"), authenticationRuntimeTest)
	runGeneratedTests(t, generated)
}

// authenticationRuntimeDSL gives each scheme its own method, then exercises
// combined requirements and alternatives with their applicable credentials.
func authenticationRuntimeDSL() {
	dsl.API("authentication rejection", func() {})
	basic := dsl.BasicAuthSecurity("basic", func() {
		dsl.Scope("read")
	})
	key := dsl.APIKeySecurity("key", func() {
		dsl.Scope("read")
	})
	bearer := dsl.BearerSecurity("bearer", func() {
		dsl.Scope("read")
	})
	jwt := dsl.JWTSecurity("jwt", func() {
		dsl.Scope("read")
		dsl.Scope("write")
	})
	oauth := dsl.OAuth2Security("oauth", func() {
		dsl.AuthorizationCodeFlow("https://issuer.example/authorize", "https://issuer.example/token", "https://issuer.example/token")
		dsl.Scope("read")
	})
	dsl.Service("auth_probe", func() {
		for _, name := range []string{"basic", "key", "bearer", "jwt", "oauth", "combined", "alternative", "open"} {
			dsl.Method(name, func() {
				switch name {
				case "basic":
					dsl.Security(basic, func() {
						dsl.Scope("read")
					})
				case "key":
					dsl.Security(key, func() {
						dsl.Scope("read")
					})
				case "bearer":
					dsl.Security(bearer, func() {
						dsl.Scope("read")
					})
				case "jwt":
					dsl.Security(jwt, func() {
						dsl.Scope("read")
					})
				case "oauth":
					dsl.Security(oauth, func() {
						dsl.Scope("read")
					})
				case "combined":
					dsl.Security(basic, jwt, func() {
						dsl.Scope("read")
					})
				case "alternative":
					dsl.Security(basic, func() {
						dsl.Scope("read")
					})
					dsl.Security(jwt, func() {
						dsl.Scope("write")
					})
				case "open":
					dsl.NoSecurity()
				}
				if name != "open" {
					dsl.Payload(func() {
						switch name {
						case "basic", "combined", "alternative":
							dsl.Username("user", dsl.String)
							dsl.Password("pass", dsl.String)
							dsl.Required("user", "pass")
						case "key":
							dsl.APIKey("key", "key", dsl.String)
							dsl.Required("key")
						case "bearer":
							dsl.BearerToken("token", dsl.String)
							dsl.Required("token")
						case "oauth":
							dsl.AccessToken("token", dsl.String)
							dsl.Required("token")
						}
						if name == "jwt" || name == "combined" || name == "alternative" {
							dsl.Token("token", dsl.String)
							dsl.Required("token")
						}
					})
				}
				dsl.Result(dsl.String)
				dsl.Error("denied")
				dsl.HTTP(func() {
					dsl.POST("/" + name)
					dsl.Response("denied", dsl.StatusUnauthorized)
				})
			})
		}
	})
	dsl.Service("auth_rpc", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("run", func() {
			dsl.Security(oauth, func() {
				dsl.Scope("read")
			})
			dsl.Payload(func() {
				dsl.AccessToken("access", dsl.String)
				dsl.Required("access")
			})
			dsl.Result(dsl.String)
			dsl.Error("denied")
			dsl.JSONRPC(func() {
				dsl.Response("denied", func() {
					dsl.Code(-32043)
				})
			})
		})
	})
}
