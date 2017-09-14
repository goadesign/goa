package genswagger_test

import (
	"bytes"
	"encoding/json"

	"github.com/go-openapi/loads"
	_ "github.com/goadesign/goa-cellar/design"
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/gen_schema"
	"github.com/goadesign/goa/goagen/gen_swagger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// validateSwagger validates that the given swagger object represents a valid Swagger spec.
func validateSwagger(swagger *genswagger.Swagger) {
	b, err := json.Marshal(swagger)
	Ω(err).ShouldNot(HaveOccurred())
	doc, err := loads.Analyzed(json.RawMessage(b), "")
	Ω(err).ShouldNot(HaveOccurred())
	Ω(doc).ShouldNot(BeNil())
}

// validateSwaggerWithFragments validates that the given swagger object represents a valid Swagger spec
// and contains fragments
func validateSwaggerWithFragments(swagger *genswagger.Swagger, fragments [][]byte) {
	b, err := json.Marshal(swagger)
	Ω(err).ShouldNot(HaveOccurred())
	doc, err := loads.Analyzed(json.RawMessage(b), "")
	Ω(err).ShouldNot(HaveOccurred())
	Ω(doc).ShouldNot(BeNil())
	for _, sub := range fragments {
		Ω(bytes.Contains(b, sub)).Should(BeTrue())
	}
}

var _ = Describe("New", func() {
	var swagger *genswagger.Swagger
	var newErr error

	BeforeEach(func() {
		swagger = nil
		newErr = nil
		dslengine.Reset()
		genschema.Definitions = make(map[string]*genschema.JSONSchema)
	})

	JustBeforeEach(func() {
		err := dslengine.Run()
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
				Metadata("swagger:tag:" + tag)
				Metadata("swagger:tag:"+tag+":desc", "Tag desc.")
				Metadata("swagger:tag:"+tag+":url", "http://example.com/tag")
				Metadata("swagger:tag:"+tag+":url:desc", "Huge docs")
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
				Paths:    make(map[string]interface{}),
				Consumes: []string{"application/json", "application/xml", "application/gob", "application/x-gob"},
				Produces: []string{"application/json", "application/xml", "application/gob", "application/x-gob"},
				Tags: []*genswagger.Tag{{Name: tag, Description: "Tag desc.", ExternalDocs: &genswagger.ExternalDocs{
					URL: "http://example.com/tag", Description: "Huge docs",
				}}},
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
					Params(func() {
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
				Ω(*swagger.Parameters[intParam].Minimum).Should(Equal(intMin))
				Ω(swagger.Parameters[numParam]).ShouldNot(BeNil())
				Ω(swagger.Parameters[numParam].Name).Should(Equal(numParam))
				Ω(swagger.Parameters[numParam].In).Should(Equal("path"))
				Ω(swagger.Parameters[numParam].Required).Should(BeTrue())
				Ω(swagger.Parameters[numParam].Type).Should(Equal("number"))
				Ω(*swagger.Parameters[numParam].Maximum).Should(Equal(floatMax))
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

		Context("with required payload", func() {
			BeforeEach(func() {
				p := Type("RequiredPayload", func() {
					Member("m1", String)
				})
				Resource("res", func() {
					Action("act", func() {
						Routing(
							PUT("/"),
						)
						Payload(p)
					})
				})
			})

			It("serializes into valid swagger JSON", func() {
				validateSwaggerWithFragments(swagger, [][]byte{
					[]byte(`"required":true`),
				})
			})
		})

		Context("with a payload of type Any", func() {
			BeforeEach(func() {
				Resource("res", func() {
					Action("act", func() {
						Routing(
							PUT("/"),
						)
						Payload(Any, func() {
							Example("example")
						})
					})
				})
			})

			It("serializes into valid swagger JSON", func() {
				validateSwaggerWithFragments(swagger, [][]byte{
					[]byte(`"ActResPayload":{"title":"ActResPayload","example":"example"}`),
				})
			})

		})

		Context("with optional payload", func() {
			BeforeEach(func() {
				p := Type("OptionalPayload", func() {
					Member("m1", String)
				})
				Resource("res", func() {
					Action("act", func() {
						Routing(
							PUT("/"),
						)
						OptionalPayload(p)
					})
				})
			})

			It("serializes into valid swagger JSON", func() {
				validateSwaggerWithFragments(swagger, [][]byte{
					[]byte(`"required":false`),
				})
			})

		})

		Context("with zero value validations", func() {
			const (
				intParam = "intParam"
				numParam = "numParam"
				strParam = "strParam"
				intMin   = 0.0
				floatMax = 0.0
			)

			BeforeEach(func() {
				PayloadWithZeroValueValidations := Type("PayloadWithZeroValueValidations", func() {
					Attribute(strParam, String, func() {
						MinLength(0)
						MaxLength(0)
					})
				})
				Resource("res", func() {
					Action("act", func() {
						Routing(
							PUT("/"),
						)
						Params(func() {
							Param(intParam, Integer, func() {
								Minimum(intMin)
							})
							Param(numParam, Number, func() {
								Maximum(floatMax)
							})
						})
						Payload(PayloadWithZeroValueValidations)
					})
				})
			})

			It("serializes into valid swagger JSON", func() {
				validateSwaggerWithFragments(swagger, [][]byte{
					// payload
					[]byte(`"minLength":0`),
					[]byte(`"maxLength":0`),
					// param
					[]byte(`"minimum":0`),
					[]byte(`"maximum":0`),
				})
			})
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
			var (
				minLength1  = 1
				maxLength10 = 10
				minimum_2   = -2.0
				maximum2    = 2.0
				minItems1   = 1
				maxItems5   = 5
			)
			BeforeEach(func() {
				Country := MediaType("application/vnd.goa.example.origin", func() {
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
						Attribute("origin", Country, "Details on wine origin")
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
					Attribute("origin", Country, "Details on wine origin")
					Required("name")
				})
				Resource("res", func() {
					Metadata("swagger:tag:res")
					Description("A wine bottle")
					DefaultMedia(BottleMedia)
					BasePath("/bottles")
					UseTrait("Authenticated")

					Action("Update", func() {
						Metadata("swagger:tag:Update")
						Metadata("swagger:summary", "a summary")
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
							Header("OptionalBoolWithDefault", Boolean, "defaults true", func() {
								Default(true)
							})
							Header("OptionalRegex", String, func() {
								Pattern(`[a-z]\d+`)
								MinLength(minLength1)
								MaxLength(maxLength10)
							})
							Header("OptionalInt", Integer, func() {
								Minimum(minimum_2)
								Maximum(maximum2)
							})
							Header("OptionalArray", ArrayOf(String), func() {
								// interpreted as MinItems & MaxItems:
								MinLength(minItems1)
								MaxLength(maxItems5)
							})
							Header("OverrideRequiredHeader")
							Header("OverrideOptionalHeader")
							Required("Authorization", "X-Account", "OverrideOptionalHeader")
						})
						Payload(UpdatePayload)
						Response(OK, func() {
							Media(CollectionOf(BottleMedia), "extended")
						})
						Response(NoContent)
						Response(NotFound, ErrorMedia)
						Response(BadRequest, ErrorMedia)
					})

					Action("hidden", func() {
						Description("Does not show up in Swagger spec")
						Metadata("swagger:generate", "false")
						Routing(GET("/hidden"))
						Response(OK)
					})
				})
				base := Design.DSLFunc
				Design.DSLFunc = func() {
					base()
					Trait("Authenticated", func() {
						Headers(func() {
							Header("header")
							Header("OverrideRequiredHeader", String, "to be overridden in Action and not marked Required")
							Header("OverrideOptionalHeader", String, "to be overridden in Action and marked Required")
							Header("OptionalResourceHeaderWithEnum", func() {
								Enum("a", "b")
							})
							Required("header", "OverrideRequiredHeader")
						})
					})
				}
			})

			It("sets the Path fields", func() {
				Ω(newErr).ShouldNot(HaveOccurred())
				Ω(swagger.Paths).Should(HaveLen(2))
				Ω(swagger.Paths["/orgs/{org}/accounts/{id}"]).ShouldNot(BeNil())
				a := swagger.Paths["/orgs/{org}/accounts/{id}"].(*genswagger.Path)
				Ω(a.Put).ShouldNot(BeNil())
				ps := a.Put.Parameters
				Ω(ps).Should(HaveLen(14))
				// check Headers in detail
				Ω(ps[3]).Should(Equal(&genswagger.Parameter{In: "header", Name: "Authorization", Type: "string", Required: true}))
				Ω(ps[4]).Should(Equal(&genswagger.Parameter{In: "header", Name: "OptionalArray", Type: "array",
					Items: &genswagger.Items{Type: "string"}, MinItems: &minItems1, MaxItems: &maxItems5}))
				Ω(ps[5]).Should(Equal(&genswagger.Parameter{In: "header", Name: "OptionalBoolWithDefault", Type: "boolean",
					Description: "defaults true", Default: true}))
				Ω(ps[6]).Should(Equal(&genswagger.Parameter{In: "header", Name: "OptionalInt", Type: "integer", Minimum: &minimum_2, Maximum: &maximum2}))
				Ω(ps[7]).Should(Equal(&genswagger.Parameter{In: "header", Name: "OptionalRegex", Type: "string",
					Pattern: `[a-z]\d+`, MinLength: &minLength1, MaxLength: &maxLength10}))
				Ω(ps[8]).Should(Equal(&genswagger.Parameter{In: "header", Name: "OptionalResourceHeaderWithEnum", Type: "string",
					Enum: []interface{}{"a", "b"}}))
				Ω(ps[9]).Should(Equal(&genswagger.Parameter{In: "header", Name: "OverrideOptionalHeader", Type: "string", Required: true}))
				Ω(ps[10]).Should(Equal(&genswagger.Parameter{In: "header", Name: "OverrideRequiredHeader", Type: "string", Required: true}))
				Ω(ps[11]).Should(Equal(&genswagger.Parameter{In: "header", Name: "X-Account", Type: "integer", Required: true}))
				Ω(ps[12]).Should(Equal(&genswagger.Parameter{In: "header", Name: "header", Type: "string", Required: true}))
				Ω(swagger.Paths["/base/bottles/{id}"]).ShouldNot(BeNil())
				b := swagger.Paths["/base/bottles/{id}"].(*genswagger.Path)
				Ω(b.Put).ShouldNot(BeNil())
				Ω(b.Put.Parameters).Should(HaveLen(14))
				Ω(b.Put.Produces).Should(Equal([]string{"application/vnd.goa.error", "application/vnd.goa.example.bottle; type=collection"}))
			})

			It("should set the inherited tag and the action tag", func() {
				tags := []string{"res", "Update"}
				a := swagger.Paths["/orgs/{org}/accounts/{id}"].(*genswagger.Path)
				Ω(a.Put).ShouldNot(BeNil())
				Ω(a.Put.Tags).Should(Equal(tags))
				b := swagger.Paths["/base/bottles/{id}"].(*genswagger.Path)
				Ω(b.Put.Tags).Should(Equal(tags))
			})

			It("sets the summary from the summary tag", func() {
				a := swagger.Paths["/orgs/{org}/accounts/{id}"].(*genswagger.Path)
				Ω(a.Put.Summary).Should(Equal("a summary"))
			})

			It("generates the media type collection schema", func() {
				Ω(swagger.Definitions).Should(HaveLen(7))
				Ω(swagger.Definitions).Should(HaveKey("GoaExampleBottleExtendedCollection"))
			})

			It("serializes into valid swagger JSON", func() { validateSwagger(swagger) })
		})

		Context("with metadata", func() {
			const gat = "gat"
			const extension = `{"foo":"bar"}`
			const stringExtension = "foo"

			var (
				unmarshaled map[string]interface{}
				_           = json.Unmarshal([]byte(extension), &unmarshaled)
			)

			BeforeEach(func() {
				Resource("res", func() {
					Metadata("swagger:tag:res")
					Metadata("struct:tag:json", "resource")
					Metadata("swagger:extension:x-resource", extension)
					Metadata("swagger:extension:x-string", stringExtension)
					Action("act", func() {
						Metadata("swagger:tag:Update")
						Metadata("struct:tag:json", "action")
						Metadata("swagger:extension:x-action", extension)
						Security("password", func() {
							Metadata("swagger:extension:x-security", extension)
						})
						Routing(
							PUT("/", func() {
								Metadata("swagger:extension:x-put", extension)
							}),
						)
						Params(func() {
							Param("param", func() {
								Metadata("swagger:extension:x-param", extension)
							})
						})
						Response(NoContent, func() {
							Metadata("swagger:extension:x-response", extension)
						})
					})
				})
				base := Design.DSLFunc
				Design.DSLFunc = func() {
					base()
					Metadata("swagger:tag:" + gat)
					Metadata("struct:tag:json", "api")
					Metadata("swagger:extension:x-api", extension)
					BasicAuthSecurity("password")
				}
			})

			It("should set the swagger object tags", func() {
				Ω(swagger.Tags).Should(HaveLen(2))
				tags := []*genswagger.Tag{
					{Name: gat, Description: "", ExternalDocs: nil, Extensions: map[string]interface{}{"x-api": unmarshaled}},
					{Name: tag, Description: "Tag desc.", ExternalDocs: &genswagger.ExternalDocs{URL: "http://example.com/tag", Description: "Huge docs"}, Extensions: map[string]interface{}{"x-api": unmarshaled}},
				}
				Ω(swagger.Tags).Should(Equal(tags))
			})

			It("should set the action tags", func() {
				p := swagger.Paths["/"].(*genswagger.Path)
				Ω(p.Put.Tags).Should(HaveLen(2))
				tags := []string{"res", "Update"}
				Ω(p.Put.Tags).Should(Equal(tags))
			})

			It("should set the swagger extensions", func() {
				Ω(swagger.Info.Extensions).Should(HaveLen(1))
				Ω(swagger.Info.Extensions["x-api"]).Should(Equal(unmarshaled))
				p := swagger.Paths["/"].(*genswagger.Path)
				Ω(p.Extensions).Should(HaveLen(1))
				Ω(p.Extensions["x-action"]).Should(Equal(unmarshaled))
				Ω(p.Put.Extensions).Should(HaveLen(1))
				Ω(p.Put.Extensions["x-put"]).Should(Equal(unmarshaled))
				Ω(p.Put.Parameters[0].Extensions).Should(HaveLen(1))
				Ω(p.Put.Parameters[0].Extensions["x-param"]).Should(Equal(unmarshaled))
				Ω(p.Put.Responses["204"].Extensions).Should(HaveLen(1))
				Ω(p.Put.Responses["204"].Extensions["x-response"]).Should(Equal(unmarshaled))
				Ω(swagger.Paths["x-resource"]).ShouldNot(BeNil())
				rs := swagger.Paths["x-resource"].(map[string]interface{})
				Ω(rs).Should(Equal(unmarshaled))
				rs2 := swagger.Paths["x-string"].(string)
				Ω(rs2).Should(Equal(stringExtension))
				Ω(swagger.SecurityDefinitions["password"].Extensions).Should(HaveLen(1))
				Ω(swagger.SecurityDefinitions["password"].Extensions["x-security"]).Should(Equal(unmarshaled))
			})

		})
	})
})
