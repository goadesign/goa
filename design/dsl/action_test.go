package dsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Describe("Action", func() {
	var name string
	var dsl func()
	var parent *ResourceDefinition

	BeforeEach(func() {
		dsl = nil
		name = ""
		InitDesign()
		DSLErrors = nil
		parent = &ResourceDefinition{}
		Reset([]DSLDefinition{parent})
	})

	JustBeforeEach(func() {
		Action(name, dsl)
	})

	Context("with only a name", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces an invalid action", func() {
			Ω(parent.Actions).Should(HaveKey(name))
			Ω(parent.Actions[name]).ShouldNot(BeNil())
			Ω(parent.Actions[name].Validate()).Should(HaveOccurred())
		})
	})

	Context("with a name and DSL defining a route", func() {
		BeforeEach(func() {
			name = "foo"
			dsl = func() { Routing(GET("/:id")) }
		})

		It("produces a valid action with the given route", func() {
			Ω(parent.Actions).Should(HaveKey(name))
			action := parent.Actions[name]
			Ω(action).ShouldNot(BeNil())
			Ω(action.Validate()).ShouldNot(HaveOccurred())
			Ω(action.Name).Should(Equal(name))
			Ω(action.Routes).Should(HaveLen(1))
		})
	})

	Context("with a name and DSL defining a description, route, headers, payload and responses", func() {
		var payload *UserTypeDefinition
		const description = "description"
		const headerName = "Foo"
		BeforeEach(func() {
			name = "foo"
			pat := AttributeDefinition{
				Type: String,
			}
			payload = &UserTypeDefinition{
				AttributeDefinition: &pat,
				TypeName:            "id",
			}
			dsl = func() {
				Description(description)
				Routing(GET("/:id"))
				Headers(func() { Header(headerName) })
				Payload(payload)
				Response(NoContent)
			}
		})

		It("produces a valid action with the given properties", func() {
			Ω(parent.Actions).Should(HaveKey(name))
			action := parent.Actions[name]
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

	Context("using a response template", func() {
		var tmplDef *ResponseTemplateDefinition
		const tmplName = "tmpl"
		const respMediaType = "media"
		const respStatus = 200
		const respName = "respName"

		BeforeEach(func() {
			name = "foo"
			t := func(v ...string) *ResponseDefinition {
				return &ResponseDefinition{
					Name:   v[0],
					Status: respStatus,
				}
			}
			tmplDef = &ResponseTemplateDefinition{
				Name:     tmplName,
				Template: t,
			}
			responseTemplates := map[string]*ResponseTemplateDefinition{tmplName: tmplDef}
			Design = &APIDefinition{
				ResponseTemplates: responseTemplates,
			}
			dsl = func() {
				Routing(GET("/:id"))
				Response(tmplName, respName, func() {
					MediaType(respMediaType)
				})
			}
		})

		It("defines the response definition using the template", func() {
			Ω(parent.Actions).Should(HaveKey(name))
			action := parent.Actions[name]
			Ω(action).ShouldNot(BeNil())
			Ω(action.Responses).ShouldNot(BeNil())
			Ω(action.Responses).Should(HaveLen(1))
			Ω(action.Responses).Should(HaveKey(tmplName))
			resp := action.Responses[tmplName]
			Ω(resp.Name).Should(Equal(respName))
			Ω(resp.Status).Should(Equal(respStatus))
			Ω(resp.MediaType).Should(Equal(respMediaType))
		})
	})
})
