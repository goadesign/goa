package design_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/design"
)

var _ = Describe("Parent", func() {
	var res, parent *design.ResourceDefinition

	JustBeforeEach(func() {
		if res == nil {
			panic("Kapow - forgot to initialize res...")
		}
		design.Design = &design.APIDefinition{Name: "test"}
		design.Design.Resources = map[string]*design.ResourceDefinition{res.Name: res}
		if parent != nil {
			design.Design.Resources[parent.Name] = parent
		}
		parent = res.Parent()
	})

	Context("a resource with a parent", func() {
		BeforeEach(func() {
			res = &design.ResourceDefinition{
				Name:       "Resource",
				ParentName: "Parent",
			}
			parent = &design.ResourceDefinition{
				Name: "Parent",
			}
		})

		It("computes the parent", func() {
			Ω(parent).ShouldNot(BeNil())
		})
	})
})

var _ = Describe("IsRequired", func() {
	var required string
	var attName string

	var attribute *design.AttributeDefinition
	var res bool

	JustBeforeEach(func() {
		integer := &design.AttributeDefinition{Type: design.Integer}
		attribute = &design.AttributeDefinition{
			Type: design.Object{required: integer},
			Validations: []design.ValidationDefinition{
				&design.RequiredValidationDefinition{Names: []string{required}},
			},
		}
		res = attribute.IsRequired(attName)
	})

	Context("called on a required field", func() {
		BeforeEach(func() {
			attName = "required"
			required = "required"
		})

		It("returns true", func() {
			Ω(res).Should(BeTrue())
		})
	})

	Context("called on a non-required field", func() {
		BeforeEach(func() {
			attName = "non-required"
			required = "required"
		})

		It("returns false", func() {
			Ω(res).Should(BeFalse())
		})
	})
})
