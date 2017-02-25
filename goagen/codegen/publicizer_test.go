package codegen_test

import (
	"fmt"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Struct publicize code generation", func() {
	Describe("Publicizer", func() {
		var att *design.AttributeDefinition
		var sourceField, targetField string
		Context("given a simple field", func() {
			BeforeEach(func() {
				att = &design.AttributeDefinition{Type: design.Integer}
				sourceField = "source"
				targetField = "target"
			})
			Context("with init false", func() {
				It("simply copies the field over", func() {
					publication := codegen.Publicizer(att, sourceField, targetField, false, 0, false)
					Ω(publication).Should(Equal(fmt.Sprintf("%s = %s", targetField, sourceField)))
				})
			})
			Context("with init true", func() {
				It("initializes and copies the field copies the field over", func() {
					publication := codegen.Publicizer(att, sourceField, targetField, false, 0, true)
					Ω(publication).Should(Equal(fmt.Sprintf("%s := %s", targetField, sourceField)))
				})
			})
		})
		Context("given an object field", func() {
			BeforeEach(func() {
				att = &design.AttributeDefinition{
					Type: design.Object{
						"foo": &design.AttributeDefinition{Type: design.String},
					},
				}
				sourceField = "source"
				targetField = "target"
			})
			It("copies the struct fields", func() {
				publication := codegen.Publicizer(att, sourceField, targetField, false, 0, false)
				Ω(publication).Should(Equal(objectPublicizeCode))
			})
		})
		Context("given a user type", func() {
			BeforeEach(func() {
				att = &design.AttributeDefinition{
					Type: &design.UserTypeDefinition{
						AttributeDefinition: &design.AttributeDefinition{
							Type: &design.Object{
								"foo": &design.AttributeDefinition{Type: design.String},
							},
						},
						TypeName: "TheUserType",
					},
				}
				sourceField = "source"
				targetField = "target"
			})
			It("calls Publicize on the source field", func() {
				publication := codegen.Publicizer(att, sourceField, targetField, false, 0, false)
				Ω(publication).Should(Equal(fmt.Sprintf("%s = %s.Publicize()", targetField, sourceField)))
			})
		})
		Context("given an array field", func() {
			Context("that contains primitive fields", func() {
				BeforeEach(func() {
					att = &design.AttributeDefinition{
						Type: &design.Array{
							ElemType: &design.AttributeDefinition{
								Type: design.String,
							},
						},
					}
					sourceField = "source"
					targetField = "target"
				})
				It("copies the array fields", func() {
					publication := codegen.Publicizer(att, sourceField, targetField, false, 0, false)
					Ω(publication).Should(Equal(fmt.Sprintf("%s = %s", targetField, sourceField)))
				})
			})
			Context("that contains user defined fields", func() {
				BeforeEach(func() {
					att = &design.AttributeDefinition{
						Type: &design.Array{
							ElemType: &design.AttributeDefinition{
								Type: &design.UserTypeDefinition{
									AttributeDefinition: &design.AttributeDefinition{
										Type: design.Object{
											"foo": &design.AttributeDefinition{Type: design.String},
										},
									},
									TypeName: "TheUserType",
								},
							},
						},
					}
					sourceField = "source"
					targetField = "target"
				})
				It("copies the array fields", func() {
					publication := codegen.Publicizer(att, sourceField, targetField, false, 0, false)
					Ω(publication).Should(Equal(arrayPublicizeCode))
				})
			})
		})
		Context("given a hash field", func() {
			Context("that contains primitive fields", func() {
				BeforeEach(func() {
					att = &design.AttributeDefinition{
						Type: &design.Hash{
							KeyType: &design.AttributeDefinition{
								Type: design.String,
							},
							ElemType: &design.AttributeDefinition{
								Type: design.String,
							},
						},
					}
					sourceField = "source"
					targetField = "target"
				})
				It("copies the hash fields", func() {
					publication := codegen.Publicizer(att, sourceField, targetField, false, 0, false)
					Ω(publication).Should(Equal(fmt.Sprintf("%s = %s", targetField, sourceField)))
				})
			})
			Context("that contains user defined fields", func() {
				BeforeEach(func() {
					att = &design.AttributeDefinition{
						Type: &design.Hash{
							KeyType: &design.AttributeDefinition{
								Type: &design.UserTypeDefinition{
									AttributeDefinition: &design.AttributeDefinition{
										Type: &design.Object{
											"foo": &design.AttributeDefinition{Type: design.String},
										},
									},
									TypeName: "TheKeyType",
								},
							},
							ElemType: &design.AttributeDefinition{
								Type: &design.UserTypeDefinition{
									AttributeDefinition: &design.AttributeDefinition{
										Type: &design.Object{
											"bar": &design.AttributeDefinition{Type: design.String},
										},
									},
									TypeName: "TheElemType",
								},
							},
						},
					}
					sourceField = "source"
					targetField = "target"
				})
				It("copies the hash fields", func() {
					publication := codegen.Publicizer(att, sourceField, targetField, false, 0, false)
					Ω(publication).Should(Equal(hashPublicizeCode))
				})
			})
		})
	})
})

const (
	objectPublicizeCode = `target = &struct {
	Foo *string ` + "`" + `form:"foo,omitempty" json:"foo,omitempty" xml:"foo,omitempty"` + "`" + `
}{}
if source.Foo != nil {
	target.Foo = source.Foo
}`

	arrayPublicizeCode = `target = make([]*TheUserType, len(source))
for i0, elem0 := range source {
	target[i0] = elem0.Publicize()
}`

	hashPublicizeCode = `target = make(map[*TheKeyType]*TheElemType, len(source))
for k0, v0 := range source {
	var pubk0 *TheKeyType
	if k0 != nil {
		pubk0 = k0.Publicize()
	}
	var pubv0 *TheElemType
	if v0 != nil {
		pubv0 = v0.Publicize()
	}
	target[pubk0] = pubv0
}`
)
