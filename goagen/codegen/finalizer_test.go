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

	Context("given an object with a primitive Number field", func() {
		BeforeEach(func() {
			att = &design.AttributeDefinition{
				Type: &design.Object{
					"foo": &design.AttributeDefinition{
						Type:         design.Number,
						DefaultValue: 0.0,
					},
				},
			}
			target = "ut"
		})
		It("finalizes the fields", func() {
			code := finalizer.Code(att, target, 0)
			Ω(code).Should(Equal(numberAssignmentCode))
		})
	})

	Context("given an object with a primitive Number field with a int default value", func() {
		BeforeEach(func() {
			att = &design.AttributeDefinition{
				Type: &design.Object{
					"foo": &design.AttributeDefinition{
						Type:         design.Number,
						DefaultValue: 50,
					},
				},
			}
			target = "ut"
		})
		It("finalizes the fields", func() {
			code := finalizer.Code(att, target, 0)
			Ω(code).Should(Equal(numberAssignmentCodeIntDefault))
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
			Ω(code).Should(Equal(recursiveAssignmentCodeA))
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
			Ω(code).Should(Equal(recursiveAssignmentCodeB))
		})
	})
})

const (
	primitiveAssignmentCode = `var defaultFoo = "bar"
if ut.Foo == nil {
	ut.Foo = &defaultFoo
}`

	numberAssignmentCodeIntDefault = `var defaultFoo = 50.000000
if ut.Foo == nil {
	ut.Foo = &defaultFoo
}`

	numberAssignmentCode = `var defaultFoo = 0.000000
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

	recursiveAssignmentCodeA = `if ut.Child != nil {
	var defaultOther = "foo"
	if ut.Child.Other == nil {
		ut.Child.Other = &defaultOther
}
}
var defaultOther = "foo"
if ut.Other == nil {
	ut.Other = &defaultOther
}`

	recursiveAssignmentCodeB = `	for _, e := range ut.Elems {
		var defaultOther = "foo"
		if e.Other == nil {
			e.Other = &defaultOther
}
	}
var defaultOther = "foo"
if ut.Other == nil {
	ut.Other = &defaultOther
}`
)
