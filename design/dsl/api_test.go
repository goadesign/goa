package dsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Describe("API", func() {
	var name string
	var dsl func()

	BeforeEach(func() {
		Design = nil
		Errors = nil
		name = ""
		dsl = nil
	})

	JustBeforeEach(func() {
		API(name, dsl)
		RunDSL()
	})

	Context("with no DSL", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces a valid API definition", func() {
			Ω(Design.Validate()).ShouldNot(HaveOccurred())
			Ω(Design.Name).Should(Equal(name))
		})
	})

	Context("with an already defined API with the same name", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces an error", func() {
			API(name, dsl)
			Ω(Errors).Should(HaveOccurred())
		})
	})

	Context("with an already defined API with a different name", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("returns an error", func() {
			API("news", dsl)
			Ω(Errors).Should(HaveOccurred())
		})
	})

	Context("with valid DSL", func() {
		JustBeforeEach(func() {
			Ω(Errors).ShouldNot(HaveOccurred())
			Ω(Design.Validate()).ShouldNot(HaveOccurred())
		})

		Context("with a description", func() {
			const description = "description"

			BeforeEach(func() {
				dsl = func() {
					Description(description)
				}
			})

			It("sets the API description", func() {
				Ω(Design.Description).Should(Equal(description))
			})
		})

		Context("with a title", func() {
			const title = "title"

			BeforeEach(func() {
				dsl = func() {
					Title(title)
				}
			})

			It("sets the API title", func() {
				Ω(Design.Title).Should(Equal(title))
			})
		})

		Context("with a terms of service", func() {
			const terms = "terms"

			BeforeEach(func() {
				dsl = func() {
					TermsOfService(terms)
				}
			})

			It("sets the API terms of service", func() {
				Ω(Design.TermsOfService).Should(Equal(terms))
			})
		})

		Context("with contact information", func() {
			const contactName = "contactName"
			const contactEmail = "contactEmail"
			const contactURL = "http://contactURL.com"

			BeforeEach(func() {
				dsl = func() {
					Contact(func() {
						Name(contactName)
						Email(contactEmail)
						URL(contactURL)
					})
				}
			})

			It("sets the contact information", func() {
				Ω(Design.Contact).Should(Equal(&ContactDefinition{
					Name:  contactName,
					Email: contactEmail,
					URL:   contactURL,
				}))
			})
		})

		Context("with license information", func() {
			const licenseName = "licenseName"
			const licenseURL = "http://licenseURL.com"

			BeforeEach(func() {
				dsl = func() {
					License(func() {
						Name(licenseName)
						URL(licenseURL)
					})
				}
			})

			It("sets the API license information", func() {
				Ω(Design.License).Should(Equal(&LicenseDefinition{
					Name: licenseName,
					URL:  licenseURL,
				}))
			})
		})

		Context("with a BasePath", func() {
			const basePath = "basePath"

			BeforeEach(func() {
				dsl = func() {
					BasePath(basePath)
				}
			})

			It("sets the API base path", func() {
				Ω(Design.BasePath).Should(Equal(basePath))
			})
		})

		Context("with BaseParams", func() {
			const param1Name = "accountID"
			const param1Type = Integer
			const param1Desc = "the account ID"
			const param2Name = "id"
			const param2Type = String
			const param2Desc = "the widget ID"

			BeforeEach(func() {
				dsl = func() {
					BaseParams(func() {
						Param(param1Name, param1Type, param1Desc)
						Param(param2Name, param2Type, param2Desc)
					})
				}
			})

			It("sets the API base parameters", func() {
				Ω(Design.BaseParams.Type).Should(BeAssignableToTypeOf(Object{}))
				params := Design.BaseParams.Type.ToObject()
				Ω(params).Should(HaveLen(2))
				Ω(params).Should(HaveKey(param1Name))
				Ω(params).Should(HaveKey(param2Name))
				Ω(params[param1Name].Type).Should(Equal(param1Type))
				Ω(params[param2Name].Type).Should(Equal(param2Type))
				Ω(params[param1Name].Description).Should(Equal(param1Desc))
				Ω(params[param2Name].Description).Should(Equal(param2Desc))
			})

			Context("and a BasePath using them", func() {
				const basePath = "/:accountID/:id"

				BeforeEach(func() {
					prevDSL := dsl
					dsl = func() {
						BasePath(basePath)
						prevDSL()
					}
				})

				It("sets both the base path and parameters", func() {
					Ω(Design.BaseParams.Type).Should(BeAssignableToTypeOf(Object{}))
					params := Design.BaseParams.Type.ToObject()
					Ω(params).Should(HaveLen(2))
					Ω(params).Should(HaveKey(param1Name))
					Ω(params).Should(HaveKey(param2Name))
					Ω(params[param1Name].Type).Should(Equal(param1Type))
					Ω(params[param2Name].Type).Should(Equal(param2Type))
					Ω(params[param1Name].Description).Should(Equal(param1Desc))
					Ω(params[param2Name].Description).Should(Equal(param2Desc))
					Ω(Design.BasePath).Should(Equal(basePath))
				})
			})
		})

		Context("with ResponseTemplates", func() {
			const respName = "NotFound2"
			const respDesc = "Resource Not Found"
			const respStatus = 404
			const respMediaType = "text/plain"
			const respTName = "OK"
			const respTDesc = "All good"
			const respTStatus = 200

			BeforeEach(func() {
				dsl = func() {
					ResponseTemplate(respName, func() {
						Description(respDesc)
						Status(respStatus)
						Media(respMediaType)
					})
					ResponseTemplate(respTName, func(mt string) {
						Description(respTDesc)
						Status(respTStatus)
						Media(mt)
					})
				}
			})

			It("sets the API responses and response templates", func() {
				Ω(Design.Responses).Should(HaveKey(respName))
				Ω(Design.Responses[respName]).ShouldNot(BeNil())
				expected := ResponseDefinition{
					Name:        respName,
					Description: respDesc,
					Status:      respStatus,
					MediaType:   respMediaType,
				}
				actual := *Design.Responses[respName]
				Ω(actual).Should(Equal(expected))

				Ω(Design.ResponseTemplates).Should(HaveLen(1))
				Ω(Design.ResponseTemplates).Should(HaveKey(respTName))
				Ω(Design.ResponseTemplates[respTName]).ShouldNot(BeNil())
			})
		})

		Context("with Traits", func() {
			const traitName = "Authenticated"

			BeforeEach(func() {
				dsl = func() {
					Trait(traitName, func() {
						Headers(func() {
							Header("Auth-Token")
							Required("Auth-Token")
						})
					})
				}
			})

			It("sets the API traits", func() {
				Ω(Design.Traits).Should(HaveLen(1))
				Ω(Design.Traits).Should(HaveKey(traitName))
			})
		})
	})
})
