package dsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Describe("DSL errors", func() {
	var dslErrorMsg string

	BeforeEach(func() {
		DSLErrors = nil
	})

	JustBeforeEach(func() {
		dslErrorMsg = DSLErrors.Error()
	})

	Context("with one error", func() {
		const errMsg = "err"

		// See NOTE below.
		const lineNumber = 30

		BeforeEach(func() {
			// NOTE: moving the line below requires updating the
			// constant above to match its number.
			ReportError(errMsg)
		})

		It("computes the location", func() {
			Ω(dslErrorMsg).Should(ContainSubstring(errMsg))
			Ω(DSLErrors).Should(HaveLen(1))
			Ω(DSLErrors[0]).ShouldNot(BeNil())
			Ω(DSLErrors[0].File).Should(HaveSuffix("runner_test.go"))
			Ω(DSLErrors[0].Line).Should(Equal(lineNumber))
		})
	})

	Context("with multiple errors", func() {
		const error1msg = "foo"
		const error2msg = "foo"

		BeforeEach(func() {
			ReportError(error1msg)
			ReportError(error2msg)
		})

		It("reports all errors", func() {
			Ω(dslErrorMsg).Should(ContainSubstring(error1msg))
			Ω(dslErrorMsg).Should(ContainSubstring(error2msg))
		})
	})

	Context("with invalid DSL", func() {
		// See NOTE below.
		const lineNumber = 66

		BeforeEach(func() {
			Design = nil
			API("foo", func() {
				// NOTE: moving the line below requires updating the
				// constant above to match its number.
				Attributes(func() {})
			})
		})

		It("reports an invalid DSL error", func() {
			Ω(dslErrorMsg).Should(ContainSubstring("invalid use of Attributes"))
			Ω(DSLErrors).Should(HaveLen(1))
			Ω(DSLErrors[0]).ShouldNot(BeNil())
			Ω(DSLErrors[0].File).Should(HaveSuffix("runner_test.go"))
			Ω(DSLErrors[0].Line).Should(Equal(lineNumber))
		})
	})

	Context("with DSL calling a function with an invalid argument type", func() {
		// See NOTE below.
		const lineNumber = 89

		BeforeEach(func() {
			Design = nil
			API("foo", func() {
				Type("bar", func() {
					// NOTE: moving the line below requires updating the
					// constant above to match its number.
					Attribute("baz", 42)
				})
			})
		})

		It("reports an incompatible type DSL error", func() {
			Ω(dslErrorMsg).Should(ContainSubstring("cannot use 42 (type int) as type"))
			Ω(DSLErrors).Should(HaveLen(1))
			Ω(DSLErrors[0]).ShouldNot(BeNil())
			Ω(DSLErrors[0].File).Should(HaveSuffix("runner_test.go"))
			Ω(DSLErrors[0].Line).Should(Equal(lineNumber))
		})
	})

})
