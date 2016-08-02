package apidsl_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MediaType", func() {
	var name string
	var dslFunc func()

	var mt *MediaTypeDefinition

	BeforeEach(func() {
		dslengine.Reset()
		name = ""
		dslFunc = nil
	})

	JustBeforeEach(func() {
		mt = MediaType(name, dslFunc)
		dslengine.Run()
	})

	Context("with no dsl and no identifier", func() {
		It("produces an error", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).Should(HaveOccurred())
		})
	})

	Context("with no dsl", func() {
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

	Context("with a content type", func() {
		const attName = "att"
		const contentType = "application/json"

		BeforeEach(func() {
			name = "application/foo"
			dslFunc = func() {
				ContentType(contentType)
				Attributes(func() {
					Attribute(attName)
				})
				View("default", func() { Attribute(attName) })
			}
		})

		It("sets the content type", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.Validate()).ShouldNot(HaveOccurred())
			Ω(mt.ContentType).Should(Equal(contentType))
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
			Ω(dslengine.Errors).Should(BeEmpty())
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
		dslengine.Reset()
		MediaType(id, dslFunc)
		duplicate = MediaType(id, dslFunc)
	})

	It("produces an error", func() {
		Ω(dslengine.Errors).Should(HaveOccurred())
		Ω(dslengine.Errors.Error()).Should(ContainSubstring("is defined twice"))
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
			Ω(func() { dslengine.Run() }).ShouldNot(Panic())
		})
	})
})

