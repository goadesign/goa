package apidsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kyokomi/goa-v1/design"
	"github.com/kyokomi/goa-v1/design/apidsl"
	"github.com/kyokomi/goa-v1/dslengine"
)

var _ = Describe("Type", func() {
	var name string
	var dsl func()

	var ut *UserTypeDefinition

	BeforeEach(func() {
		dslengine.Reset()
		name = ""
		dsl = nil
	})

	JustBeforeEach(func() {
		apidsl.Type(name, dsl)
		dslengine.Run()
		ut, _ = Design.Types[name]
	})

	Context("with no dsl and no name", func() {
		It("produces an invalid type definition", func() {
			Ω(ut).ShouldNot(BeNil())
			Ω(ut.Validate("test", Design)).Should(HaveOccurred())
		})
	})

	Context("with no dsl", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces a valid type definition", func() {
			Ω(ut).ShouldNot(BeNil())
			Ω(ut.Validate("test", Design)).ShouldNot(HaveOccurred())
		})
	})

	Context("with attributes", func() {
		const attName = "att"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				apidsl.Attribute(attName)
			}
		})

		It("sets the attributes", func() {
			Ω(ut).ShouldNot(BeNil())
			Ω(ut.Validate("test", Design)).ShouldNot(HaveOccurred())
			Ω(ut.AttributeDefinition).ShouldNot(BeNil())
			Ω(ut.Type).Should(BeAssignableToTypeOf(Object{}))
			o := ut.Type.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(attName))
		})
	})

	Context("with a name and uuid datatype", func() {
		const attName = "att"
		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				apidsl.Attribute(attName, UUID)
			}
		})

		It("produces an attribute of date type", func() {
			Ω(ut).ShouldNot(BeNil())
			Ω(ut.Validate("test", Design)).ShouldNot(HaveOccurred())
			Ω(ut.AttributeDefinition).ShouldNot(BeNil())
			Ω(ut.Type).Should(BeAssignableToTypeOf(Object{}))
			o := ut.Type.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(attName))
			Ω(o[attName].Type).Should(Equal(UUID))
		})
	})

	Context("with a name and date datatype", func() {
		const attName = "att"
		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				apidsl.Attribute(attName, DateTime)
			}
		})

		It("produces an attribute of date type", func() {
			Ω(ut).ShouldNot(BeNil())
			Ω(ut.Validate("test", Design)).ShouldNot(HaveOccurred())
			Ω(ut.AttributeDefinition).ShouldNot(BeNil())
			Ω(ut.Type).Should(BeAssignableToTypeOf(Object{}))
			o := ut.Type.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(attName))
			Ω(o[attName].Type).Should(Equal(DateTime))
		})
	})
})

