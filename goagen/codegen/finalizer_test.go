package codegen_test

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Struct finalize code generation", func() {
	var (
		att       *design.AttributeDefinition
		target    string
		finalizer *codegen.Finalizer
	)

	BeforeEach(func() {
		finalizer = codegen.NewFinalizer()
	})

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
			code := finalizer.Code(att, target, 0)
			Ω(code).Should(Equal(primitiveAssignmentCode))
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
			code := finalizer.Code(att, target, 0)
			Ω(code).Should(Equal(arrayAssignmentCode))
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
			code := finalizer.Code(att, target, 0)
			Ω(code).Should(Equal(hashAssignmentCode))
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
			code := finalizer.Code(att, target, 0)
			Ω(code).Should(Equal(datetimeAssignmentCode))
		})
	})

	Context("given a recursive user type", func() {
		BeforeEach(func() {
			var (
				rt  = &design.UserTypeDefinition{TypeName: "recursive"}
				obj = &design.Object{
					"child": &design.AttributeDefinition{Type: rt},
					"other": &design.AttributeDefinition{
						Type:         design.String,
						DefaultValue: "foo",
					},
				}
			)
			rt.AttributeDefinition = &design.AttributeDefinition{Type: obj}

			att = &design.AttributeDefinition{Type: rt}
			target = "ut"
		})
		It("finalizes the recursive type fields", func() {
			code := finalizer.Code(att, target, 0)
			Ω(code).Should(Equal(recursiveAssignmentCode))
		})
	})

	Context("given a recursive user type with an array attribute", func() {
		BeforeEach(func() {
			var (
				rt  = &design.UserTypeDefinition{TypeName: "recursive"}
				ar  = &design.Array{ElemType: &design.AttributeDefinition{Type: rt}}
				obj = &design.Object{
					"elems": &design.AttributeDefinition{Type: ar},
					"other": &design.AttributeDefinition{
						Type:         design.String,
						DefaultValue: "foo",
					},
				}
			)
			rt.AttributeDefinition = &design.AttributeDefinition{Type: obj}

			att = &design.AttributeDefinition{Type: rt}
			target = "ut"
		})
		It("finalizes the recursive type fields", func() {
			code := finalizer.Code(att, target, 0)
			Ω(code).Should(Equal(recursiveAssignmentCode))
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

	recursiveAssignmentCode = `var defaultOther = "foo"
if ut.Other == nil {
	ut.Other = &defaultOther
}`
)
