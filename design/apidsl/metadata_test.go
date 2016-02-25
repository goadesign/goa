package apidsl_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metadata", func() {
	var name string
	var dt DataType
	var dsl func()
	var mtd *MediaTypeDefinition
	var api *APIDefinition
	var rd *ResourceDefinition
	var metadataKey string
	var metadataValue string

	BeforeEach(func() {
		InitDesign()
		dslengine.Errors = nil
		name = ""
		dsl = nil
		dt = nil
	})

	Context("with Metadata declaration", func() {
		JustBeforeEach(func() {
			api = API("Example API", func() {
				Metadata(metadataKey, metadataValue)
			})

			rd = Resource("Example Resource", func() {
				Metadata(metadataKey, metadataValue)
				Action("Example Action", func() {
					Metadata(metadataKey, metadataValue)
				})
				Response("Example Response", func() {
					Metadata(metadataKey, metadataValue)
				})
			})

			mtd = MediaType("Example MediaType", func() {
				Metadata(metadataKey, metadataValue)
				Attribute("Example Attribute", func() {
					Metadata(metadataKey, metadataValue)
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
				Ω(rd.Responses["Example Response"].Metadata).To(Equal(expected))
				Ω(mtd.Metadata).To(Equal(expected))

				var mtdAttribute AttributeDefinition
				mtd.Type.ToObject().IterateAttributes(func(s string, ad *AttributeDefinition) error {
					mtdAttribute = *ad
					return nil
				})
				Ω(mtdAttribute.Metadata).To(Equal(expected))
			})
		})
		Context("with valid metadata string", func() {
			BeforeEach(func() {
				metadataKey = "struct:tag=json"
				metadataValue = "myName,omitempty"
			})

			It("has metadata", func() {
				expected := dslengine.MetadataDefinition{"struct:tag=json": {"myName,omitempty"}}
				Ω(api.Metadata).To(Equal(expected))
				Ω(rd.Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Metadata).To(Equal(expected))
				Ω(rd.Responses["Example Response"].Metadata).To(Equal(expected))
				Ω(mtd.Metadata).To(Equal(expected))

				var mtdAttribute AttributeDefinition
				mtd.Type.ToObject().IterateAttributes(func(s string, ad *AttributeDefinition) error {
					mtdAttribute = *ad
					return nil
				})
				Ω(mtdAttribute.Metadata).To(Equal(expected))
			})
		})
		Context("with invalid metadata string", func() {
			BeforeEach(func() {
				metadataKey = "abc123一二三"
				metadataValue = "˜µ≤≈ç√"
			})

			It("has metadata", func() {
				expected := dslengine.MetadataDefinition{"abc123一二三": {"˜µ≤≈ç√"}}
				Ω(api.Metadata).To(Equal(expected))
				Ω(rd.Metadata).To(Equal(expected))
				Ω(rd.Actions["Example Action"].Metadata).To(Equal(expected))
				Ω(rd.Responses["Example Response"].Metadata).To(Equal(expected))
				Ω(mtd.Metadata).To(Equal(expected))

				var mtdAttribute AttributeDefinition
				mtd.Type.ToObject().IterateAttributes(func(s string, ad *AttributeDefinition) error {
					mtdAttribute = *ad
					return nil
				})
				Ω(mtdAttribute.Metadata).To(Equal(expected))
			})
		})

	})

	Context("with no Metadata declaration", func() {
		JustBeforeEach(func() {
			api = API("Example API", func() {})

			rd = Resource("Example Resource", func() {
				Action("Example Action", func() {
				})
				Response("Example Response", func() {
				})
			})

			mtd = MediaType("Example MediaType", func() {
				Attribute("Example Attribute", func() {
				})
			})

			dslengine.Run()
		})
		It("has no metadata", func() {
			Ω(api.Metadata).To(BeNil())
			Ω(rd.Metadata).To(BeNil())
			Ω(mtd.Metadata).To(BeNil())
			var mtdAttribute AttributeDefinition
			mtd.Type.ToObject().IterateAttributes(func(s string, ad *AttributeDefinition) error {
				mtdAttribute = *ad
				return nil
			})
			Ω(mtdAttribute.Metadata).To(BeNil())
		})
	})

})
