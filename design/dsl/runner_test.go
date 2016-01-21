package dsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/dsl"
)

var _ = Describe("DSL execution", func() {
	Context("with cyclical type dependencies", func() {
		const type1Name = "type1Name"
		const type2Name = "type2Name"
		const att1Name = "att1Name"
		const att2Name = "att2Name"

		BeforeEach(func() {
			InitDesign()

			API("foo", func() {})

			var type1, type2 *UserTypeDefinition

			type1 = Type(type1Name, func() {
				Attribute(att1Name, type2)
			})
			type2 = Type(type2Name, func() {
				Attribute(att2Name, type1)
			})
		})

		JustBeforeEach(func() {
			RunDSL()
		})

		It("still produces the correct metadata", func() {
			Ω(Errors).Should(BeEmpty())
			Ω(Design.Types).Should(HaveLen(2))
			t1 := Design.Types[type1Name]
			t2 := Design.Types[type2Name]
			Ω(t1).ShouldNot(BeNil())
			Ω(t2).ShouldNot(BeNil())
			Ω(t1.Type).Should(BeAssignableToTypeOf(Object{}))
			Ω(t2.Type).Should(BeAssignableToTypeOf(Object{}))
			o1 := t1.Type.(Object)
			o2 := t2.Type.(Object)
			Ω(o1).Should(HaveKey(att1Name))
			Ω(o2).Should(HaveKey(att2Name))
			Ω(o1[att1Name].Type).Should(Equal(t2))
			Ω(o2[att2Name].Type).Should(Equal(t1))
		})
	})
})

var _ = Describe("DSL errors", func() {
	var ErrorMsg string

	BeforeEach(func() {
		Errors = nil
	})

	JustBeforeEach(func() {
		ErrorMsg = Errors.Error()
	})

	Context("with one error", func() {
		const errMsg = "err"

		// See NOTE below.
		const lineNumber = 75

		BeforeEach(func() {
			// NOTE: moving the line below requires updating the
			// constant above to match its number.
			ReportError(errMsg)
		})

		It("computes the location", func() {
			Ω(ErrorMsg).Should(ContainSubstring(errMsg))
			Ω(Errors).Should(HaveLen(1))
			Ω(Errors[0]).ShouldNot(BeNil())
			Ω(Errors[0].File).Should(HaveSuffix("runner_test.go"))
			Ω(Errors[0].Line).Should(Equal(lineNumber))
		})
	})

	Context("with multiple errors", func() {
		const error1msg = "foo1"
		const error2msg = "foo2"

		BeforeEach(func() {
			ReportError(error1msg)
			ReportError(error2msg)
		})

		It("reports all errors", func() {
			Ω(ErrorMsg).Should(ContainSubstring(error1msg))
			Ω(ErrorMsg).Should(ContainSubstring(error2msg))
		})
	})

	Context("with invalid DSL", func() {
		// See NOTE below.
		const lineNumber = 111

		BeforeEach(func() {
			InitDesign()
			API("foo", func() {
				// NOTE: moving the line below requires updating the
				// constant above to match its number.
				Attributes(func() {})
			})
			RunDSL()
		})

		It("reports an invalid DSL error", func() {
			Ω(ErrorMsg).Should(ContainSubstring("invalid use of Attributes"))
			Ω(Errors).Should(HaveLen(1))
			Ω(Errors[0]).ShouldNot(BeNil())
			Ω(Errors[0].File).Should(HaveSuffix("runner_test.go"))
			Ω(Errors[0].Line).Should(Equal(lineNumber))
		})
	})

	Context("with DSL calling a function with an invalid argument type", func() {
		// See NOTE below.
		const lineNumber = 134

		BeforeEach(func() {
			InitDesign()
			Type("bar", func() {
				// NOTE: moving the line below requires updating the
				// constant above to match its number.
				Attribute("baz", 42)
			})
			RunDSL()
		})

		It("reports an incompatible type DSL error", func() {
			Ω(ErrorMsg).Should(ContainSubstring("cannot use 42 (type int) as type"))
			Ω(Errors).Should(HaveLen(1))
			Ω(Errors[0]).ShouldNot(BeNil())
			Ω(Errors[0].File).Should(HaveSuffix("runner_test.go"))
			Ω(Errors[0].Line).Should(Equal(lineNumber))
		})
	})
})
