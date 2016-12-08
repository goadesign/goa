package codegen_test

import (
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/codegen"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("validation code generation", func() {
	BeforeEach(func() {
		codegen.TempCount = 0
	})

	Describe("ValidationChecker", func() {
		Context("given an attribute definition and validations", func() {
			var attType design.DataType
			var validation *dslengine.ValidationDefinition

			att := new(design.AttributeDefinition)
			target := "val"
			context := "context"
			var code string // generated code

			JustBeforeEach(func() {
				att.Type = attType
				att.Validation = validation
				code = codegen.NewValidator().Code(att, false, false, false, target, context, 1, false)
			})

			Context("of enum", func() {
				BeforeEach(func() {
					attType = design.Integer
					validation = &dslengine.ValidationDefinition{
						Values: []interface{}{1, 2, 3},
					}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(enumValCode))
				})
			})

			Context("of pattern", func() {
				BeforeEach(func() {
					attType = design.String
					validation = &dslengine.ValidationDefinition{
						Pattern: ".*",
					}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(patternValCode))
				})
			})

			Context("of min value 0", func() {
				BeforeEach(func() {
					attType = design.Integer
					min := 0.0
					validation = &dslengine.ValidationDefinition{
						Minimum: &min,
					}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(minValCode))
				})
			})

			Context("of array min length 1", func() {
				BeforeEach(func() {
					attType = &design.Array{
						ElemType: &design.AttributeDefinition{
							Type: design.String,
						},
					}
					min := 1
					validation = &dslengine.ValidationDefinition{
						MinLength: &min,
					}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(arrayMinLengthValCode))
				})
			})

			Context("of string min length 2", func() {
				BeforeEach(func() {
					attType = design.String
					min := 2
					validation = &dslengine.ValidationDefinition{
						MinLength: &min,
					}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(stringMinLengthValCode))
				})
			})

			Context("of embedded object", func() {
				var catt, ccatt *design.AttributeDefinition

				BeforeEach(func() {
					enumVal := &dslengine.ValidationDefinition{
						Values: []interface{}{1, 2, 3},
					}
					ccatt = &design.AttributeDefinition{
						Type:       design.Integer,
						Validation: enumVal,
					}
					catt = &design.AttributeDefinition{
						Type: design.Object{"bar": ccatt},
					}

					attType = design.Object{"foo": catt}
				})
				Context("and the parent is optional", func() {
					BeforeEach(func() {
						validation = nil
					})
					It("checks the child & parent object are not nil", func() {
						Ω(code).Should(Equal(embeddedValCode))
					})
				})
				Context("and the parent is required", func() {
					BeforeEach(func() {
						validation = &dslengine.ValidationDefinition{
							Required: []string{"foo"},
						}
					})
					It("checks the child & parent object are not nil", func() {
						Ω(code).Should(Equal(embeddedRequiredValCode))
					})
				})
				Context("with a child attribute with struct:tag:name metadata", func() {
					const fieldTag = "FOO"

					BeforeEach(func() {
						catt.Metadata = dslengine.MetadataDefinition{"struct:field:name": []string{fieldTag}}
						ccatt.Metadata = nil
						validation = nil
					})

					It("produces the validation go code using the field tag", func() {
						Ω(code).Should(Equal(strings.Replace(tagCode, "__tag__", fieldTag, -1)))
					})
				})
				Context("with a grand child attribute with struct:tag:name metadata", func() {
					const fieldTag = "FOO"

					BeforeEach(func() {
						catt.Metadata = nil
						ccatt.Metadata = dslengine.MetadataDefinition{"struct:field:name": []string{fieldTag}}
						validation = nil
					})

					It("produces the validation go code using the field tag", func() {
						Ω(code).Should(Equal(strings.Replace(tagChildCode, "__tag__", fieldTag, -1)))
					})
				})

			})

			Context("of required user type attribute with no validation", func() {
				var ut *design.UserTypeDefinition

				BeforeEach(func() {
					ut = &design.UserTypeDefinition{
						TypeName: "UT",
						AttributeDefinition: &design.AttributeDefinition{
							Type: design.Object{
								"bar": &design.AttributeDefinition{Type: design.String},
							},
						},
					}
					uatt := &design.AttributeDefinition{Type: design.Dup(ut)}
					arr := &design.AttributeDefinition{
						Type: &design.Array{
							ElemType: &design.AttributeDefinition{Type: ut},
						},
					}

					attType = design.Object{"foo": arr, "foo2": uatt}
					validation = &dslengine.ValidationDefinition{
						Required: []string{"foo"},
					}
				})

				Context("with only direct required attributes", func() {
					BeforeEach(func() {
						validation = &dslengine.ValidationDefinition{
							Required: []string{"foo"},
						}
					})

					It("does not call Validate on the user type attribute", func() {
						Ω(code).Should(Equal(utCode))
					})
				})

				Context("with required attributes on inner attribute", func() {
					BeforeEach(func() {
						ut.AttributeDefinition.Validation = &dslengine.ValidationDefinition{
							Required: []string{"bar"},
						}
						validation = nil
					})

					It("calls Validate on the user type attribute", func() {
						Ω(code).Should(Equal(utRequiredCode))
					})
				})
			})

		})
	})
})

