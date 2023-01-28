package apidsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kyokomi/goa-v1/design"
	"github.com/kyokomi/goa-v1/design/apidsl"
	"github.com/kyokomi/goa-v1/dslengine"
)

var _ = Describe("Security", func() {
	BeforeEach(func() {
		dslengine.Reset()
	})

	It("should have no security DSL when none are defined", func() {
		apidsl.API("secure", nil)
		dslengine.Run()
		Ω(Design.SecuritySchemes).Should(BeNil())
		Ω(dslengine.Errors).ShouldNot(HaveOccurred())
	})

	It("should be the fully valid and well defined, live on the happy path", func() {
		apidsl.API("secure", func() {
			apidsl.Host("example.com")
			apidsl.Scheme("http")

			apidsl.BasicAuthSecurity("basic_authz", func() {
				apidsl.Description("desc")
			})

			apidsl.OAuth2Security("googAuthz", func() {
				apidsl.Description("desc")
				apidsl.AccessCodeFlow("/auth", "/token")
				apidsl.Scope("user:read", "Read users")
			})

			apidsl.APIKeySecurity("a_key", func() {
				apidsl.Description("desc")
				apidsl.Query("access_token")
			})

			apidsl.JWTSecurity("jwt", func() {
				apidsl.Description("desc")
				apidsl.Header("Authorization")
				apidsl.TokenURL("/token")
				apidsl.Scope("user:read", "Read users")
				apidsl.Scope("user:write", "Write users")
			})
		})

		dslengine.Run()

		Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		Ω(Design.SecuritySchemes).Should(HaveLen(4))

		Ω(Design.SecuritySchemes[0].Kind).Should(Equal(BasicAuthSecurityKind))
		Ω(Design.SecuritySchemes[0].Description).Should(Equal("desc"))

		Ω(Design.SecuritySchemes[1].Kind).Should(Equal(OAuth2SecurityKind))
		Ω(Design.SecuritySchemes[1].AuthorizationURL).Should(Equal("http://example.com/auth"))
		Ω(Design.SecuritySchemes[1].TokenURL).Should(Equal("http://example.com/token"))
		Ω(Design.SecuritySchemes[1].Flow).Should(Equal("accessCode"))

		Ω(Design.SecuritySchemes[2].Kind).Should(Equal(APIKeySecurityKind))
		Ω(Design.SecuritySchemes[2].In).Should(Equal("query"))
		Ω(Design.SecuritySchemes[2].Name).Should(Equal("access_token"))

		Ω(Design.SecuritySchemes[3].Kind).Should(Equal(JWTSecurityKind))
		Ω(Design.SecuritySchemes[3].TokenURL).Should(Equal("http://example.com/token"))
		Ω(Design.SecuritySchemes[3].Scopes).Should(HaveLen(2))
	})

	Context("with basic security", func() {
		It("should fail because of duplicate In declaration", func() {
			apidsl.API("", func() {
				apidsl.BasicAuthSecurity("broken_basic_authz", func() {
					apidsl.Description("desc")
					apidsl.Header("Authorization")
					apidsl.Query("access_token")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})

		It("should fail because of invalid declaration of OAuth2Flow", func() {
			apidsl.API("", func() {
				apidsl.BasicAuthSecurity("broken_basic_authz", func() {
					apidsl.Description("desc")
					apidsl.ImplicitFlow("invalid")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})

		It("should fail because of invalid declaration of TokenURL", func() {
			apidsl.API("", func() {
				apidsl.BasicAuthSecurity("broken_basic_authz", func() {
					apidsl.Description("desc")
					apidsl.TokenURL("/token")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})

		It("should fail because of invalid declaration of TokenURL", func() {
			apidsl.API("", func() {
				apidsl.BasicAuthSecurity("broken_basic_authz", func() {
					apidsl.Description("desc")
					apidsl.TokenURL("in valid")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})

		It("should fail because of invalid declaration of Header", func() {
			apidsl.API("", func() {
				apidsl.BasicAuthSecurity("broken_basic_authz", func() {
					apidsl.Description("desc")
					apidsl.Header("invalid")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})
	})

	Context("with oauth2 security", func() {
		It("should pass with valid values when well defined", func() {
			apidsl.API("", func() {
				apidsl.Host("example.com")
				apidsl.Scheme("http")
				apidsl.OAuth2Security("googAuthz", func() {
					apidsl.Description("Use Goog's Auth")
					apidsl.AccessCodeFlow("/auth", "/token")
					apidsl.Scope("scope:1", "Desc 1")
					apidsl.Scope("scope:2", "Desc 2")
				})
			})
			apidsl.Resource("one", func() {
				apidsl.Action("first", func() {
					apidsl.Routing(apidsl.GET("/first"))
					apidsl.Security("googAuthz", func() {
						apidsl.Scope("scope:1")
					})
				})
			})

			dslengine.Run()

			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(Design.SecuritySchemes).Should(HaveLen(1))
			scheme := Design.SecuritySchemes[0]
			Ω(scheme.Description).Should(Equal("Use Goog's Auth"))
			Ω(scheme.AuthorizationURL).Should(Equal("http://example.com/auth"))
			Ω(scheme.TokenURL).Should(Equal("http://example.com/token"))
			Ω(scheme.Flow).Should(Equal("accessCode"))
			Ω(scheme.Scopes["scope:1"]).Should(Equal("Desc 1"))
			Ω(scheme.Scopes["scope:2"]).Should(Equal("Desc 2"))
		})

		It("should fail because of invalid declaration of Header", func() {
			apidsl.API("", func() {
				apidsl.OAuth2Security("googAuthz", func() {
					apidsl.Header("invalid")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})

	})

	Context("with resources and actions", func() {
		It("should fallback properly to lower-level security", func() {
			apidsl.API("", func() {
				apidsl.JWTSecurity("jwt", func() {
					apidsl.TokenURL("/token")
					apidsl.Scope("read", "Read")
					apidsl.Scope("write", "Write")
				})
				apidsl.BasicAuthSecurity("password")

				apidsl.Security("jwt")
			})
			apidsl.Resource("one", func() {
				apidsl.Action("first", func() {
					apidsl.Routing(apidsl.GET("/first"))
					apidsl.NoSecurity()
				})
				apidsl.Action("second", func() {
					apidsl.Routing(apidsl.GET("/second"))
				})
			})
			apidsl.Resource("two", func() {
				apidsl.Security("password")

				apidsl.Action("third", func() {
					apidsl.Routing(apidsl.GET("/third"))
				})
				apidsl.Action("fourth", func() {
					apidsl.Routing(apidsl.GET("/fourth"))
					apidsl.Security("jwt")
				})
			})
			apidsl.Resource("three", func() {
				apidsl.Action("fifth", func() {
					apidsl.Routing(apidsl.GET("/fifth"))
				})
			})
			apidsl.Resource("auth", func() {
				apidsl.NoSecurity()

				apidsl.Action("auth", func() {
					apidsl.Routing(apidsl.GET("/auth"))
				})
				apidsl.Action("refresh", func() {
					apidsl.Routing(apidsl.GET("/refresh"))
					apidsl.Security("jwt")
				})
			})

			dslengine.Run()

			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(Design.SecuritySchemes).Should(HaveLen(2))
			Ω(Design.Resources["one"].Actions["first"].Security).Should(BeNil())
			Ω(Design.Resources["one"].Actions["second"].Security.Scheme.SchemeName).Should(Equal("jwt"))
			Ω(Design.Resources["two"].Actions["third"].Security.Scheme.SchemeName).Should(Equal("password"))
			Ω(Design.Resources["two"].Actions["fourth"].Security.Scheme.SchemeName).Should(Equal("jwt"))
			Ω(Design.Resources["three"].Actions["fifth"].Security.Scheme.SchemeName).Should(Equal("jwt"))
			Ω(Design.Resources["auth"].Actions["auth"].Security).Should(BeNil())
			Ω(Design.Resources["auth"].Actions["refresh"].Security.Scheme.SchemeName).Should(Equal("jwt"))
		})
	})
})
