package apidsl_test

import (
	"strconv"

	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Action", func() {
	var name string
	var dsl func()
	var action *ActionDefinition

	BeforeEach(func() {
		dslengine.Reset()
		name = ""
		dsl = nil
	})

	JustBeforeEach(func() {
		Resource("res", func() {
			Action(name, dsl)
		})
		dslengine.Run()
		if r, ok := Design.Resources["res"]; ok {
			action = r.Actions[name]
		}
	})

	Context("with only a name and a route", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces an invalid action", func() {
			Ω(dslengine.Errors).Should(HaveOccurred())
		})
	})

	Context("with a name and DSL defining a route", func() {
		var route = GET("/:id")

		BeforeEach(func() {
			name = "foo"
			dsl = func() { Routing(route) }
		})

		It("produces a valid action definition with the route and default status of 200 set", func() {
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(action).ShouldNot(BeNil())
			Ω(action.Name).Should(Equal(name))
			Ω(action.Validate()).ShouldNot(HaveOccurred())
			Ω(action.Routes).ShouldNot(BeNil())
			Ω(action.Routes).Should(HaveLen(1))
			Ω(action.Routes[0]).Should(Equal(route))
		})

		Context("with an empty params DSL", func() {
			BeforeEach(func() {
				olddsl := dsl
				dsl = func() { olddsl(); Params(func() {}) }
				name = "foo"
			})

			It("produces a valid action", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			})
		})

		Context("with a metadata", func() {
			BeforeEach(func() {
				metadatadsl := func() { Metadata("swagger:extension:x-get", `{"foo":"bar"}`) }
				route = GET("/:id", metadatadsl)
				name = "foo"
			})

			It("produces a valid action definition with the route with the metadata", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(action).ShouldNot(BeNil())
				Ω(action.Name).Should(Equal(name))
				Ω(action.Validate()).ShouldNot(HaveOccurred())
				Ω(action.Routes).ShouldNot(BeNil())
				Ω(action.Routes).Should(HaveLen(1))
				Ω(action.Routes[0]).Should(Equal(route))
				Ω(action.Routes[0].Metadata).ShouldNot(BeNil())
				Ω(action.Routes[0].Metadata).Should(Equal(
					dslengine.MetadataDefinition{"swagger:extension:x-get": []string{`{"foo":"bar"}`}},
				))
			})
		})
	})

	Context("with a string payload", func() {
		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Routing(GET("/:id"))
				Payload(String)
			}
		})

		It("produces a valid action with the given properties", func() {
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(action).ShouldNot(BeNil())
			Ω(action.Validate()).ShouldNot(HaveOccurred())
			Ω(action.Payload).ShouldNot(BeNil())
			Ω(action.Payload.Type).Should(Equal(String))
		})
	})

	Context("with a name and DSL defining a description, route, headers, payload and responses", func() {
		const typeName = "typeName"
		const description = "description"
		const headerName = "Foo"

		BeforeEach(func() {
			Type(typeName, func() {
				Attribute("name")
			})
			name = "foo"
			dsl = func() {
				Description(description)
				Routing(GET("/:id"))
				Headers(func() { Header(headerName) })
				Payload(typeName)
				Response(NoContent)
			}
		})

		It("produces a valid action with the given properties", func() {
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(action).ShouldNot(BeNil())
			Ω(action.Validate()).ShouldNot(HaveOccurred())
			Ω(action.Name).Should(Equal(name))
			Ω(action.Description).Should(Equal(description))
			Ω(action.Routes).Should(HaveLen(1))
			Ω(action.Responses).ShouldNot(BeNil())
			Ω(action.Responses).Should(HaveLen(1))
			Ω(action.Responses).Should(HaveKey("NoContent"))
			Ω(action.Headers.Type.(Object)).Should(HaveKey(headerName))
			Ω(action.Headers).ShouldNot(BeNil())
			Ω(action.Headers.Type).Should(BeAssignableToTypeOf(Object{}))
			Ω(action.Headers.Type.(Object)).Should(HaveLen(1))
			Ω(action.Headers.Type.(Object)).Should(HaveKey(headerName))
		})
	})

	Context("using a response with a media type modifier", func() {
		const mtID = "application/vnd.app.foo+json"

		BeforeEach(func() {
			MediaType(mtID, func() {
				Attributes(func() { Attribute("foo") })
				View("default", func() { Attribute("foo") })
			})
			name = "foo"
			dsl = func() {
				Routing(GET("/:id"))
				Response(OK, mtID)
			}
		})

		It("produces a response that keeps the modifier", func() {
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(action).ShouldNot(BeNil())
			Ω(action.Validate()).ShouldNot(HaveOccurred())
			Ω(action.Responses).ShouldNot(BeNil())
			Ω(action.Responses).Should(HaveLen(1))
			Ω(action.Responses).Should(HaveKey("OK"))
			resp := action.Responses["OK"]
			Ω(resp.MediaType).Should(Equal(mtID))
		})
	})

	Context("using a response template", func() {
		const tmplName = "tmpl"
		const respMediaType = "media"
		const respStatus = 200
		const respName = "respName"

		BeforeEach(func() {
			name = "foo"
			API("test", func() {
				ResponseTemplate(tmplName, func(status, name string) {
					st, err := strconv.Atoi(status)
					if err != nil {
						dslengine.ReportError(err.Error())
						return
					}
					Status(st)
				})
			})
		})

		Context("called correctly", func() {
			BeforeEach(func() {
				dsl = func() {
					Routing(GET("/:id"))
					Response(tmplName, strconv.Itoa(respStatus), respName, func() {
						Media(respMediaType)
					})
				}
			})

			It("defines the response definition using the template", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(action).ShouldNot(BeNil())
				Ω(action.Responses).ShouldNot(BeNil())
				Ω(action.Responses).Should(HaveLen(1))
				Ω(action.Responses).Should(HaveKey(tmplName))
				resp := action.Responses[tmplName]
				Ω(resp.Name).Should(Equal(tmplName))
				Ω(resp.Status).Should(Equal(respStatus))
				Ω(resp.MediaType).Should(Equal(respMediaType))
			})
		})

		Context("called incorrectly", func() {
			BeforeEach(func() {
				dsl = func() {
					Routing(GET("/id"))
					Response(tmplName, "not an integer", respName, func() {
						Media(respMediaType)
					})
				}
			})

			It("fails", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})
	})
})

