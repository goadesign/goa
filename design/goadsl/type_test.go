package goadsl_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/goadsl"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Type", func() {
	var name string
	var goadsl func()

	var ut *UserTypeDefinition

	BeforeEach(func() {
		InitDesign()
		dslengine.Errors = nil
		name = ""
		goadsl = nil
	})

	JustBeforeEach(func() {
		Type(name, goadsl)
		dslengine.Run()
		Ω(dslengine.Errors).ShouldNot(HaveOccurred())
		ut, _ = Design.Types[name]
	})

	Context("with no goadsl and no name", func() {
		It("produces an invalid type definition", func() {
			Ω(ut).ShouldNot(BeNil())
			Ω(ut.Validate("test", Design)).Should(HaveOccurred())
		})
	})

	Context("with no goadsl", func() {
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
			goadsl = func() {
				Attribute(attName)
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
	Context("with a name and date datatype", func() {
		const attName = "att"
		BeforeEach(func() {
			name = "foo"
			goadsl = func() {
				Attribute(attName, DateTime)
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
