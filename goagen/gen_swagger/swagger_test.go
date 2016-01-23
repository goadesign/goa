package genswagger_test

import (
	"encoding/json"

	"github.com/go-swagger/go-swagger/spec"
	_ "github.com/goadesign/goa-cellar/design"
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/goagen/gen_schema"
	"github.com/goadesign/goa/goagen/gen_swagger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Save cellar example API definition for tests.
var cellarDesign *APIDefinition

func init() {
	cellarDesign = Design
}

// validateSwagger validates that the given swagger object represents a valid Swagger spec.
func validateSwagger(swagger *genswagger.Swagger) {
	b, err := json.Marshal(swagger)
	Ω(err).ShouldNot(HaveOccurred())
	doc, err := spec.New(b, "")
	Ω(err).ShouldNot(HaveOccurred())
	Ω(doc).ShouldNot(BeNil())
	// TBD calling the swagger validator below causes Travis to hang...
	// err = validate.Spec(doc, strfmt.NewFormats())
	// Ω(err).ShouldNot(HaveOccurred())
}

var _ = Describe("New", func() {
	var swagger *genswagger.Swagger
	var newErr error

	BeforeEach(func() {
		swagger = nil
		newErr = nil
		InitDesign()
		genschema.Definitions = make(map[string]*genschema.JSONSchema)
	})

	JustBeforeEach(func() {
		err := RunDSL()
		Ω(err).ShouldNot(HaveOccurred())
		swagger, newErr = genswagger.New(Design)
	})

	Context("with a valid API definition", func() {
		const (
			title        = "title"
			description  = "description"
			terms        = "terms"
			contactEmail = "contactEmail@goa.design"
			contactName  = "contactName"
			contactURL   = "http://contactURL.com"
			license      = "license"
			licenseURL   = "http://licenseURL.com"
			host         = "host"
			scheme       = "https"
			basePath     = "/base"
			tag          = "tag"
			docDesc      = "doc description"
			docURL       = "http://docURL.com"
		)

		BeforeEach(func() {
			API("test", func() {
				Title(title)
				Metadata("tags", `[{"name": "`+tag+`"}]`)
				Description(description)
				TermsOfService(terms)
				Contact(func() {
					Email(contactEmail)
					Name(contactName)
					URL(contactURL)
				})
				License(func() {
					Name(license)
					URL(licenseURL)
				})
				Docs(func() {
					Description(docDesc)
					URL(docURL)
				})
				Host(host)
				Scheme(scheme)
				BasePath(basePath)
			})
		})

		It("sets all the basic fields", func() {
			Ω(newErr).ShouldNot(HaveOccurred())
			Ω(swagger).Should(Equal(&genswagger.Swagger{
				Swagger: "2.0",
				Info: &genswagger.Info{
					Title:          title,
					Description:    description,
					TermsOfService: terms,
					Contact: &ContactDefinition{
						Name:  contactName,
						Email: contactEmail,
						URL:   contactURL,
					},
					License: &LicenseDefinition{
						Name: license,
						URL:  licenseURL,
					},
					Version: "",
				},
				Host:     host,
				BasePath: basePath,
				Schemes:  []string{"https"},
				Paths:    make(map[string]*genswagger.Path),
				Consumes: []string{"application/json", "application/xml", "text/xml", "application/gob", "application/x-gob"},
				Produces: []string{"application/json", "application/xml", "text/xml", "application/gob", "application/x-gob"},
				Tags:     []*genswagger.Tag{{Name: tag}},
				ExternalDocs: &genswagger.ExternalDocs{
					Description: docDesc,
					URL:         docURL,
				},
			}))
		})

		It("serializes into valid swagger JSON", func() { validateSwagger(swagger) })

		Context("with base params", func() {
			const (
				basePath    = "/s/:strParam/i/:intParam/n/:numParam/b/:boolParam"
				strParam    = "strParam"
				intParam    = "intParam"
				numParam    = "numParam"
				boolParam   = "boolParam"
				queryParam  = "queryParam"
				description = "description"
				intMin      = 1.0
				floatMax    = 2.4
				enum1       = "enum1"
				enum2       = "enum2"
			)

			BeforeEach(func() {
				base := Design.DSLFunc
				Design.DSLFunc = func() {
					base()
					BasePath(basePath)
					BaseParams(func() {
						Param(strParam, String, func() {
							Description(description)
							Format("email")
						})
						Param(intParam, Integer, func() {
							Minimum(intMin)
						})
						Param(numParam, Number, func() {
							Maximum(floatMax)
						})
						Param(boolParam, Boolean)
						Param(queryParam, func() {
							Enum(enum1, enum2)
						})
					})
				}
			})

			It("sets the BasePath and Parameters fields", func() {
				Ω(newErr).ShouldNot(HaveOccurred())
				Ω(swagger.BasePath).Should(Equal(basePath))
				Ω(swagger.Parameters).Should(HaveLen(5))
				Ω(swagger.Parameters[strParam]).ShouldNot(BeNil())
				Ω(swagger.Parameters[strParam].Name).Should(Equal(strParam))
				Ω(swagger.Parameters[strParam].In).Should(Equal("path"))
				Ω(swagger.Parameters[strParam].Description).Should(Equal("description"))
				Ω(swagger.Parameters[strParam].Required).Should(BeTrue())
				Ω(swagger.Parameters[strParam].Type).Should(Equal("string"))
				Ω(swagger.Parameters[strParam].Format).Should(Equal("email"))
				Ω(swagger.Parameters[intParam]).ShouldNot(BeNil())
				Ω(swagger.Parameters[intParam].Name).Should(Equal(intParam))
				Ω(swagger.Parameters[intParam].In).Should(Equal("path"))
				Ω(swagger.Parameters[intParam].Required).Should(BeTrue())
				Ω(swagger.Parameters[intParam].Type).Should(Equal("integer"))
				Ω(swagger.Parameters[intParam].Minimum).Should(Equal(intMin))
				Ω(swagger.Parameters[numParam]).ShouldNot(BeNil())
				Ω(swagger.Parameters[numParam].Name).Should(Equal(numParam))
				Ω(swagger.Parameters[numParam].In).Should(Equal("path"))
				Ω(swagger.Parameters[numParam].Required).Should(BeTrue())
				Ω(swagger.Parameters[numParam].Type).Should(Equal("number"))
				Ω(swagger.Parameters[numParam].Maximum).Should(Equal(floatMax))
				Ω(swagger.Parameters[boolParam]).ShouldNot(BeNil())
				Ω(swagger.Parameters[boolParam].Name).Should(Equal(boolParam))
				Ω(swagger.Parameters[boolParam].In).Should(Equal("path"))
				Ω(swagger.Parameters[boolParam].Required).Should(BeTrue())
				Ω(swagger.Parameters[boolParam].Type).Should(Equal("boolean"))
				Ω(swagger.Parameters[queryParam]).ShouldNot(BeNil())
				Ω(swagger.Parameters[queryParam].Name).Should(Equal(queryParam))
				Ω(swagger.Parameters[queryParam].In).Should(Equal("query"))
				Ω(swagger.Parameters[queryParam].Type).Should(Equal("string"))
				Ω(swagger.Parameters[queryParam].Enum).Should(Equal([]interface{}{enum1, enum2}))
			})

			It("serializes into valid swagger JSON", func() { validateSwagger(swagger) })
		})

		Context("with response templates", func() {
			const okName = "OK"
			const okDesc = "OK description"
			const notFoundName = "NotFound"
			const notFoundDesc = "NotFound description"
			const notFoundMt = "application/json"
			const headerName = "headerName"

			BeforeEach(func() {
				account := MediaType("application/vnd.goa.test.account", func() {
					Description("Account")
					Attributes(func() {
						Attribute("id", Integer)
						Attribute("href", String)
					})
					View("default", func() {
						Attribute("id")
						Attribute("href")
					})
					View("link", func() {
						Attribute("id")
						Attribute("href")
					})
				})
				mt := MediaType("application/vnd.goa.test.bottle", func() {
					Description("A bottle of wine")
					Attributes(func() {
						Attribute("id", Integer, "ID of bottle")
						Attribute("href", String, "API href of bottle")
						Attribute("account", account, "Owner account")
						Links(func() {
							Link("account") // Defines a link to the Account media type
						})
						Required("id", "href")
					})
					View("default", func() {
						Attribute("id")
						Attribute("href")
						Attribute("links") // Default view renders links
					})
					View("extended", func() {
						Attribute("id")
						Attribute("href")
						Attribute("account") // Extended view renders account inline
						Attribute("links")   // Extended view also renders links
					})
				})
				base := Design.DSLFunc
				Design.DSLFunc = func() {
					base()
					ResponseTemplate(okName, func() {
						Description(okDesc)
						Status(404)
						Media(mt)
						Headers(func() {
							Header(headerName, func() {
								Format("hostname")
							})
						})
					})
					ResponseTemplate(notFoundName, func() {
						Description(notFoundDesc)
						Status(404)

						Media(notFoundMt)
					})
				}
			})

			It("sets the Responses fields", func() {
				Ω(newErr).ShouldNot(HaveOccurred())
				Ω(swagger.Responses).Should(HaveLen(2))
				Ω(swagger.Responses[notFoundName]).ShouldNot(BeNil())
				Ω(swagger.Responses[notFoundName].Description).Should(Equal(notFoundDesc))
				Ω(swagger.Responses[okName]).ShouldNot(BeNil())
				Ω(swagger.Responses[okName].Description).Should(Equal(okDesc))
			})

			It("serializes into valid swagger JSON", func() { validateSwagger(swagger) })
		})

		Context("with resources", func() {
			BeforeEach(func() {
				Origin := MediaType("application/vnd.goa.example.origin", func() {
					Description("Origin of bottle")
					Attributes(func() {
						Attribute("id")
						Attribute("href")
						Attribute("country")
					})
					View("default", func() {
						Attribute("id")
						Attribute("href")
						Attribute("country")
					})
					View("tiny", func() {
						Attribute("id")
					})
				})
				BottleMedia := MediaType("application/vnd.goa.example.bottle", func() {
					Description("A bottle of wine")
					Attributes(func() {
						Attribute("id", Integer, "ID of bottle")
						Attribute("href", String, "API href of bottle")
						Attribute("origin", Origin, "Details on wine origin")
						Links(func() {
							Link("origin", "tiny")
						})
						Required("id", "href")
					})
					View("default", func() {
						Attribute("id")
						Attribute("href")
						Attribute("links")
					})
					View("extended", func() {
						Attribute("id")
						Attribute("href")
						Attribute("origin")
						Attribute("links")
					})
				})
				UpdatePayload := Type("UpdatePayload", func() {
					Description("Type of create and upload action payloads")
					Attribute("name", String, "name of bottle")
					Attribute("origin", Origin, "Details on wine origin")
					Required("name")
				})
				Resource("res", func() {
					Metadata("tags", `[{"name": "res"}]`)
					Description("A wine bottle")
					DefaultMedia(BottleMedia)
					BasePath("/bottles")
					UseTrait("Authenticated")

					Action("Update", func() {
						Metadata("tags", `[{"name": "Update"}]`)
						Description("Update account")
						Docs(func() {
							Description("docs")
							URL("http://cellarapi.com/docs/actions/update")
						})
						Routing(
							PUT("/:id"),
							PUT("//orgs/:org/accounts/:id"),
						)
						Params(func() {
							Param("org", String)
							Param("id", Integer)
							Param("sort", func() {
								Enum("asc", "desc")
							})
						})
						Headers(func() {
							Header("Authorization", String)
							Header("X-Account", Integer)
							Required("Authorization", "X-Account")
						})
						Payload(UpdatePayload)
						Response(NoContent)
						Response(NotFound)
					})
				})
				base := Design.DSLFunc
				Design.DSLFunc = func() {
					base()
					Trait("Authenticated", func() {
						Headers(func() {
							Header("header")
							Required("header")
						})
					})
				}
			})

			It("sets the Path fields", func() {
				Ω(newErr).ShouldNot(HaveOccurred())
				Ω(swagger.Paths).Should(HaveLen(2))
				Ω(swagger.Paths["/orgs/{org}/accounts/{id}"]).ShouldNot(BeNil())
				Ω(swagger.Paths["/orgs/{org}/accounts/{id}"].Put).ShouldNot(BeNil())
				Ω(swagger.Paths["/orgs/{org}/accounts/{id}"].Put.Parameters).Should(HaveLen(4))
				Ω(swagger.Paths["/bottles/{id}"]).ShouldNot(BeNil())
				Ω(swagger.Paths["/bottles/{id}"].Put).ShouldNot(BeNil())
				Ω(swagger.Paths["/bottles/{id}"].Put.Parameters).Should(HaveLen(4))
			})

			It("should set the inherited tag and the action tag", func() {
				tags := []string{"res", "Update"}
				Ω(swagger.Paths["/orgs/{org}/accounts/{id}"].Put.Tags).Should(Equal(tags))
				Ω(swagger.Paths["/bottles/{id}"].Put.Tags).Should(Equal(tags))
			})

			It("serializes into valid swagger JSON", func() { validateSwagger(swagger) })
		})
	})

	Context("using the cellar example API definition", func() {
		BeforeEach(func() {
			Design = cellarDesign
		})

		It("serializes into valid swagger JSON", func() { validateSwagger(swagger) })
	})

})
