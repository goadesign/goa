package dsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Describe("Response", func() {
	var name string
	var dsl func()

	var res *ResponseDefinition

	BeforeEach(func() {
		Design = nil
		DSLErrors = nil
		name = ""
		dsl = nil
	})

	JustBeforeEach(func() {
		API("test", func() {
			Resource("res", func() {
				Action("action", func() {
					Response(name, dsl)
				})
			})
		})
		if r, ok := Design.Resources["res"]; ok {
			if a, ok := r.Actions["action"]; ok {
				res = a.Responses[name]
			}
		}
	})

	Context("with no DSL and no name", func() {
		It("produces an invalid action definition", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).Should(HaveOccurred())
		})
	})

	Context("with no DSL", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces an invalid action definition", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).Should(HaveOccurred())
		})
	})

	Context("with a status", func() {
		const status = 201

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Status(status)
			}
		})

		It("produces a valid action definition and sets the status", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.Status).Should(Equal(status))
		})
	})

	Context("with a status and description", func() {
		const status = 201
		const description = "desc"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Status(status)
				Description(description)
			}
		})

		It("sets the status and description", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.Status).Should(Equal(status))
			Ω(res.Description).Should(Equal(description))
		})
	})

})
