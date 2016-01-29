package engine_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/engine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			engine.RunDSL()
		})

		It("still produces the correct metadata", func() {
			Ω(engine.Errors).Should(BeEmpty())
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
		engine.Errors = nil
	})

	JustBeforeEach(func() {
		ErrorMsg = engine.Errors.Error()
	})

	Context("with one error", func() {
		const errMsg = "err"

		// See NOTE below.
		const lineNumber = 75

		BeforeEach(func() {
			// NOTE: moving the line below requires updating the
			// constant above to match its number.
			engine.ReportError(errMsg)
		})

		It("computes the location", func() {
			Ω(ErrorMsg).Should(ContainSubstring(errMsg))
			Ω(engine.Errors).Should(HaveLen(1))
			Ω(engine.Errors[0]).ShouldNot(BeNil())
			Ω(engine.Errors[0].File).Should(HaveSuffix("runner_test.go"))
			Ω(engine.Errors[0].Line).Should(Equal(lineNumber))
		})
	})

	Context("with multiple errors", func() {
		const error1msg = "foo1"
		const error2msg = "foo2"

		BeforeEach(func() {
			engine.ReportError(error1msg)
			engine.ReportError(error2msg)
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
			engine.RunDSL()
		})

		It("reports an invalid DSL error", func() {
			Ω(ErrorMsg).Should(ContainSubstring("invalid use of Attributes"))
			Ω(engine.Errors).Should(HaveLen(1))
			Ω(engine.Errors[0]).ShouldNot(BeNil())
			Ω(engine.Errors[0].File).Should(HaveSuffix("runner_test.go"))
			Ω(engine.Errors[0].Line).Should(Equal(lineNumber))
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
			engine.RunDSL()
		})

		It("reports an incompatible type DSL error", func() {
			Ω(ErrorMsg).Should(ContainSubstring("cannot use 42 (type int) as type"))
			Ω(engine.Errors).Should(HaveLen(1))
			Ω(engine.Errors[0]).ShouldNot(BeNil())
			Ω(engine.Errors[0].File).Should(HaveSuffix("runner_test.go"))
			Ω(engine.Errors[0].Line).Should(Equal(lineNumber))
		})
	})
})
