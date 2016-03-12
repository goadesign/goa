package apidsl_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security", func() {
	BeforeEach(func() {
		dslengine.Reset()
	})

	It("should have no security DSL when none are defined", func() {
		API("secure", nil)
		dslengine.Run()
		Ω(Design.SecurityMethods).Should(BeNil())
		Ω(dslengine.Errors).ShouldNot(HaveOccurred())
	})

	It("should be the fully valid and well defined, live on the happy path", func() {
		API("secure", func() {
			BasicAuthSecurity("basic_authz", func() {
				Description("desc")
			})

			OAuth2Security("googAuthz", func() {
				Description("desc")
				AuthorizationURL("http://example.com/auth")
				TokenURL("http://example.com/token")
			})

			APIKeySecurity("jwt", func() {
				Description("desc")
				InHeader("Authorization")
			})

			APIKeySecurity("a_key", func() {
				Description("desc")
				InQuery("access_token")
			})

			OtherSecurity("custom", "apiKey", func() {
				Description("desc")
				OAuth2Flow("whatever")
				AuthorizationURL("http://example.com/auth")
				TokenURL("http://example.com/token")
			})
		})

		dslengine.Run()

		Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		Ω(Design.SecurityMethods).Should(HaveLen(5))

		Ω(Design.SecurityMethods[0].Kind).Should(Equal(BasicAuthSecurityKind))
		Ω(Design.SecurityMethods[0].Description).Should(Equal("desc"))

		Ω(Design.SecurityMethods[1].Kind).Should(Equal(OAuth2SecurityKind))
		Ω(Design.SecurityMethods[1].AuthorizationURL).Should(Equal("http://example.com/auth"))
		Ω(Design.SecurityMethods[1].TokenURL).Should(Equal("http://example.com/token"))

		Ω(Design.SecurityMethods[2].Kind).Should(Equal(APIKeySecurityKind))
		Ω(Design.SecurityMethods[2].In).Should(Equal("header"))
		Ω(Design.SecurityMethods[2].Name).Should(Equal("Authorization"))

		Ω(Design.SecurityMethods[3].Kind).Should(Equal(APIKeySecurityKind))
		Ω(Design.SecurityMethods[3].In).Should(Equal("query"))
		Ω(Design.SecurityMethods[3].Name).Should(Equal("access_token"))

		Ω(Design.SecurityMethods[4].Kind).Should(Equal(OtherSecurityKind))
		Ω(Design.SecurityMethods[4].Flow).Should(Equal("whatever"))
		Ω(Design.SecurityMethods[4].AuthorizationURL).Should(Equal("http://example.com/auth"))
		Ω(Design.SecurityMethods[4].TokenURL).Should(Equal("http://example.com/token"))
	})

	Context("with basic security", func() {
		It("should fail because of duplicate In declaration", func() {
			API("", func() {
				BasicAuthSecurity("broken_basic_authz", func() {
					Description("desc")
					InHeader("Authorization")
					InQuery("access_token")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})

		It("should fail because of invalid declaration of OAuth2Flow", func() {
			API("", func() {
				BasicAuthSecurity("broken_basic_authz", func() {
					Description("desc")
					OAuth2Flow("invalid")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})

		It("should fail because of invalid declaration of AuthorizationURL", func() {
			API("", func() {
				BasicAuthSecurity("broken_basic_authz", func() {
					Description("desc")
					AuthorizationURL("invalid")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})

		It("should fail because of invalid declaration of TokenURL", func() {
			API("", func() {
				BasicAuthSecurity("broken_basic_authz", func() {
					Description("desc")
					TokenURL("invalid")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})

		It("should fail because of invalid declaration of InHeader", func() {
			API("", func() {
				BasicAuthSecurity("broken_basic_authz", func() {
					Description("desc")
					InHeader("invalid")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})
	})

	Context("with oauth2 security", func() {
		It("should pass with valid values when well defined", func() {
			API("", func() {
				OAuth2Security("googAuthz", func() {
					Description("Use Goog's Auth")
					AuthorizationURL("https://example.com/auth")
					TokenURL("https://example.com/token")
					Scope("scope:1", "Desc 1")
					Scope("scope:2", "Desc 2")
				})
			})
			Resource("one", func() {
				Action("first", func() {
					Routing(GET("/first"))
					Security("googAuthz", func() {
						Scope("scope:1")
					})
				})
			})

			dslengine.Run()

			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(Design.SecurityMethods).Should(HaveLen(1))
			method := Design.SecurityMethods[0]
			Ω(method.Description).Should(Equal("Use Goog's Auth"))
			Ω(method.AuthorizationURL).Should(Equal("https://example.com/auth"))
			Ω(method.TokenURL).Should(Equal("https://example.com/token"))
			Ω(method.Scopes["scope:1"]).Should(Equal("Desc 1"))
			Ω(method.Scopes["scope:2"]).Should(Equal("Desc 2"))
		})

		It("should fail because of invalid declaration of InHeader", func() {
			API("", func() {
				OAuth2Security("googAuthz", func() {
					InHeader("invalid")
				})
			})
			dslengine.Run()
			Ω(dslengine.Errors).Should(HaveOccurred())
		})

	})

	Context("with resources and actions", func() {
		It("should fallback properly to lower-level security", func() {
			API("", func() {
				OtherSecurity("jwt", "apiKey")
				BasicAuthSecurity("password")

				Security("jwt")
			})
			Resource("one", func() {
				Action("first", func() {
					Routing(GET("/first"))
					NoSecurity()
				})
				Action("second", func() {
					Routing(GET("/second"))
				})
			})
			Resource("two", func() {
				Security("password")

				Action("third", func() {
					Routing(GET("/third"))
				})
				Action("fourth", func() {
					Routing(GET("/fourth"))
					Security("jwt")
				})
			})
			Resource("three", func() {
				Action("fifth", func() {
					Routing(GET("/fifth"))
				})
			})
			Resource("auth", func() {
				NoSecurity()

				Action("auth", func() {
					Routing(GET("/auth"))
				})
				Action("refresh", func() {
					Routing(GET("/refresh"))
					Security("jwt")
				})
			})

			dslengine.Run()

			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(Design.SecurityMethods).Should(HaveLen(2))
			Ω(Design.Resources["one"].Actions["first"].Security).Should(BeNil())
			Ω(Design.Resources["one"].Actions["second"].Security.Method).Should(Equal("jwt"))
			Ω(Design.Resources["two"].Actions["third"].Security.Method).Should(Equal("password"))
			Ω(Design.Resources["two"].Actions["fourth"].Security.Method).Should(Equal("jwt"))
			Ω(Design.Resources["three"].Actions["fifth"].Security.Method).Should(Equal("jwt"))
			Ω(Design.Resources["auth"].Actions["auth"].Security).Should(BeNil())
			Ω(Design.Resources["auth"].Actions["refresh"].Security.Method).Should(Equal("jwt"))
		})
	})

	// Test APIKEySecurity
	// Test OtherSecurity

	Context("with some defined API", func() {
		var name string
		var dsl func()

		BeforeEach(func() {
			dslengine.Reset()
			name = ""
			dsl = nil
		})

		JustBeforeEach(func() {
			API("secure", dsl)
			dslengine.Run()
		})

		// Add more context/it blocks here
	})

})