const (
	enumValCode = `	if val != nil {
		if !(*val == 1 || *val == 2 || *val == 3) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError(` + "`context`" + `, *val, []interface{}{1, 2, 3}))
		}
	}`

	patternValCode = `	if val != nil {
		if ok := goa.ValidatePattern(` + "`.*`" + `, *val); !ok {
			err = goa.MergeErrors(err, goa.InvalidPatternError(` + "`context`" + `, *val, ` + "`.*`" + `))
		}
	}`

	minValCode = `	if val != nil {
		if *val < 0 {
			err = goa.MergeErrors(err, goa.InvalidRangeError(` + "`" + `context` + "`" + `, *val, 0, true))
		}
	}`

	arrayMinLengthValCode = `	if val != nil {
		if len(val) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError(` + "`" + `context` + "`" + `, val, len(val), 1, true))
		}
	}`

	stringMinLengthValCode = `	if val != nil {
		if utf8.RuneCountInString(*val) < 2 {
			err = goa.MergeErrors(err, goa.InvalidLengthError(` + "`" + `context` + "`" + `, *val, utf8.RuneCountInString(*val), 2, true))
		}
	}`

	embeddedValCode = `	if val.Foo != nil {
		if val.Foo.Bar != nil {
			if !(*val.Foo.Bar == 1 || *val.Foo.Bar == 2 || *val.Foo.Bar == 3) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError(` + "`" + `context.foo.bar` + "`" + `, *val.Foo.Bar, []interface{}{1, 2, 3}))
			}
		}
	}`

	embeddedRequiredValCode = `	if val.Foo == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(` + "`context`" + `, "foo"))
	}
	if val.Foo != nil {
		if val.Foo.Bar != nil {
			if !(*val.Foo.Bar == 1 || *val.Foo.Bar == 2 || *val.Foo.Bar == 3) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError(` + "`" + `context.foo.bar` + "`" + `, *val.Foo.Bar, []interface{}{1, 2, 3}))
			}
		}
	}`

	tagCode = `	if val.__tag__ != nil {
		if val.__tag__.Bar != nil {
			if !(*val.__tag__.Bar == 1 || *val.__tag__.Bar == 2 || *val.__tag__.Bar == 3) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError(` + "`" + `context.foo.bar` + "`" + `, *val.__tag__.Bar, []interface{}{1, 2, 3}))
			}
		}
	}`

	tagChildCode = `	if val.Foo != nil {
		if val.Foo.__tag__ != nil {
			if !(*val.Foo.__tag__ == 1 || *val.Foo.__tag__ == 2 || *val.Foo.__tag__ == 3) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError(` + "`" + `context.foo.bar` + "`" + `, *val.Foo.__tag__, []interface{}{1, 2, 3}))
			}
		}
	}`

	utCode = `	if val.Foo == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(` + "`context`" + `, "foo"))
	}`

	utRequiredCode = `	for _, e := range val.Foo {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}`
)