var _ = Describe("ArrayOf", func() {
	Context("used on a global variable", func() {
		var (
			ut *UserTypeDefinition
			ar *Array
		)
		BeforeEach(func() {
			dslengine.Reset()
			ut = apidsl.Type("example", func() {
				apidsl.Attribute("id")
			})
			ar = apidsl.ArrayOf(ut)
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		It("produces a array type", func() {
			Ω(ar).ShouldNot(BeNil())
			Ω(ar.Kind()).Should(Equal(ArrayKind))
			Ω(ar.ElemType.Type).Should(Equal(ut))
		})
	})

	Context("with a DSL", func() {
		var (
			pattern = "foo"
			ar      *Array
		)

		BeforeEach(func() {
			dslengine.Reset()
			ar = apidsl.ArrayOf(String, func() {
				apidsl.Pattern(pattern)
			})
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		It("records the validations", func() {
			Ω(ar).ShouldNot(BeNil())
			Ω(ar.Kind()).Should(Equal(ArrayKind))
			Ω(ar.ElemType.Type).Should(Equal(String))
			Ω(ar.ElemType.Validation).ShouldNot(BeNil())
			Ω(ar.ElemType.Validation.Pattern).Should(Equal(pattern))
		})
	})

	Context("defined with the type name", func() {
		var ar *UserTypeDefinition
		BeforeEach(func() {
			dslengine.Reset()
			apidsl.Type("name", func() {
				apidsl.Attribute("id")
			})
			ar = apidsl.Type("names", func() {
				apidsl.Attribute("ut", apidsl.ArrayOf("name"))
			})
		})

		JustBeforeEach(func() {
			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		It("produces a user type", func() {
			Ω(ar).ShouldNot(BeNil())
			Ω(ar.TypeName).Should(Equal("names"))
			Ω(ar.Type).ShouldNot(BeNil())
			Ω(ar.Type.ToObject()).ShouldNot(BeNil())
			Ω(ar.Type.ToObject()).Should(HaveKey("ut"))
			ut := ar.Type.ToObject()["ut"]
			Ω(ut.Type).ShouldNot(BeNil())
			Ω(ut.Type).Should(BeAssignableToTypeOf(&Array{}))
			et := ut.Type.ToArray().ElemType
			Ω(et).ShouldNot(BeNil())
			Ω(et.Type).Should(BeAssignableToTypeOf(&UserTypeDefinition{}))
			Ω(et.Type.(*UserTypeDefinition).TypeName).Should(Equal("name"))
		})
	})

	Context("defined with a media type name", func() {
		var mt *MediaTypeDefinition
		BeforeEach(func() {
			dslengine.Reset()
			mt = apidsl.MediaType("application/vnd.test", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("ut", apidsl.ArrayOf("application/vnd.test"))
				})
				apidsl.View("default", func() {
					apidsl.Attribute("ut")
				})
			})
		})

		JustBeforeEach(func() {
			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		It("produces a user type", func() {
			Ω(mt).ShouldNot(BeNil())
			Ω(mt.TypeName).Should(Equal("Test"))
			Ω(mt.Type).ShouldNot(BeNil())
			Ω(mt.Type.ToObject()).ShouldNot(BeNil())
			Ω(mt.Type.ToObject()).Should(HaveKey("ut"))
			ut := mt.Type.ToObject()["ut"]
			Ω(ut.Type).ShouldNot(BeNil())
			Ω(ut.Type).Should(BeAssignableToTypeOf(&Array{}))
			et := ut.Type.ToArray().ElemType
			Ω(et).ShouldNot(BeNil())
			Ω(et.Type).Should(BeAssignableToTypeOf(&MediaTypeDefinition{}))
			Ω(et.Type.(*MediaTypeDefinition).TypeName).Should(Equal("Test"))
		})
	})
})

var _ = Describe("HashOf", func() {
	Context("used on a global variable", func() {
		var (
			kt *UserTypeDefinition
			vt *UserTypeDefinition
			ha *Hash
		)
		BeforeEach(func() {
			dslengine.Reset()
			kt = apidsl.Type("key", func() {
				apidsl.Attribute("id")
			})
			vt = apidsl.Type("val", func() {
				apidsl.Attribute("id")
			})
			ha = apidsl.HashOf(kt, vt)
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		It("produces a hash type", func() {
			Ω(ha).ShouldNot(BeNil())
			Ω(ha.Kind()).Should(Equal(HashKind))
			Ω(ha.KeyType.Type).Should(Equal(kt))
			Ω(ha.ElemType.Type).Should(Equal(vt))
		})
	})

	Context("with DSLs", func() {
		var (
			kp = "foo"
			vp = "bar"
			ha *Hash
		)

		BeforeEach(func() {
			dslengine.Reset()
			ha = apidsl.HashOf(String, String, func() { apidsl.Pattern(kp) }, func() { apidsl.Pattern(vp) })
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			dslengine.Run()
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		})

		It("records the validations", func() {
			Ω(ha).ShouldNot(BeNil())
			Ω(ha.Kind()).Should(Equal(HashKind))
			Ω(ha.KeyType.Type).Should(Equal(String))
			Ω(ha.KeyType.Validation).ShouldNot(BeNil())
			Ω(ha.KeyType.Validation.Pattern).Should(Equal(kp))
			Ω(ha.ElemType.Type).Should(Equal(String))
			Ω(ha.ElemType.Validation).ShouldNot(BeNil())
			Ω(ha.ElemType.Validation.Pattern).Should(Equal(vp))
		})
	})
})
