package design_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/design"
)

var _ = Describe("Struct", func() {
	Context("given an attribute definition", func() {
		var att *design.AttributeDefinition
		var object design.Object
		var st string

		JustBeforeEach(func() {
			if att == nil {
				att = new(design.AttributeDefinition)
			}
			att.Type = object
			st = att.Struct()
		})

		Context("of primitive types", func() {
			BeforeEach(func() {
				object = design.Object{
					"foo": &design.AttributeDefinition{Type: design.Integer},
					"bar": &design.AttributeDefinition{Type: design.String},
				}
			})

			It(".Struct() produces the struct go code", func() {
				expected := "struct {\n" +
					"	Bar string `json:\"bar,omitempty\"`\n" +
					"	Foo int `json:\"foo,omitempty\"`\n" +
					"}"
				Ω(st).Should(Equal(expected))
			})
		})

		Context("of array of primitive types", func() {
			BeforeEach(func() {
				elemType := &design.AttributeDefinition{Type: design.Integer}
				array := &design.Array{ElemType: elemType}
				object = design.Object{
					"foo": &design.AttributeDefinition{Type: array},
				}
			})

			It(".Struct() produces the struct go code", func() {
				Ω(st).Should(Equal("struct {\n\tFoo []int `json:\"foo,omitempty\"`\n}"))
			})
		})

		Context("of array of objects", func() {
			BeforeEach(func() {
				obj := design.Object{
					"bar": &design.AttributeDefinition{Type: design.Integer},
				}
				elemType := &design.AttributeDefinition{Type: obj}
				array := &design.Array{ElemType: elemType}
				object = design.Object{
					"foo": &design.AttributeDefinition{Type: array},
				}
			})

			It(".Struct() produces the struct go code", func() {
				expected := "struct {\n" +
					"	Foo []struct {\n" +
					"	Bar int `json:\"bar,omitempty\"`\n" +
					"} `json:\"foo,omitempty\"`\n" +
					"}"
				Ω(st).Should(Equal(expected))
			})
		})

		Context("with required fields", func() {
			BeforeEach(func() {
				object = design.Object{
					"foo": &design.AttributeDefinition{Type: design.Integer},
				}
				required := &design.RequiredValidationDefinition{
					Names: []string{"foo"},
				}
				att = &design.AttributeDefinition{
					Validations: []design.ValidationDefinition{required},
				}
			})

			It(".Struct() produces the struct go code", func() {
				expected := "struct {\n" +
					"	Foo int `json:\"foo\"`\n" +
					"}"
				Ω(st).Should(Equal(expected))
			})
		})

	})

})
