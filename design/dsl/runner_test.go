package dsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Describe("DSL errors", func() {
	var dslErrorMsg string

	JustBeforeEach(func() {
		dslErrorMsg = DSLErrors.Error()
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

})
