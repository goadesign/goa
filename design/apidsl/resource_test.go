package apidsl_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource", func() {
	var name string
	var dsl func()

	var res *ResourceDefinition

	BeforeEach(func() {
		dslengine.Reset()
		name = ""
		dsl = nil
	})

	JustBeforeEach(func() {
		res = Resource(name, dsl)
		dslengine.Run()
	})

	Context("with no dsl and no name", func() {
		It("produces an invalid resource definition", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).Should(HaveOccurred())
		})
	})

	Context("with no dsl", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces a valid resource definition and defaults the media type to text/plain", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.MediaType).Should(Equal("text/plain"))
		})
	})

	Context("with a description", func() {
		const description = "desc"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Description(description)
			}
		})

		It("sets the description", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.Description).Should(Equal(description))
		})
	})

	Context("with a parent resource that does not exist", func() {
		const parent = "parent"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Parent(parent)
			}
		})

		It("sets the parent and produces an invalid resource definition", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.ParentName).Should(Equal(parent))
			Ω(res.Validate()).Should(HaveOccurred())
		})
	})

	Context("with actions", func() {
		const actionName = "action"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Action(actionName, func() { Routing(PUT(":/id")) })
			}
		})

		It("sets the actions", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.Actions).Should(HaveLen(1))
			Ω(res.Actions).Should(HaveKey(actionName))
		})
	})

	Context("with metadata and actions", func() {
		const actionName = "action"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Metadata("swagger:generate", "false")
				Action(actionName, func() { Routing(PUT(":/id")) })
			}
		})

		It("sets the actions", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Finalize).ShouldNot(Panic())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.Actions).Should(HaveLen(1))
			Ω(res.Actions).Should(HaveKey(actionName))
		})
	})

	Context("with metadata and files", func() {
		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Metadata("swagger:generate", "false")
				Files("path", "filename")
			}
		})

		It("sets the files", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Finalize).ShouldNot(Panic())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.FileServers).Should(HaveLen(1))
		})
	})

	Context("with a canonical action that does not exist", func() {
		const can = "can"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				CanonicalActionName(can)
			}
		})

		It("sets the canonical action and produces an invalid resource definition", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.CanonicalActionName).Should(Equal(can))
			Ω(res.Validate()).Should(HaveOccurred())
		})
	})

	Context("with a canonical action that does exist", func() {
		const can = "can"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				Action(can, func() { Routing(PUT(":/id")) })
				CanonicalActionName(can)
			}
		})

		It("sets the canonical action and produces a valid resource definition", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.CanonicalActionName).Should(Equal(can))
			Ω(res.Validate()).ShouldNot(HaveOccurred())
		})
	})

	Context("with a base path", func() {
		const basePath = "basePath"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				BasePath(basePath)
			}
		})

		It("sets the base path", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.BasePath).Should(Equal(basePath))
		})
	})

	Context("with base params", func() {
		const basePath = "basePath/:paramID"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				BasePath(basePath)
				Params(func() {
					Param("paramID")
				})
			}
		})

		It("sets the base path and params", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.BasePath).Should(Equal(basePath))
			Ω(res.Params).ShouldNot(BeNil())
			Ω(res.Params.Type).ShouldNot(BeNil())
			Ω(res.Params.Type.ToObject()).Should(HaveKey("paramID"))
		})
	})

	Context("with a media type name", func() {
		const mediaType = "application/mt"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				DefaultMedia(mediaType)
			}
		})

		It("sets the media type", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.MediaType).Should(Equal(mediaType))
		})
	})

	Context("with a view name", func() {
		const mediaType = "application/mt"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				DefaultMedia(mediaType, "compact")
			}
		})

		It("sets the media type", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.MediaType).Should(Equal(mediaType))
			Ω(res.DefaultViewName).Should(Equal("compact"))
		})
	})

	Context("with an invalid media type", func() {
		var mediaType = &MediaTypeDefinition{Identifier: "application/foo"}

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				DefaultMedia(mediaType)
			}
		})

		It("fails", func() {
			Ω(dslengine.Errors).Should(HaveOccurred())
		})
	})

	Context("with a valid media type", func() {
		const typeName = "typeName"
		const identifier = "application/vnd.raphael.goa.test"

		var mediaType = &MediaTypeDefinition{
			UserTypeDefinition: &UserTypeDefinition{
				TypeName: typeName,
			},
			Identifier: identifier,
		}

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				DefaultMedia(mediaType)
			}
		})

		It("sets the media type", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.MediaType).Should(Equal(identifier))
		})
	})

	Context("with a valid media type using a modifier", func() {
		const typeName = "typeName"
		const identifier = "application/vnd.raphael.goa.test+json"

		var mediaType = &MediaTypeDefinition{
			UserTypeDefinition: &UserTypeDefinition{
				TypeName: typeName,
			},
			Identifier: identifier,
		}

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				DefaultMedia(mediaType)
			}
		})

		It("sets the media type and keeps the modifier", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.MediaType).Should(Equal(identifier))
		})
	})

	Context("with a trait that does not exist", func() {
		BeforeEach(func() {
			name = "foo"
			dsl = func() { UseTrait("Authenticated") }
		})

		It("returns an error", func() {
			Ω(dslengine.Errors).Should(HaveOccurred())
		})
	})

	Context("with a trait that exists", func() {
		const description = "desc"
		const traitName = "descTrait"

		BeforeEach(func() {
			name = "foo"
			dsl = func() { UseTrait(traitName) }
			API("test", func() {
				Trait(traitName, func() {
					Description(description)
				})
			})
		})

		It("runs the trait", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.Description).Should(Equal(description))
		})
	})
})
