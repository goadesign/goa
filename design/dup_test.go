package design_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dup", func() {
	var dt DataType
	var dup DataType

	JustBeforeEach(func() {
		dup = Dup(dt)
	})

	Context("with a primitive type", func() {
		BeforeEach(func() {
			dt = Integer
		})

		It("returns the same value", func() {
			Ω(dup).Should(Equal(dt))
		})
	})

	Context("with an array type", func() {
		var elemType = Integer

		BeforeEach(func() {
			dt = &Array{
				ElemType: &AttributeDefinition{Type: elemType},
			}
		})

		It("returns a duplicate array type", func() {
			Ω(dup).Should(Equal(dt))
			Ω(dup == dt).Should(BeFalse())
			Ω(dup.(*Array).ElemType == dt.(*Array).ElemType).Should(BeFalse())
		})
	})

	Context("with a hash type", func() {
		var keyType = String
		var elemType = Integer

		BeforeEach(func() {
			dt = &Hash{
				KeyType:  &AttributeDefinition{Type: keyType},
				ElemType: &AttributeDefinition{Type: elemType},
			}
		})

		It("returns a duplicate hash type", func() {
			Ω(dup).Should(Equal(dt))
			Ω(dup == dt).Should(BeFalse())
			Ω(dup.(*Hash).KeyType == dt.(*Hash).KeyType).Should(BeFalse())
			Ω(dup.(*Hash).ElemType == dt.(*Hash).ElemType).Should(BeFalse())
		})
	})

	Context("with a user type", func() {
		const typeName = "foo"
		var att = &AttributeDefinition{Type: Integer}

		BeforeEach(func() {
			dt = &UserTypeDefinition{
				TypeName:            typeName,
				AttributeDefinition: att,
			}
		})

		It("returns a duplicate user type", func() {
			Ω(dup).Should(Equal(dt))
			Ω(dup == dt).Should(BeFalse())
			Ω(dup.(*UserTypeDefinition).AttributeDefinition == att).Should(BeFalse())
		})
	})

	Context("with a media type", func() {
		var obj = Object{"att": &AttributeDefinition{Type: Integer}}
		var ut = &UserTypeDefinition{
			TypeName:            "foo",
			AttributeDefinition: &AttributeDefinition{Type: obj},
		}
		const identifier = "vnd.application/test"
		var links = map[string]*LinkDefinition{
			"link": {Name: "att", View: "default"},
		}
		var views = map[string]*ViewDefinition{
			"default": {
				Name:                "default",
				AttributeDefinition: &AttributeDefinition{Type: obj},
			},
		}

		BeforeEach(func() {
			dt = &MediaTypeDefinition{
				UserTypeDefinition: ut,
				Identifier:         identifier,
				Links:              links,
				Views:              views,
			}
		})

		It("returns a duplicate media type", func() {
			Ω(dup).Should(Equal(dt))
			Ω(dup == dt).Should(BeFalse())
			Ω(dup.(*MediaTypeDefinition).UserTypeDefinition == ut).Should(BeFalse())
		})
	})

	Context("with two media types referring to each other", func() {
		var ut *UserTypeDefinition

		BeforeEach(func() {
			mt := &MediaTypeDefinition{Identifier: "application/mt1"}
			mt2 := &MediaTypeDefinition{Identifier: "application/mt2"}
			obj1 := Object{"att": &AttributeDefinition{Type: mt2}}
			obj2 := Object{"att": &AttributeDefinition{Type: mt}}

			att1 := &AttributeDefinition{Type: obj1}
			ut = &UserTypeDefinition{AttributeDefinition: att1}
			link1 := &LinkDefinition{Name: "att", View: "default"}
			view1 := &ViewDefinition{AttributeDefinition: att1, Name: "default"}
			mt.UserTypeDefinition = ut
			mt.Links = map[string]*LinkDefinition{"att": link1}
			mt.Views = map[string]*ViewDefinition{"default": view1}

			att2 := &AttributeDefinition{Type: obj2}
			ut2 := &UserTypeDefinition{AttributeDefinition: att2}
			link2 := &LinkDefinition{Name: "att", View: "default"}
			view2 := &ViewDefinition{AttributeDefinition: att2, Name: "default"}
			mt2.UserTypeDefinition = ut2
			mt2.Links = map[string]*LinkDefinition{"att": link2}
			mt2.Views = map[string]*ViewDefinition{"default": view2}

			dt = mt
		})

		It("duplicates without looping infinitly", func() {
			Ω(dup).Should(Equal(dt))
			Ω(dup == dt).Should(BeFalse())
			Ω(dup.(*MediaTypeDefinition).UserTypeDefinition == ut).Should(BeFalse())
		})
	})
})

var _ = Describe("DupAtt", func() {
	var att *AttributeDefinition
	var dup *AttributeDefinition

	JustBeforeEach(func() {
		dup = DupAtt(att)
	})

	Context("with an attribute with a type which is a media type", func() {
		BeforeEach(func() {
			att = &AttributeDefinition{Type: &MediaTypeDefinition{}}
		})

		It("does not clone the type", func() {
			Ω(dup == att).Should(BeFalse())
			Ω(dup.Type == att.Type).Should(BeTrue())
		})
	})
})