var _ = Describe("Payload", func() {
	Context("with a payload definition", func() {
		BeforeEach(func() {
			dslengine.Reset()

			Resource("foo", func() {
				Action("bar", func() {
					Routing(GET(""))
					Payload(func() {
						Member("name")
						Required("name")
					})
				})
			})
		})

		JustBeforeEach(func() {
			dslengine.Run()
		})

		It("generates the payload type", func() {
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(Design).ShouldNot(BeNil())
			Ω(Design.Resources).Should(HaveKey("foo"))
			Ω(Design.Resources["foo"].Actions).Should(HaveKey("bar"))
			Ω(Design.Resources["foo"].Actions["bar"].Payload).ShouldNot(BeNil())
		})
	})

	Context("with an array", func() {
		BeforeEach(func() {
			dslengine.Reset()

			Resource("foo", func() {
				Action("bar", func() {
					Routing(GET(""))
					Payload(ArrayOf(Integer))
				})
			})
		})

		JustBeforeEach(func() {
			dslengine.Run()
		})

		It("sets the payload type", func() {
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(Design).ShouldNot(BeNil())
			Ω(Design.Resources).Should(HaveKey("foo"))
			Ω(Design.Resources["foo"].Actions).Should(HaveKey("bar"))
			Ω(Design.Resources["foo"].Actions["bar"].Payload).ShouldNot(BeNil())
			Ω(Design.Resources["foo"].Actions["bar"].Payload.Type).ShouldNot(BeNil())
			Ω(Design.Resources["foo"].Actions["bar"].Payload.Type.IsArray()).Should(BeTrue())
			Ω(Design.Resources["foo"].Actions["bar"].Payload.Type.ToArray().ElemType).ShouldNot(BeNil())
			Ω(Design.Resources["foo"].Actions["bar"].Payload.Type.ToArray().ElemType.Type).Should(Equal(Integer))
		})
	})

	Context("with a hash", func() {
		BeforeEach(func() {
			dslengine.Reset()

			Resource("foo", func() {
				Action("bar", func() {
					Routing(GET(""))
					Payload(HashOf(String, Integer))
				})
			})
		})

		JustBeforeEach(func() {
			dslengine.Run()
		})

		It("sets the payload type", func() {
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(Design).ShouldNot(BeNil())
			Ω(Design.Resources).Should(HaveKey("foo"))
			Ω(Design.Resources["foo"].Actions).Should(HaveKey("bar"))
			Ω(Design.Resources["foo"].Actions["bar"].Payload).ShouldNot(BeNil())
			Ω(Design.Resources["foo"].Actions["bar"].Payload.Type).ShouldNot(BeNil())
			Ω(Design.Resources["foo"].Actions["bar"].Payload.Type.IsHash()).Should(BeTrue())
			Ω(Design.Resources["foo"].Actions["bar"].Payload.Type.ToHash().ElemType).ShouldNot(BeNil())
			Ω(Design.Resources["foo"].Actions["bar"].Payload.Type.ToHash().KeyType).ShouldNot(BeNil())
			Ω(Design.Resources["foo"].Actions["bar"].Payload.Type.ToHash().ElemType.Type).Should(Equal(Integer))
			Ω(Design.Resources["foo"].Actions["bar"].Payload.Type.ToHash().KeyType.Type).Should(Equal(String))
		})
	})

})
