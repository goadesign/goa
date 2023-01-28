package test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kyokomi/goa-v1/design"
	"github.com/kyokomi/goa-v1/design/apidsl"
	"github.com/kyokomi/goa-v1/dslengine"
)

// Global test definitions
const apiName = "API"
const apiDescription = "API description"
const resourceName = "R"
const resourceDescription = "R description"
const typeName = "T"
const typeDescription = "T description"
const mediaTypeIdentifier = "mt/json"
const mediaTypeDescription = "MT description"

var _ = apidsl.API(apiName, func() {
	apidsl.Description(apiDescription)
})

var _ = apidsl.Resource(resourceName, func() {
	apidsl.Description(resourceDescription)
})

var _ = apidsl.Type(typeName, func() {
	apidsl.Description(typeDescription)
	apidsl.Attribute("bar")
})

var _ = apidsl.MediaType(mediaTypeIdentifier, func() {
	apidsl.Description(mediaTypeDescription)
	apidsl.Attributes(func() { apidsl.Attribute("foo") })
	apidsl.View("default", func() { apidsl.Attribute("foo") })
})

func init() {
	dslengine.Run()

	var _ = Describe("DSL execution", func() {
		Context("with global DSL definitions", func() {
			It("runs the DSL", func() {
				Ω(dslengine.Errors).Should(BeEmpty())

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

				Ω(Design.MediaTypes).Should(HaveKey(mediaTypeIdentifier))
				Ω(Design.MediaTypes[mediaTypeIdentifier]).ShouldNot(BeNil())
				Ω(Design.MediaTypes[mediaTypeIdentifier].Identifier).Should(Equal(mediaTypeIdentifier))
				Ω(Design.MediaTypes[mediaTypeIdentifier].Description).Should(Equal(mediaTypeDescription))
			})
		})
	})
}
