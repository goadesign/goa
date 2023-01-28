package apidsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kyokomi/goa-v1/design"
	"github.com/kyokomi/goa-v1/design/apidsl"
	"github.com/kyokomi/goa-v1/dslengine"
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
		mt = apidsl.MediaType(name, dslFunc)
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
				apidsl.Attributes(func() {
					apidsl.Attribute(attName)
				})
				apidsl.View("default", func() { apidsl.Attribute(attName) })
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
				apidsl.ContentType(contentType)
				apidsl.Attributes(func() {
					apidsl.Attribute(attName)
				})
				apidsl.View("default", func() { apidsl.Attribute(attName) })
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
				apidsl.Description(description)
				apidsl.Attributes(func() {
					apidsl.Attribute("attName")
				})
				apidsl.View("default", func() { apidsl.Attribute("attName") })
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
				apidsl.Attributes(func() {
					apidsl.Attribute("foo")
				})
				apidsl.View("default", func() {
					apidsl.Attribute("foo")
				})
				apidsl.View("link", func() {
					apidsl.Attribute("foo")
				})
			})
			mt2 = NewMediaTypeDefinition("application/mt2", "application/mt2", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("foo")
				})
				apidsl.View("l2v", func() {
					apidsl.Attribute("foo")
				})
				apidsl.View("default", func() {
					apidsl.Attribute("foo")
				})
			})
			Design.MediaTypes = make(map[string]*MediaTypeDefinition)
			Design.MediaTypes["application/mt1"] = mt1
			Design.MediaTypes["application/mt2"] = mt2
			dslFunc = func() {
				apidsl.Attributes(func() {
					apidsl.Attributes(func() {
						apidsl.Attribute(link1Name, mt1)
						apidsl.Attribute(link2Name, mt2)
					})
					apidsl.Links(func() {
						apidsl.Link(link1Name)
						apidsl.Link(link2Name, link2View)
					})
					apidsl.View("default", func() {
						apidsl.Attribute(link1Name)
						apidsl.Attribute(link2Name)
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
				apidsl.Attributes(func() {
					apidsl.Attribute(viewAtt)
				})
				apidsl.View(viewName, func() {
					apidsl.Attribute(viewAtt)
				})
				apidsl.View("default", func() {
					apidsl.Attribute(viewAtt)
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
		apidsl.Attributes(func() {
			apidsl.Attribute(attName)
		})
		apidsl.View("default", func() { apidsl.Attribute(attName) })
	}

	BeforeEach(func() {
		dslengine.Reset()
		apidsl.MediaType(id, dslFunc)
		duplicate = apidsl.MediaType(id, dslFunc)
	})

	It("produces an error", func() {
		Ω(dslengine.Errors).Should(HaveOccurred())
		Ω(dslengine.Errors.Error()).Should(ContainSubstring("is defined twice"))
	})

	Context("with a response definition using the duplicate", func() {
		BeforeEach(func() {
			apidsl.Resource("foo", func() {
				apidsl.Action("show", func() {
					apidsl.Routing(apidsl.GET(""))
					apidsl.Response(OK, func() {
						apidsl.Media(duplicate)
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
			mt := apidsl.MediaType("application/vnd.example", func() {
				apidsl.Attribute("id")
				apidsl.View("default", func() {
					apidsl.Attribute("id")
				})
			})
			col = apidsl.CollectionOf(mt)
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		It("produces a media type", func() {
			Ω(col).ShouldNot(BeNil())
			Ω(col.Identifier).Should(Equal("application/vnd.example; type=collection"))
			Ω(col.TypeName).ShouldNot(BeEmpty())
			Ω(Design.MediaTypes).Should(HaveKey(col.Identifier))
		})
	})

	Context("defined with a collection identifier", func() {
		var col *MediaTypeDefinition
		BeforeEach(func() {
			dslengine.Reset()
			mt := apidsl.MediaType("application/vnd.example", func() {
				apidsl.Attribute("id")
				apidsl.View("default", func() {
					apidsl.Attribute("id")
				})
			})
			col = apidsl.CollectionOf(mt, "application/vnd.examples")
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		It("produces a media type", func() {
			Ω(col).ShouldNot(BeNil())
			Ω(col.Identifier).Should(Equal("application/vnd.examples"))
			Ω(col.TypeName).ShouldNot(BeEmpty())
			Ω(Design.MediaTypes).Should(HaveKey(col.Identifier))
		})
	})

	Context("defined with the media type identifier", func() {
		var col *MediaTypeDefinition
		BeforeEach(func() {
			dslengine.Reset()
			apidsl.MediaType("application/vnd.example+json", func() {
				apidsl.Attribute("id")
				apidsl.View("default", func() {
					apidsl.Attribute("id")
				})
			})
			col = apidsl.MediaType("application/vnd.parent+json", func() {
				apidsl.Attribute("mt", apidsl.CollectionOf("application/vnd.example"))
				apidsl.View("default", func() {
					apidsl.Attribute("mt")
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
			mt := apidsl.MediaType("application/vnd.example+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", String, "test1 desc", func() {
						apidsl.Example("test1")
					})
					apidsl.Attribute("test2", String, "test2 desc", func() {
						apidsl.NoExample()
					})
					apidsl.Attribute("test3", Integer, "test3 desc", func() {
						apidsl.Minimum(1)
					})
					apidsl.Attribute("test4", String, func() {
						apidsl.Format("email")
						apidsl.Pattern("@")
					})
					apidsl.Attribute("test5", Any)

					apidsl.Attribute("test-failure1", Integer, func() {
						apidsl.Minimum(0)
						apidsl.Maximum(0)
					})
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
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
			ut := apidsl.Type("example", func() {
				apidsl.Attribute("test1", Integer)
				apidsl.Attribute("test2", Any)
			})

			mt := apidsl.MediaType("application/vnd.example+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", apidsl.HashOf(String, Integer))
					apidsl.Attribute("test2", apidsl.HashOf(Any, String))
					apidsl.Attribute("test3", apidsl.HashOf(String, Any))
					apidsl.Attribute("test4", apidsl.HashOf(Any, Any))

					apidsl.Attribute("test-with-user-type-1", apidsl.HashOf(String, ut))
					apidsl.Attribute("test-with-user-type-2", apidsl.HashOf(Any, ut))

					apidsl.Attribute("test-with-array-1", apidsl.HashOf(String, apidsl.ArrayOf(Integer)))
					apidsl.Attribute("test-with-array-2", apidsl.HashOf(String, apidsl.ArrayOf(Any)))
					apidsl.Attribute("test-with-array-3", apidsl.HashOf(String, apidsl.ArrayOf(ut)))
					apidsl.Attribute("test-with-array-4", apidsl.HashOf(Any, apidsl.ArrayOf(String)))
					apidsl.Attribute("test-with-array-5", apidsl.HashOf(Any, apidsl.ArrayOf(Any)))
					apidsl.Attribute("test-with-array-6", apidsl.HashOf(Any, apidsl.ArrayOf(ut)))

					apidsl.Attribute("test-with-example-1", apidsl.HashOf(String, Boolean), func() {
						apidsl.Example(map[string]bool{})
					})
					apidsl.Attribute("test-with-example-2", apidsl.HashOf(Any, Boolean), func() {
						apidsl.Example(map[string]int{})
					})
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
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
			mt := apidsl.MediaType("vnd.application/foo", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("foo", "vnd.application/bar")
					apidsl.Attribute("others", Integer, func() {
						apidsl.Minimum(3)
						apidsl.Maximum(3)
					})
				})
				apidsl.View("default", func() {
					apidsl.Attribute("foo")
					apidsl.Attribute("others")
				})
			})

			mt2 := apidsl.MediaType("vnd.application/bar", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("bar", mt)
					apidsl.Attribute("others", Integer, func() {
						apidsl.Minimum(1)
						apidsl.Maximum(2)
					})
				})
				apidsl.View("default", func() {
					apidsl.Attribute("bar")
					apidsl.Attribute("others")
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
			mt := apidsl.MediaType("application/vnd.example+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", String, "test1 desc", func() {
						apidsl.Example("test1")
					})
					apidsl.Attribute("test2", String, "test2 desc", func() {
						apidsl.NoExample()
					})
					apidsl.Attribute("test3", Integer, "test3 desc", func() {
						apidsl.Minimum(1)
					})
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
					apidsl.Attribute("test2")
					apidsl.Attribute("test3")
				})
			})

			pmt := apidsl.MediaType("application/vnd.example.parent+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", String, "test1 desc", func() {
						apidsl.Example("test1")
					})
					apidsl.Attribute("test2", String, "test2 desc", func() {
						apidsl.NoExample()
					})
					apidsl.Attribute("test3", Integer, "test3 desc", func() {
						apidsl.Minimum(1)
					})
					apidsl.Attribute("test4", mt, "test4 desc")
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
					apidsl.Attribute("test2")
					apidsl.Attribute("test3")
					apidsl.Attribute("test4")
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
			mt := apidsl.MediaType("application/vnd.example+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", String, "test1 desc", func() {
						apidsl.Example("test1")
					})
					apidsl.Attribute("test2", String, "test2 desc", func() {
						apidsl.NoExample()
					})
					apidsl.Attribute("test3", Integer, "test3 desc", func() {
						apidsl.Minimum(1)
					})
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
					apidsl.Attribute("test2")
					apidsl.Attribute("test3")
				})
			})

			pmt := apidsl.MediaType("application/vnd.example.parent+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", String, "test1 desc", func() {
						apidsl.Example("test1")
					})
					apidsl.Attribute("test2", String, "test2 desc", func() {
						apidsl.NoExample()
					})
					apidsl.Attribute("test3", String, "test3 desc", func() {
						apidsl.Pattern("^1$")
					})
					apidsl.Attribute("test4", apidsl.CollectionOf(mt), "test4 desc")
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
					apidsl.Attribute("test2")
					apidsl.Attribute("test3")
					apidsl.Attribute("test4")
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
			mt := apidsl.MediaType("application/vnd.example.child+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", String, "test1 desc")
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
				})
			})

			pmt := apidsl.MediaType("application/vnd.example.parent+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", String, "test1 desc", func() {
						apidsl.Example("test1")
					})
					apidsl.Attribute("test2", String, "test2 desc", func() {
						apidsl.NoExample()
					})
					apidsl.Attribute("test3", mt, "test3 desc")
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
					apidsl.Attribute("test2")
					apidsl.Attribute("test3")
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
			mt := apidsl.MediaType("application/vnd.example.child+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", String, "test1 desc")
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
				})
			})

			pmt := apidsl.MediaType("application/vnd.example.parent+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", String, "test1 desc", func() {
						apidsl.Example("test1")
					})
					apidsl.Attribute("test2", String, "test2 desc", func() {
						apidsl.NoExample()
					})
					apidsl.Attribute("test3", apidsl.CollectionOf(mt), "test3 desc")
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
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
			ut := apidsl.Type("example", func() {
				apidsl.Attribute("test1", Integer, func() {
					apidsl.Minimum(-200)
					apidsl.Maximum(-100)
				})
			})

			mt := apidsl.MediaType("application/vnd.example+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("test1", apidsl.ArrayOf(Any), func() {
						apidsl.MinLength(0)
						apidsl.MaxLength(10)
					})
					apidsl.Attribute("test2", apidsl.ArrayOf(Any), func() {
						apidsl.MinLength(1000)
						apidsl.MaxLength(2000)
					})
					apidsl.Attribute("test3", apidsl.ArrayOf(Any), func() {
						apidsl.MinLength(1000)
						apidsl.MaxLength(1000)
					})

					apidsl.Attribute("test-failure1", apidsl.ArrayOf(ut), func() {
						apidsl.MinLength(0)
						apidsl.MaxLength(0)
					})
				})
				apidsl.View("default", func() {
					apidsl.Attribute("test1")
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
