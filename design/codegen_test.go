package design_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/design"
)

var _ = Describe("SourceCode", func() {
	Context("given an attribute definition with fields", func() {
		var att *design.AttributeDefinition
		var object design.Object
		var st string

		JustBeforeEach(func() {
			if att == nil {
				att = new(design.AttributeDefinition)
			}
			att.Type = object
			st = design.SourceCode(att)
		})

		Context("of primitive types", func() {
			BeforeEach(func() {
				object = design.Object{
					"foo": &design.AttributeDefinition{Type: design.Integer},
					"bar": &design.AttributeDefinition{Type: design.String},
				}
			})

			It("produces the struct go code", func() {
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

			It("produces the struct go code", func() {
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

			It("produces the struct go code", func() {
				expected := "struct {\n" +
					"	Foo []struct {\n" +
					"	Bar int `json:\"bar,omitempty\"`\n" +
					"} `json:\"foo,omitempty\"`\n" +
					"}"
				Ω(st).Should(Equal(expected))
			})
		})

		Context("that are required", func() {
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

			It("produces the struct go code", func() {
				expected := "struct {\n" +
					"	Foo int `json:\"foo\"`\n" +
					"}"
				Ω(st).Should(Equal(expected))
			})
		})

	})

	Context("given an array", func() {
		var elemType *design.AttributeDefinition
		var source string

		JustBeforeEach(func() {
			array := &design.Array{ElemType: elemType}
			att := &design.AttributeDefinition{Type: array}
			source = design.SourceCode(att)
		})

		Context("of primitive type", func() {
			BeforeEach(func() {
				elemType = &design.AttributeDefinition{Type: design.Integer}
			})

			It("SourceCode() produces the array go code", func() {
				Ω(source).Should(Equal("[]int"))
			})

		})

		Context("of object type", func() {
			BeforeEach(func() {
				object := design.Object{
					"foo": &design.AttributeDefinition{Type: design.Integer},
					"bar": &design.AttributeDefinition{Type: design.String},
				}
				elemType = &design.AttributeDefinition{Type: object}
			})

			It("produces the array go code", func() {
				Ω(source).Should(Equal("[]struct {\n\tBar string `json:\"bar,omitempty\"`\n\tFoo int `json:\"foo,omitempty\"`\n}"))
			})
		})
	})

})
