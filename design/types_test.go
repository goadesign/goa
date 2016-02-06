package design_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Project", func() {
	var mt *MediaTypeDefinition
	var view string

	var projected *MediaTypeDefinition
	var links *UserTypeDefinition
	var prErr error

	JustBeforeEach(func() {
		projected, links, prErr = mt.Project(view)
	})

	Context("with a media type with a default and a tiny view", func() {
		BeforeEach(func() {
			GeneratedMediaTypes = nil
			mt = &MediaTypeDefinition{
				UserTypeDefinition: &UserTypeDefinition{
					AttributeDefinition: &AttributeDefinition{
						Type: Object{
							"att1": &AttributeDefinition{Type: Integer},
							"att2": &AttributeDefinition{Type: String},
						},
					},
					TypeName: "Foo",
				},
				Identifier: "vnd.application/foo",
				Views: map[string]*ViewDefinition{
					"default": &ViewDefinition{
						Name: "default",
						AttributeDefinition: &AttributeDefinition{
							Type: Object{
								"att1": &AttributeDefinition{Type: String},
								"att2": &AttributeDefinition{Type: String},
							},
						},
					},
					"tiny": &ViewDefinition{
						Name: "tiny",
						AttributeDefinition: &AttributeDefinition{
							Type: Object{
								"att2": &AttributeDefinition{Type: String},
							},
						},
					},
				},
			}
		})

		Context("using the empty view", func() {
			BeforeEach(func() {
				view = ""
			})

			It("returns an error", func() {
				Ω(prErr).Should(HaveOccurred())
			})
		})

		Context("using the default view", func() {
			BeforeEach(func() {
				view = "default"
			})

			It("returns a media type with the default view attributes", func() {
				Ω(prErr).ShouldNot(HaveOccurred())
				Ω(projected).ShouldNot(BeNil())
				Ω(projected.Type).Should(BeAssignableToTypeOf(Object{}))
				Ω(projected.Type.ToObject()).Should(HaveKey("att1"))
				att := projected.Type.ToObject()["att1"]
				Ω(att).ShouldNot(BeNil())
				Ω(att.Type).ShouldNot(BeNil())
				Ω(att.Type.Kind()).Should(Equal(IntegerKind))
			})
		})

		Context("using the tiny view", func() {
			BeforeEach(func() {
				view = "tiny"
			})

			It("returns a media type with the default view attributes", func() {
				Ω(prErr).ShouldNot(HaveOccurred())
				Ω(projected).ShouldNot(BeNil())
				Ω(projected.Type).Should(BeAssignableToTypeOf(Object{}))
				Ω(projected.Type.ToObject()).Should(HaveKey("att2"))
				att := projected.Type.ToObject()["att2"]
				Ω(att).ShouldNot(BeNil())
				Ω(att.Type).ShouldNot(BeNil())
				Ω(att.Type.Kind()).Should(Equal(StringKind))
			})
		})

		Context("with a versioned media type", func() {

			BeforeEach(func() {
				view = "default"
				mt.APIVersions = []string{"v1"}
			})

			It("sets the version in the projected media type", func() {
				Ω(prErr).ShouldNot(HaveOccurred())
				Ω(projected).ShouldNot(BeNil())
				Ω(projected.Type).Should(BeAssignableToTypeOf(Object{}))
				Ω(projected.Type.ToObject()).Should(HaveKey("att1"))
				att := projected.Type.ToObject()["att1"]
				Ω(att).ShouldNot(BeNil())
				Ω(att.Type).ShouldNot(BeNil())
				Ω(att.Type.Kind()).Should(Equal(IntegerKind))
				Ω(projected.APIVersions).Should(Equal(mt.APIVersions))
			})
		})

	})

	Context("with media types with view attributes with a cyclical dependency", func() {
		const id = "vnd.application/MT1"
		const typeName = "Mt1"
		var mt2 *MediaTypeDefinition

		BeforeEach(func() {
			InitDesign()
			API("test", func() {})
			mt = MediaType(id, func() {
				TypeName(typeName)
				Attributes(func() {
					Attribute("att", "vnd.application/MT2")
				})
				Links(func() {
					Link("att", "default")
				})
				View("default", func() {
					Attribute("att")
					Attribute("links")
				})
			})
			mt2 = MediaType("vnd.application/MT2", func() {
				TypeName("Mt2")
				Attributes(func() {
					Attribute("att2", mt)
				})
				Links(func() {
					Link("att2", "default")
				})
				View("default", func() {
					Attribute("att2")
					Attribute("links")
				})
			})
			err := dslengine.Run()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		Context("using the default view", func() {
			BeforeEach(func() {
				view = "default"
			})

			It("returns the projected media type with links", func() {
				Ω(prErr).ShouldNot(HaveOccurred())
				Ω(projected).ShouldNot(BeNil())
				Ω(projected.Type).Should(BeAssignableToTypeOf(Object{}))
				Ω(projected.Type.ToObject()).Should(HaveKey("att"))
				l := projected.Type.ToObject()["links"]
				Ω(l.Type.(*UserTypeDefinition).AttributeDefinition).Should(Equal(links.AttributeDefinition))
				Ω(links.Type.ToObject()).Should(HaveKey("att"))
			})
		})
	})
})
