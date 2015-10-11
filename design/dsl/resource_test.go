package dsl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Describe("Resource", func() {
	var name string
	var dsl func()

	var res *ResourceDefinition

	BeforeEach(func() {
		Design = nil
		Errors = nil
		name = ""
		dsl = nil
	})

	JustBeforeEach(func() {
		res = Resource(name, dsl)
		RunDSL()
	})

	Context("with no DSL and no name", func() {
		It("produces an invalid resource definition", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).Should(HaveOccurred())
		})
	})

	Context("with no DSL", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces a valid resource definition and defaults the media type to plain/text", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.MediaType).Should(Equal("plain/text"))
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

	Context("with a media type name", func() {
		const mediaType = "mt"

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				MediaType(mediaType)
			}
		})

		It("sets the media type", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.MediaType).Should(Equal(mediaType))
		})
	})

	Context("with an invalid media type", func() {
		var mediaType = &MediaTypeDefinition{Identifier: "foo"}

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				MediaType(mediaType)
			}
		})

		It("fails", func() {
			Ω(Errors).Should(HaveOccurred())
		})
	})

	Context("with a valid media type", func() {
		const typeName = "typeName"
		const identifier = "vnd.raphael.goa.test"

		var mediaType = &MediaTypeDefinition{
			UserTypeDefinition: &UserTypeDefinition{
				TypeName: typeName,
			},
			Identifier: identifier,
		}

		BeforeEach(func() {
			name = "foo"
			dsl = func() {
				MediaType(mediaType)
			}
		})

		It("sets the media type", func() {
			Ω(res).ShouldNot(BeNil())
			Ω(res.Validate()).ShouldNot(HaveOccurred())
			Ω(res.MediaType).Should(Equal(identifier))
		})
	})

	Context("with a trait that does not exist", func() {
		BeforeEach(func() {
			name = "foo"
			dsl = func() { Trait("Authenticated") }
		})

		It("returns an error", func() {
			Ω(Errors).Should(HaveOccurred())
		})
	})

	Context("with a trait that exists", func() {
		const description = "desc"
		const traitName = "descTrait"

		BeforeEach(func() {
			name = "foo"
			dsl = func() { Trait(traitName) }
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
