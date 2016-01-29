package design_test

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsRequired", func() {
	var required string
	var attName string

	var attribute *design.AttributeDefinition
	var res bool

	JustBeforeEach(func() {
		integer := &design.AttributeDefinition{Type: design.Integer}
		attribute = &design.AttributeDefinition{
			Type: design.Object{required: integer},
			Validations: []dslengine.ValidationDefinition{
				&dslengine.RequiredValidationDefinition{Names: []string{required}},
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