var _ = Describe("CollectionOf", func() {
	Context("used on a global variable", func() {
		var col *MediaTypeDefinition
		BeforeEach(func() {
			dslengine.Reset()
			mt := MediaType("application/vnd.example", func() {
				Attribute("id")
				View("default", func() {
					Attribute("id")
				})
			})
			col = CollectionOf(mt)
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
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
			dslengine.Reset()
			MediaType("application/vnd.example+json", func() {
				Attribute("id")
				View("default", func() {
					Attribute("id")
				})
			})
			col = MediaType("application/vnd.parent+json", func() {
				Attribute("mt", CollectionOf("application/vnd.example"))
				View("default", func() {
					Attribute("mt")
				})
			})
		})

		JustBeforeEach(func() {
			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
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

var _ = Describe("Example", func() {
	Context("defined examples in a media type", func() {
		BeforeEach(func() {
			dslengine.Reset()
			ProjectedMediaTypes = make(MediaTypeRoot)
		})
		It("produces a media type with examples", func() {
			mt := MediaType("application/vnd.example+json", func() {
				Attributes(func() {
					Attribute("test1", String, "test1 desc", func() {
						Example("test1")
					})
					Attribute("test2", String, "test2 desc", func() {
						NoExample()
					})
					Attribute("test3", Integer, "test3 desc", func() {
						Minimum(1)
					})
					Attribute("test4", String, func() {
						Format("email")
						Pattern("@")
					})
					Attribute("test5", Any)

					Attribute("test-failure1", Integer, func() {
						Minimum(0)
						Maximum(0)
					})
				})
				View("default", func() {
					Attribute("test1")
				})
			})

			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())

			Ω(mt).ShouldNot(BeNil())
			attr := mt.Type.ToObject()["test1"]
			Ω(attr.Example).Should(Equal("test1"))
			attr = mt.Type.ToObject()["test2"]
			Ω(attr.Example).Should(Equal("-"))
			attr = mt.Type.ToObject()["test3"]
			Ω(attr.Example).Should(BeNumerically(">=", 1))
			attr = mt.Type.ToObject()["test4"]
			Ω(attr.Example).Should(MatchRegexp(`\w+@`))
			attr = mt.Type.ToObject()["test5"]
			Ω(attr.Example).ShouldNot(BeNil())
			attr = mt.Type.ToObject()["test-failure1"]
			Ω(attr.Example).Should(Equal(0))
		})

		It("produces a media type with HashOf examples", func() {
			ut := Type("example", func() {
				Attribute("test1", Integer)
				Attribute("test2", Any)
			})

			mt := MediaType("application/vnd.example+json", func() {
				Attributes(func() {
					Attribute("test1", HashOf(String, Integer))
					Attribute("test2", HashOf(Any, String))
					Attribute("test3", HashOf(String, Any))
					Attribute("test4", HashOf(Any, Any))

					Attribute("test-with-user-type-1", HashOf(String, ut))
					Attribute("test-with-user-type-2", HashOf(Any, ut))

					Attribute("test-with-array-1", HashOf(String, ArrayOf(Integer)))
					Attribute("test-with-array-2", HashOf(String, ArrayOf(Any)))
					Attribute("test-with-array-3", HashOf(String, ArrayOf(ut)))
					Attribute("test-with-array-4", HashOf(Any, ArrayOf(String)))
					Attribute("test-with-array-5", HashOf(Any, ArrayOf(Any)))
					Attribute("test-with-array-6", HashOf(Any, ArrayOf(ut)))

					Attribute("test-with-example-1", HashOf(String, Boolean), func() {
						Example(map[string]bool{})
					})
					Attribute("test-with-example-2", HashOf(Any, Boolean), func() {
						Example(map[string]int{})
					})
				})
				View("default", func() {
					Attribute("test1")
				})
			})

			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(mt).ShouldNot(BeNil())

			attr := mt.Type.ToObject()["test1"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string]int{}))
			attr = mt.Type.ToObject()["test2"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[interface{}]string{}))
			attr = mt.Type.ToObject()["test3"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string]interface{}{}))
			attr = mt.Type.ToObject()["test4"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[interface{}]interface{}{}))

			attr = mt.Type.ToObject()["test-with-user-type-1"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string]map[string]interface{}{}))
			for _, utattr := range attr.Example.(map[string]map[string]interface{}) {
				Expect(utattr).Should(HaveKey("test1"))
				Expect(utattr).Should(HaveKey("test2"))
				Expect(utattr["test1"]).Should(BeAssignableToTypeOf(int(0)))
			}
			attr = mt.Type.ToObject()["test-with-user-type-2"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[interface{}]map[string]interface{}{}))
			for _, utattr := range attr.Example.(map[interface{}]map[string]interface{}) {
				Expect(utattr).Should(HaveKey("test1"))
				Expect(utattr).Should(HaveKey("test2"))
				Expect(utattr["test1"]).Should(BeAssignableToTypeOf(int(0)))
			}

			attr = mt.Type.ToObject()["test-with-array-1"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string][]int{}))
			attr = mt.Type.ToObject()["test-with-array-2"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string][]interface{}{}))
			attr = mt.Type.ToObject()["test-with-array-3"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string][]map[string]interface{}{}))
			attr = mt.Type.ToObject()["test-with-array-4"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[interface{}][]string{}))
			attr = mt.Type.ToObject()["test-with-array-5"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[interface{}][]interface{}{}))
			attr = mt.Type.ToObject()["test-with-array-6"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[interface{}][]map[string]interface{}{}))

			attr = mt.Type.ToObject()["test-with-example-1"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string]bool{}))
			attr = mt.Type.ToObject()["test-with-example-2"]
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string]int{}))
		})

		It("produces a media type with examples in cyclical dependencies", func() {
			mt := MediaType("vnd.application/foo", func() {
				Attributes(func() {
					Attribute("foo", "vnd.application/bar")
					Attribute("others", Integer, func() {
						Minimum(3)
						Maximum(3)
					})
				})
				View("default", func() {
					Attribute("foo")
					Attribute("others")
				})
			})

			mt2 := MediaType("vnd.application/bar", func() {
				Attributes(func() {
					Attribute("bar", mt)
					Attribute("others", Integer, func() {
						Minimum(1)
						Maximum(2)
					})
				})
				View("default", func() {
					Attribute("bar")
					Attribute("others")
				})
			})

			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())

			Ω(mt).ShouldNot(BeNil())
			attr := mt.Type.ToObject()["foo"]
			Ω(attr.Example).ShouldNot(BeNil())
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string]interface{}{}))
			attrChild := attr.Example.(map[string]interface{})
			Ω(attrChild).Should(HaveKey("bar"))
			Ω(attrChild["others"]).Should(BeNumerically(">=", 1))
			Ω(attrChild["others"]).Should(BeNumerically("<=", 2))
			attr = mt.Type.ToObject()["others"]
			Ω(attr.Example).Should(Equal(3))

			Ω(mt2).ShouldNot(BeNil())
			attr = mt2.Type.ToObject()["bar"]
			Ω(attr.Example).ShouldNot(BeNil())
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string]interface{}{}))
			attrChild = attr.Example.(map[string]interface{})
			Ω(attrChild).Should(HaveKey("foo"))
			Ω(attrChild["others"]).Should(Equal(3))
			attr = mt2.Type.ToObject()["others"]
			Ω(attr.Example).Should(BeNumerically(">=", 1))
			Ω(attr.Example).Should(BeNumerically("<=", 2))
		})

		It("produces media type examples from the linked media type", func() {
			mt := MediaType("application/vnd.example+json", func() {
				Attributes(func() {
					Attribute("test1", String, "test1 desc", func() {
						Example("test1")
					})
					Attribute("test2", String, "test2 desc", func() {
						NoExample()
					})
					Attribute("test3", Integer, "test3 desc", func() {
						Minimum(1)
					})
				})
				View("default", func() {
					Attribute("test1")
					Attribute("test2")
					Attribute("test3")
				})
			})

			pmt := MediaType("application/vnd.example.parent+json", func() {
				Attributes(func() {
					Attribute("test1", String, "test1 desc", func() {
						Example("test1")
					})
					Attribute("test2", String, "test2 desc", func() {
						NoExample()
					})
					Attribute("test3", Integer, "test3 desc", func() {
						Minimum(1)
					})
					Attribute("test4", mt, "test4 desc")
				})
				View("default", func() {
					Attribute("test1")
					Attribute("test2")
					Attribute("test3")
					Attribute("test4")
				})
			})

			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())

			Ω(mt).ShouldNot(BeNil())
			attr := mt.Type.ToObject()["test1"]
			Ω(attr.Example).Should(Equal("test1"))
			attr = mt.Type.ToObject()["test2"]
			Ω(attr.Example).Should(Equal("-"))
			attr = mt.Type.ToObject()["test3"]
			Ω(attr.Example).Should(BeNumerically(">=", 1))

			Ω(pmt).ShouldNot(BeNil())
			attr = pmt.Type.ToObject()["test1"]
			Ω(attr.Example).Should(Equal("test1"))
			attr = pmt.Type.ToObject()["test2"]
			Ω(attr.Example).Should(Equal("-"))
			attr = pmt.Type.ToObject()["test3"]
			Ω(attr.Example).Should(BeNumerically(">=", 1))
			attr = pmt.Type.ToObject()["test4"]
			Ω(attr.Example).ShouldNot(BeNil())
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string]interface{}{}))
			attrChild := attr.Example.(map[string]interface{})
			Ω(attrChild["test1"]).Should(Equal("test1"))
			Ω(attrChild["test2"]).Should(Equal("-"))
			Ω(attrChild["test3"]).Should(BeNumerically(">=", 1))
		})

		It("produces media type examples from the linked media type collection with custom examples", func() {
			mt := MediaType("application/vnd.example+json", func() {
				Attributes(func() {
					Attribute("test1", String, "test1 desc", func() {
						Example("test1")
					})
					Attribute("test2", String, "test2 desc", func() {
						NoExample()
					})
					Attribute("test3", Integer, "test3 desc", func() {
						Minimum(1)
					})
				})
				View("default", func() {
					Attribute("test1")
					Attribute("test2")
					Attribute("test3")
				})
			})

			pmt := MediaType("application/vnd.example.parent+json", func() {
				Attributes(func() {
					Attribute("test1", String, "test1 desc", func() {
						Example("test1")
					})
					Attribute("test2", String, "test2 desc", func() {
						NoExample()
					})
					Attribute("test3", String, "test3 desc", func() {
						Pattern("^1$")
					})
					Attribute("test4", CollectionOf(mt), "test4 desc")
				})
				View("default", func() {
					Attribute("test1")
					Attribute("test2")
					Attribute("test3")
					Attribute("test4")
				})
			})

			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())

			Ω(mt).ShouldNot(BeNil())
			attr := mt.Type.ToObject()["test1"]
			Ω(attr.Example).Should(Equal("test1"))
			attr = mt.Type.ToObject()["test2"]
			Ω(attr.Example).Should(Equal("-"))
			attr = mt.Type.ToObject()["test3"]
			Ω(attr.Example).Should(BeNumerically(">=", 1))

			Ω(pmt).ShouldNot(BeNil())
			attr = pmt.Type.ToObject()["test1"]
			Ω(attr.Example).Should(Equal("test1"))
			attr = pmt.Type.ToObject()["test2"]
			Ω(attr.Example).Should(Equal("-"))
			attr = pmt.Type.ToObject()["test3"]
			Ω(attr.Example).Should(Equal("1"))
			attr = pmt.Type.ToObject()["test4"]
			Ω(attr.Example).ShouldNot(BeNil())
			Expect(attr.Example).Should(BeAssignableToTypeOf([]map[string]interface{}{}))
			attrChildren := attr.Example.([]map[string]interface{})
			Ω(attrChildren).Should(HaveLen(1))
			Ω(attrChildren[0]).Should(BeAssignableToTypeOf(map[string]interface{}{}))
			Ω(attrChildren[0]["test1"]).Should(Equal("test1"))
			Ω(attrChildren[0]["test2"]).Should(Equal("-"))
			Ω(attrChildren[0]["test3"]).Should(BeNumerically(">=", 1))
		})

		It("produces media type examples from the linked media type without custom examples", func() {
			mt := MediaType("application/vnd.example.child+json", func() {
				Attributes(func() {
					Attribute("test1", String, "test1 desc")
				})
				View("default", func() {
					Attribute("test1")
				})
			})

			pmt := MediaType("application/vnd.example.parent+json", func() {
				Attributes(func() {
					Attribute("test1", String, "test1 desc", func() {
						Example("test1")
					})
					Attribute("test2", String, "test2 desc", func() {
						NoExample()
					})
					Attribute("test3", mt, "test3 desc")
				})
				View("default", func() {
					Attribute("test1")
					Attribute("test2")
					Attribute("test3")
				})
			})

			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			Ω(mt).ShouldNot(BeNil())
			attr := mt.Type.ToObject()["test1"]
			cexample := attr.Example
			Ω(cexample).ShouldNot(BeEmpty())

			Ω(pmt).ShouldNot(BeNil())
			attr = pmt.Type.ToObject()["test1"]
			Ω(attr.Example).Should(Equal("test1"))
			attr = pmt.Type.ToObject()["test2"]
			Ω(attr.Example).Should(Equal("-"))
			attr = pmt.Type.ToObject()["test3"]
			Ω(attr.Example).ShouldNot(BeNil())
			Expect(attr.Example).Should(BeAssignableToTypeOf(map[string]interface{}{}))
		})

		It("produces media type examples from the linked media type collection without custom examples", func() {
			mt := MediaType("application/vnd.example.child+json", func() {
				Attributes(func() {
					Attribute("test1", String, "test1 desc")
				})
				View("default", func() {
					Attribute("test1")
				})
			})

			pmt := MediaType("application/vnd.example.parent+json", func() {
				Attributes(func() {
					Attribute("test1", String, "test1 desc", func() {
						Example("test1")
					})
					Attribute("test2", String, "test2 desc", func() {
						NoExample()
					})
					Attribute("test3", CollectionOf(mt), "test3 desc")
				})
				View("default", func() {
					Attribute("test1")
				})
			})

			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())

			Ω(mt).ShouldNot(BeNil())
			attr := mt.Type.ToObject()["test1"]
			cexample := attr.Example
			Ω(cexample).ShouldNot(BeEmpty())

			Ω(pmt).ShouldNot(BeNil())
			attr = pmt.Type.ToObject()["test1"]
			Ω(attr.Example).Should(Equal("test1"))
			attr = pmt.Type.ToObject()["test2"]
			Ω(attr.Example).Should(Equal("-"))
			attr = pmt.Type.ToObject()["test3"]
			Ω(attr.Example).ShouldNot(BeNil())
			Expect(attr.Example).Should(BeAssignableToTypeOf([]map[string]interface{}{}))
			attrChildren := attr.Example.([]map[string]interface{})
			Ω(len(attrChildren)).Should(BeNumerically(">=", 1))
		})

		It("produces a media type with appropriate MinLength and MaxLength examples", func() {
			ut := Type("example", func() {
				Attribute("test1", Integer, func() {
					Minimum(-200)
					Maximum(-100)
				})
			})

			mt := MediaType("application/vnd.example+json", func() {
				Attributes(func() {
					Attribute("test1", ArrayOf(Any), func() {
						MinLength(0)
						MaxLength(10)
					})
					Attribute("test2", ArrayOf(Any), func() {
						MinLength(1000)
						MaxLength(2000)
					})
					Attribute("test3", ArrayOf(Any), func() {
						MinLength(1000)
						MaxLength(1000)
					})

					Attribute("test-failure1", ArrayOf(ut), func() {
						MinLength(0)
						MaxLength(0)
					})
				})
				View("default", func() {
					Attribute("test1")
				})
			})

			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())

			Ω(mt).ShouldNot(BeNil())
			attr := mt.Type.ToObject()["test1"]
			Ω(attr.Example).Should(BeAssignableToTypeOf([]interface{}{}))
			Ω(len(attr.Example.([]interface{}))).Should(BeNumerically("<=", 10))
			attr = mt.Type.ToObject()["test2"]
			Ω(attr.Example).Should(BeAssignableToTypeOf([]interface{}{}))
			Ω(attr.Example.([]interface{})).Should(HaveLen(10))
			attr = mt.Type.ToObject()["test3"]
			Ω(attr.Example).Should(BeAssignableToTypeOf([]interface{}{}))
			Ω(attr.Example.([]interface{})).Should(HaveLen(10))
			attr = mt.Type.ToObject()["test-failure1"]
			Ω(attr.Example).Should(BeNil())
		})
	})
})
