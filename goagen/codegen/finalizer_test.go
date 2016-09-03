package codegen_test

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Struct finalize code generation", func() {
	Describe("RecursiveFinalizer", func() {
		var att *design.AttributeDefinition
		var target string
		Context("given an object with a primitive field", func() {
			BeforeEach(func() {
				att = &design.AttributeDefinition{
					Type: &design.Object{
						"foo": &design.AttributeDefinition{
							Type:         design.String,
							DefaultValue: "bar",
						},
					},
				}
				target = "ut"
			})
			It("finalizes the fields", func() {
				assignments := codegen.RecursiveFinalizer(att, target, 0)
				Ω(assignments).Should(Equal(primitiveAssignmentCode))
			})
		})
		Context("given an array field", func() {
			BeforeEach(func() {
				att = &design.AttributeDefinition{
					Type: &design.Object{
						"foo": &design.AttributeDefinition{
							Type: &design.Array{
								ElemType: &design.AttributeDefinition{
									Type: design.String,
								},
							},
							DefaultValue: []interface{}{"bar", "baz"},
						},
					},
				}
				target = "ut"
			})
			It("finalizes the array fields", func() {
				assignments := codegen.RecursiveFinalizer(att, target, 0)
				Ω(assignments).Should(Equal(arrayAssignmentCode))
			})
		})
		Context("given a hash field", func() {
			BeforeEach(func() {
				att = &design.AttributeDefinition{
					Type: &design.Object{
						"foo": &design.AttributeDefinition{
							Type: &design.Hash{
								KeyType: &design.AttributeDefinition{
									Type: design.String,
								},
								ElemType: &design.AttributeDefinition{
									Type: design.String,
								},
							},
							DefaultValue: map[interface{}]interface{}{"bar": "baz"},
						},
					},
				}
				target = "ut"
			})
			It("finalizes the hash fields", func() {
				assignments := codegen.RecursiveFinalizer(att, target, 0)
				Ω(assignments).Should(Equal(hashAssignmentCode))
			})
		})
		Context("given a datetime field", func() {
			BeforeEach(func() {
				att = &design.AttributeDefinition{
					Type: &design.Object{
						"foo": &design.AttributeDefinition{
							Type:         design.DateTime,
							DefaultValue: interface{}("1978-06-30T10:00:00+09:00"),
						},
					},
				}
				target = "ut"
			})
			It("finalizes the hash fields", func() {
				assignments := codegen.RecursiveFinalizer(att, target, 0)
				Ω(assignments).Should(Equal(datetimeAssignmentCode))
			})
		})

	})
})

const (
	primitiveAssignmentCode = `var defaultFoo = "bar"
if ut.Foo == nil {
	ut.Foo = &defaultFoo
}`

	arrayAssignmentCode = `if ut.Foo == nil {
	ut.Foo = []string{"bar", "baz"}
}`

	hashAssignmentCode = `if ut.Foo == nil {
	ut.Foo = map[string]string{"bar": "baz"}
}`
	datetimeAssignmentCode = `var defaultFoo, _ = time.Parse(time.RFC3339, "1978-06-30T10:00:00+09:00")
if ut.Foo == nil {
	ut.Foo = &defaultFoo
}`
)
