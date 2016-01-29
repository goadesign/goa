package dsl_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/engine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MediaType", func() {
	var name string
	var dslFunc func()

	var mt *MediaTypeDefinition

	BeforeEach(func() {
		InitDesign()
		engine.Errors = nil
		name = ""
		dslFunc = nil
	})

	JustBeforeEach(func() {
		mt = MediaType(name, dslFunc)
		engine.RunDSL()
		Ω(engine.Errors).ShouldNot(HaveOccurred())
	})

	Context("with no DSL and no identifier", func() {
		It("produces an error", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).Should(HaveOccurred())
		})
	})

	Context("with no DSL", func() {
		BeforeEach(func() {
			name = "application/foo"
		})

		It("produces an error", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).Should(HaveOccurred())
		})
	})

	Context("with attributes", func() {
		const attName = "att"

		BeforeEach(func() {
			name = "application/foo"
			dslFunc = func() {
				Attributes(func() {
					Attribute(attName)
				})
				View("default", func() { Attribute(attName) })
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

	Context("with a description", func() {
		const description = "desc"

		BeforeEach(func() {
			name = "application/foo"
			dslFunc = func() {
				Description(description)
				Attributes(func() {
					Attribute("attName")
				})
				View("default", func() { Attribute("attName") })
			}
		})

		It("sets the description", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).ShouldNot(HaveOccurred())
			Ω(mt.Description).Should(Equal(description))
		})
	})

	Context("with links", func() {
		const linkName = "link"
		var link1Name, link2Name string
		var link2View string
		var mt1, mt2 *MediaTypeDefinition

		BeforeEach(func() {
			name = "foo"
			link1Name = "l1"
			link2Name = "l2"
			link2View = "l2v"
			mt1 = NewMediaTypeDefinition("application/mt1", "application/mt1", func() {
				Attributes(func() {
					Attribute("foo")
				})
				View("default", func() {
					Attribute("foo")
				})
				View("link", func() {
					Attribute("foo")
				})
			})
			mt2 = NewMediaTypeDefinition("application/mt2", "application/mt2", func() {
				Attributes(func() {
					Attribute("foo")
				})
				View("l2v", func() {
					Attribute("foo")
				})
				View("default", func() {
					Attribute("foo")
				})
			})
			Design.MediaTypes = make(map[string]*MediaTypeDefinition)
			Design.MediaTypes["application/mt1"] = mt1
			Design.MediaTypes["application/mt2"] = mt2
			dslFunc = func() {
				Attributes(func() {
					Attributes(func() {
						Attribute(link1Name, mt1)
						Attribute(link2Name, mt2)
					})
					Links(func() {
						Link(link1Name)
						Link(link2Name, link2View)
					})
					View("default", func() {
						Attribute(link1Name)
						Attribute(link2Name)
					})
				})
			}
		})

		It("sets the links", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(engine.Errors).Should(BeEmpty())
			Ω(mt.Validate()).ShouldNot(HaveOccurred())
			Ω(mt.Links).ShouldNot(BeNil())
			Ω(mt.Links).Should(HaveLen(2))
			Ω(mt.Links).Should(HaveKey(link1Name))
			Ω(mt.Links[link1Name].Name).Should(Equal(link1Name))
			Ω(mt.Links[link1Name].View).Should(Equal("link"))
			Ω(mt.Links[link1Name].Parent).Should(Equal(mt))
			Ω(mt.Links[link2Name].Name).Should(Equal(link2Name))
			Ω(mt.Links[link2Name].View).Should(Equal(link2View))
			Ω(mt.Links[link2Name].Parent).Should(Equal(mt))
		})
	})

	Context("with views", func() {
		const viewName = "view"
		const viewAtt = "att"

		BeforeEach(func() {
			name = "application/foo"
			dslFunc = func() {
				Attributes(func() {
					Attribute(viewAtt)
				})
				View(viewName, func() {
					Attribute(viewAtt)
				})
				View("default", func() {
					Attribute(viewAtt)
				})
			}
		})

		It("sets the views", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).ShouldNot(HaveOccurred())
			Ω(mt.Views).ShouldNot(BeNil())
			Ω(mt.Views).Should(HaveLen(2))
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

var _ = Describe("Duplicate media types", func() {
	var mt *MediaTypeDefinition
	var duplicate *MediaTypeDefinition
	const id = "application/foo"
	const attName = "bar"
	var dslFunc = func() {
		Attributes(func() {
			Attribute(attName)
		})
		View("default", func() { Attribute(attName) })
	}

	BeforeEach(func() {
		InitDesign()
		engine.Errors = nil
		mt = MediaType(id, dslFunc)
		Ω(engine.Errors).ShouldNot(HaveOccurred())
		duplicate = MediaType(id, dslFunc)
	})

	It("produces an error", func() {
		Ω(engine.Errors).Should(HaveOccurred())
	})

	Context("with a response definition using the duplicate", func() {
		BeforeEach(func() {
			Resource("foo", func() {
				Action("show", func() {
					Routing(GET(""))
					Response(OK, func() {
						Media(duplicate)
					})
				})
			})
		})

		It("does not panic", func() {
			Ω(func() { engine.RunDSL() }).ShouldNot(Panic())
		})
	})
})

var _ = Describe("CollectionOf", func() {
	Context("used on a global variable", func() {
		var col *MediaTypeDefinition
		BeforeEach(func() {
			InitDesign()
			mt := MediaType("application/vnd.example", func() { Attribute("id") })
			engine.Errors = nil
			col = CollectionOf(mt)
			Ω(engine.Errors).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			engine.RunDSL()
			Ω(engine.Errors).ShouldNot(HaveOccurred())
		})

		It("produces a media type", func() {
			Ω(col).ShouldNot(BeNil())
			Ω(col.Identifier).ShouldNot(BeEmpty())
			Ω(col.TypeName).ShouldNot(BeEmpty())
			Ω(Design.MediaTypes).Should(HaveKey(col.Identifier))
		})
	})

	Context("defined with the media type identifier", func() {
		var col *MediaTypeDefinition
		BeforeEach(func() {
			InitDesign()
			MediaType("application/vnd.example+json", func() { Attribute("id") })
			col = MediaType("application/vnd.parent+json", func() { Attribute("mt", CollectionOf("application/vnd.example")) })
		})

		JustBeforeEach(func() {
			engine.RunDSL()
			Ω(engine.Errors).ShouldNot(HaveOccurred())
		})

		It("produces a media type", func() {
			Ω(col).ShouldNot(BeNil())
			Ω(col.Identifier).Should(Equal("application/vnd.parent+json"))
			Ω(col.TypeName).Should(Equal("Parent"))
			Ω(col.Type).ShouldNot(BeNil())
			Ω(col.Type.ToObject()).ShouldNot(BeNil())
			Ω(col.Type.ToObject()).Should(HaveKey("mt"))
			mt := col.Type.ToObject()["mt"]
			Ω(mt.Type).ShouldNot(BeNil())
			Ω(mt.Type).Should(BeAssignableToTypeOf(&MediaTypeDefinition{}))
			Ω(mt.Type.Name()).Should(Equal("array"))
			et := mt.Type.ToArray().ElemType
			Ω(et).ShouldNot(BeNil())
			Ω(et.Type).Should(BeAssignableToTypeOf(&MediaTypeDefinition{}))
			Ω(et.Type.(*MediaTypeDefinition).Identifier).Should(Equal("application/vnd.example+json"))
		})
	})
})
