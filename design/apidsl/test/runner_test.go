package test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

var _ = API(apiName, func() {
	Description(apiDescription)
})

var _ = Resource(resourceName, func() {
	Description(resourceDescription)
})

var _ = Type(typeName, func() {
	Description(typeDescription)
	Attribute("bar")
})

var _ = MediaType(mediaTypeIdentifier, func() {
	Description(mediaTypeDescription)
	Attributes(func() { Attribute("foo") })
	View("default", func() { Attribute("foo") })
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
