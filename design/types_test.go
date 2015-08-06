package design_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/design"
)

var _ = Describe("Struct", func() {
	Context("given an array", func() {
		var elemType *design.AttributeDefinition
		var st string

		JustBeforeEach(func() {
			array := &design.Array{ElemType: elemType}
			st = array.Struct()
		})

		Context("of primitive type", func() {
			BeforeEach(func() {
				elemType = &design.AttributeDefinition{Type: design.Integer}
			})

			It(".Struct() produces the array go code", func() {
				Ω(st).Should(Equal("[]int"))
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

			It(".Struct() produces the array go code", func() {
				Ω(st).Should(Equal("[]struct {\n\tBar string `json:\"bar,omitempty\"`\n\tFoo int `json:\"foo,omitempty\"`\n}"))
			})
		})
	})

})
