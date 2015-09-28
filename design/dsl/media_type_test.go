package dsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Describe("MediaType", func() {
	var name string
	var dsl func()

	var mt *MediaTypeDefinition

	BeforeEach(func() {
		Design = nil
		Errors = nil
		name = ""
		dsl = nil
	})

	JustBeforeEach(func() {
		mt = MediaType(name, dsl)
		RunDSL()
		Ω(Errors).ShouldNot(HaveOccurred())
	})

	Context("with no DSL and no identifier", func() {
		It("produces a valid media type definition", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).ShouldNot(HaveOccurred())
		})
	})

	Context("with no DSL", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces a valid media type definition", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).ShouldNot(HaveOccurred())
		})
	})

	Context("with a description", func() {
		const description = "desc"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Description(description)
			}
		})

		It("sets the description", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).ShouldNot(HaveOccurred())
			Ω(mt.Description).Should(Equal(description))
		})
	})

	Context("with attributes", func() {
		const attName = "att"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Attributes(func() {
					Attribute(attName)
				})
			}
		})

		It("sets the attributes", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).ShouldNot(HaveOccurred())
			Ω(mt.AttributeDefinition).ShouldNot(BeNil())
			Ω(mt.Type).Should(BeAssignableToTypeOf(Object{}))
			o := mt.Type.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(attName))
		})
	})

	Context("with links", func() {
		const linkName = "link"
		var link1Name, link2Name, link3Name, link4Name string
		var link2View, link4View string
		var link3Type, link4Type *MediaTypeDefinition

		BeforeEach(func() {
			name = "foo"
			link1Name = "l1"
			link2Name = "l2"
			link2View = "l2v"
			link3Name = "l3"
			link3Type = &MediaTypeDefinition{Identifier: "l3t"}
			link4Name = "l4"
			link4View = "l4v"
			link4Type = &MediaTypeDefinition{Identifier: "l4t"}
			dsl = func() {
				Attributes(func() {
					Links(func() {
						Link(link1Name)
						Link(link2Name, link2View)
						Link(link3Name, link3Type)
						Link(link4Name, link4View, link4Type)
					})
				})
			}
		})

		It("sets the links", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).ShouldNot(HaveOccurred())
			Ω(mt.Links).ShouldNot(BeNil())
			Ω(mt.Links).Should(HaveLen(4))
			Ω(mt.Links).Should(HaveKey(link1Name))
			Ω(mt.Links[link1Name].Name).Should(Equal(link1Name))
			Ω(mt.Links[link1Name].MediaType).Should(BeNil())
			Ω(mt.Links[link1Name].View).Should(Equal("default"))
			Ω(mt.Links[link1Name].Parent).Should(Equal(mt))
			Ω(mt.Links[link2Name].Name).Should(Equal(link2Name))
			Ω(mt.Links[link2Name].MediaType).Should(BeNil())
			Ω(mt.Links[link2Name].View).Should(Equal(link2View))
			Ω(mt.Links[link2Name].Parent).Should(Equal(mt))
			Ω(mt.Links[link3Name].Name).Should(Equal(link3Name))
			Ω(mt.Links[link3Name].MediaType).Should(Equal(link3Type))
			Ω(mt.Links[link3Name].View).Should(Equal("default"))
			Ω(mt.Links[link3Name].Parent).Should(Equal(mt))
			Ω(mt.Links[link4Name].Name).Should(Equal(link4Name))
			Ω(mt.Links[link4Name].MediaType).Should(Equal(link4Type))
			Ω(mt.Links[link4Name].View).Should(Equal(link4View))
			Ω(mt.Links[link4Name].Parent).Should(Equal(mt))
		})
	})

	Context("with views", func() {
		const viewName = "view"
		const viewAtt = "att"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				View(viewName, func() {
					Attribute(viewAtt)
				})
			}
		})

		It("sets the views", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).ShouldNot(HaveOccurred())
			Ω(mt.Views).ShouldNot(BeNil())
			Ω(mt.Views).Should(HaveLen(1))
			Ω(mt.Views).Should(HaveKey(viewName))
			v := mt.Views[viewName]
			Ω(v.Name).Should(Equal(viewName))
			Ω(v.Parent).Should(Equal(mt))
			Ω(v.AttributeDefinition).ShouldNot(BeNil())
			Ω(v.AttributeDefinition.Type).Should(BeAssignableToTypeOf(Object{}))
			o := v.AttributeDefinition.Type.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(viewAtt))
			Ω(o[viewAtt]).ShouldNot(BeNil())
			Ω(o[viewAtt].Type).Should(Equal(String))
		})
	})
})
