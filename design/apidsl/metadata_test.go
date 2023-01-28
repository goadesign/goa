package apidsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kyokomi/goa-v1/design"
	"github.com/kyokomi/goa-v1/design/apidsl"
	"github.com/kyokomi/goa-v1/dslengine"
)

var _ = Describe("Metadata", func() {
	var mtd *MediaTypeDefinition
	var api *APIDefinition
	var rd *ResourceDefinition
	var metadataKey string
	var metadataValue string

	BeforeEach(func() {
		dslengine.Reset()
	})

	Context("with Metadata declaration", func() {
		JustBeforeEach(func() {
			api = apidsl.API("Example API", func() {
				apidsl.Metadata(metadataKey, metadataValue)
				apidsl.BasicAuthSecurity("password")
			})

			rd = apidsl.Resource("Example Resource", func() {
				apidsl.Metadata(metadataKey, metadataValue)
				apidsl.Action("Example Action", func() {
					apidsl.Metadata(metadataKey, metadataValue)
					apidsl.Routing(
						apidsl.GET("/", func() {
							apidsl.Metadata(metadataKey, metadataValue)
						}),
					)
					apidsl.Security("password", func() {
						apidsl.Metadata(metadataKey, metadataValue)
					})
				})
				apidsl.Response("Example Response", func() {
					apidsl.Metadata(metadataKey, metadataValue)
				})
			})

			mtd = apidsl.MediaType("Example MediaType", func() {
				apidsl.Metadata(metadataKey, metadataValue)
				apidsl.Attribute("Example Attribute", func() {
					apidsl.Metadata(metadataKey, metadataValue)
				})
			})

			dslengine.Run()
		})

		Context("with blank metadata string", func() {
			BeforeEach(func() {
				metadataKey = ""
				metadataValue = ""
			})

			It("has metadata", func() {
				expected := dslengine.MetadataDefinition{"": {""}}
				Ω(api.Metadata).To(Equal(expected))
				Ω(rd.Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Routes[0].Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Security.Scheme.Metadata).To(Equal(expected))
				Ω(rd.Responses["Example Response"].Metadata).To(Equal(expected))
				Ω(mtd.Metadata).To(Equal(expected))

				var mtdAttribute *AttributeDefinition
				mtdAttribute = mtd.Type.ToObject()["Example Attribute"]
				Ω(mtdAttribute.Metadata).To(Equal(expected))
			})
		})
		Context("with valid metadata string", func() {
			BeforeEach(func() {
				metadataKey = "struct:tag:json"
				metadataValue = "myName,omitempty"
			})

			It("has metadata", func() {
				expected := dslengine.MetadataDefinition{"struct:tag:json": {"myName,omitempty"}}
				Ω(api.Metadata).To(Equal(expected))
				Ω(rd.Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Routes[0].Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Security.Scheme.Metadata).To(Equal(expected))
				Ω(rd.Responses["Example Response"].Metadata).To(Equal(expected))
				Ω(mtd.Metadata).To(Equal(expected))

				var mtdAttribute *AttributeDefinition
				mtdAttribute = mtd.Type.ToObject()["Example Attribute"]
				Ω(mtdAttribute.Metadata).To(Equal(expected))
			})
		})
		Context("with unicode metadata string", func() {
			BeforeEach(func() {
				metadataKey = "abc123一二三"
				metadataValue = "˜µ≤≈ç√"
			})

			It("has metadata", func() {
				expected := dslengine.MetadataDefinition{"abc123一二三": {"˜µ≤≈ç√"}}
				Ω(api.Metadata).To(Equal(expected))
				Ω(rd.Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Routes[0].Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Security.Scheme.Metadata).To(Equal(expected))
				Ω(rd.Responses["Example Response"].Metadata).To(Equal(expected))
				Ω(mtd.Metadata).To(Equal(expected))

				var mtdAttribute *AttributeDefinition
				mtdAttribute = mtd.Type.ToObject()["Example Attribute"]
				Ω(mtdAttribute.Metadata).To(Equal(expected))
			})
		})

	})

	Context("with no Metadata declaration", func() {
		JustBeforeEach(func() {
			api = apidsl.API("Example API", func() {})

			rd = apidsl.Resource("Example Resource", func() {
				apidsl.Action("Example Action", func() {
				})
				apidsl.Response("Example Response", func() {
				})
			})

			mtd = apidsl.MediaType("Example MediaType", func() {
				apidsl.Attribute("Example Attribute", func() {
				})
			})

			dslengine.Run()
		})
		It("has no metadata", func() {
			Ω(api.Metadata).To(BeNil())
			Ω(rd.Metadata).To(BeNil())
			Ω(mtd.Metadata).To(BeNil())

			var mtdAttribute *AttributeDefinition
			mtdAttribute = mtd.Type.ToObject()["Example Attribute"]
			Ω(mtdAttribute.Metadata).To(BeNil())
		})
	})

})
