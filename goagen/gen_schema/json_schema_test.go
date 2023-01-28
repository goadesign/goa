package genschema_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kyokomi/goa-v1/design"
	"github.com/kyokomi/goa-v1/design/apidsl"
	"github.com/kyokomi/goa-v1/dslengine"
	genschema "github.com/kyokomi/goa-v1/goagen/gen_schema"
)

var _ = Describe("TypeSchema", func() {
	var typ design.DataType

	var s *genschema.JSONSchema

	BeforeEach(func() {
		typ = nil
		s = nil
		dslengine.Reset()
		design.ProjectedMediaTypes = make(design.MediaTypeRoot)
	})

	JustBeforeEach(func() {
		s = genschema.TypeSchema(design.Design, typ)
	})

	Context("with a media type", func() {
		BeforeEach(func() {
			apidsl.MediaType("application/foo.bar", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("bar", func() { apidsl.ReadOnly() })
				})
				apidsl.View("default", func() {
					apidsl.Attribute("bar")
				})
			})

			Ω(dslengine.Run()).ShouldNot(HaveOccurred())
			typ = design.Design.MediaTypes["application/foo.bar"]
		})

		It("returns a proper JSON schema type", func() {
			Ω(s).ShouldNot(BeNil())
			Ω(s.Ref).Should(Equal("#/definitions/FooBar"))
		})
	})

	Context("with a media type with self-referencing attributes", func() {
		BeforeEach(func() {
			apidsl.MediaType("application/vnd.menu+json", func() {
				apidsl.Attributes(func() {
					apidsl.Attribute("name", design.String, "The name of an application")
					apidsl.Attribute("children", apidsl.CollectionOf("application/vnd.menu+json"), func() {
						apidsl.View("nameonly")
					})

				})
				apidsl.View("default", func() {
					apidsl.Attribute("name")
					apidsl.Attribute("children", func() {
						apidsl.View("nameonly")
					})
				})
				apidsl.View("nameonly", func() {
					apidsl.Attribute("name")
				})
			})

			Ω(func() { dslengine.Run() }).ShouldNot(Panic())
			Ω(dslengine.Errors).ShouldNot(HaveOccurred())
			typ = design.Design.MediaTypes["application/vnd.menu"]
		})

		It("returns a proper JSON schema type", func() {
			Ω(s).ShouldNot(BeNil())
			Ω(s.Ref).Should(Equal("#/definitions/Menu"))
		})

	})
})
