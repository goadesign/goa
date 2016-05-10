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
			Type:       design.Object{required: integer},
			Validation: &dslengine.ValidationDefinition{Required: []string{required}},
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

var _ = Describe("IterateHeaders", func() {
	It("works when Parent.Headers is nil", func() {
		// create a Resource with no headers, Action with one header
		resource := &design.ResourceDefinition{}
		action := &design.ActionDefinition{
			Parent: resource,
			Headers: &design.AttributeDefinition{
				Type: design.Object{
					"a": &design.AttributeDefinition{Type: design.String},
				},
			},
		}
		names := []string{}
		// iterator that collects header names
		it := func(name string, _ bool, _ *design.AttributeDefinition) error {
			names = append(names, name)
			return nil
		}
		Ω(action.IterateHeaders(it)).Should(Succeed(), "despite action.Parent.Headers being nil")
		Ω(names).Should(ConsistOf("a"))
	})
})
