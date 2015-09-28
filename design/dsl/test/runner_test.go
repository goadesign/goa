package test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

// Global test definitions
const apiName = "API"
const apiDescription = "API description"
const resourceName = "R"
const resourceDescription = "R description"
const typeName = "T"
const typeDescription = "T description"
const mediaTypeName = "MT"
const mediaTypeDescription = "MT description"

var _ = API(apiName, func() {
	Description(apiDescription)
})

var _ = Resource(resourceName, func() {
	Description(resourceDescription)
})

var _ = Type(typeName, func() {
	Description(typeDescription)
})

var _ = MediaType(mediaTypeName, func() {
	Description(mediaTypeDescription)
})

func init() {
	RunDSL()

	var _ = Describe("DSL execution", func() {
		Context("with global DSL definitions", func() {
			It("runs the DSL", func() {
				Ω(Errors).Should(BeEmpty())

				Ω(Design).ShouldNot(BeNil())
				Ω(Design.Name).Should(Equal(apiName))
				Ω(Design.Description).Should(Equal(apiDescription))

				Ω(Design.Resources).Should(HaveKey(resourceName))
				Ω(Design.Resources[resourceName]).ShouldNot(BeNil())
				Ω(Design.Resources[resourceName].Name).Should(Equal(resourceName))
				Ω(Design.Resources[resourceName].Description).Should(Equal(resourceDescription))

				Ω(Design.Types).Should(HaveKey(typeName))
				Ω(Design.Types[typeName]).ShouldNot(BeNil())
				Ω(Design.Types[typeName].TypeName).Should(Equal(typeName))
				Ω(Design.Types[typeName].Description).Should(Equal(typeDescription))

				Ω(Design.MediaTypes).Should(HaveKey(mediaTypeName))
				Ω(Design.MediaTypes[mediaTypeName]).ShouldNot(BeNil())
				Ω(Design.MediaTypes[mediaTypeName].TypeName).Should(Equal(mediaTypeName))
				Ω(Design.MediaTypes[mediaTypeName].Description).Should(Equal(mediaTypeDescription))
			})
		})
	})
}
