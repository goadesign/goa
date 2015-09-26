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
		Resource("res", func() {
			Action("action", func() {
				Response(name, dsl)
			})
		})
		RunDSL()
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

		It("produces a valid action definition and sets the status and parent", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.Status).Should(Equal(status))
			Ω(res.Parent).ShouldNot(BeNil())
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

	Context("with a status and name override", func() {
		const status = 201
		const respName = "name"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Status(status)
				Name(respName)
			}
		})

		It("sets the status and name", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.Status).Should(Equal(status))
			Ω(res.Name).Should(Equal(respName))
		})
	})

	Context("with a status and media type", func() {
		const status = 201
		const mediaType = "mt"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Status(status)
				MediaType(mediaType)
			}
		})

		It("sets the status and media type", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.Status).Should(Equal(status))
			Ω(res.MediaType).Should(Equal(mediaType))
		})
	})

	Context("with a status and headers", func() {
		const status = 201
		const headerName = "Location"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Status(status)
				Headers(func() {
					Header(headerName)
				})
			}
		})

		It("sets the status and headers", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.Status).Should(Equal(status))
			Ω(res.Headers).ShouldNot(BeNil())
			Ω(res.Headers.Type).Should(BeAssignableToTypeOf(Object{}))
			o := res.Headers.Type.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(headerName))
		})
	})

})
