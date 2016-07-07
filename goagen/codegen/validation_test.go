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
				code = codegen.RecursiveChecker(att, false, false, false, target, context, 1, false)
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

			Context("of min length 1", func() {
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
					Ω(code).Should(Equal(minLengthValCode))
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

	minLengthValCode = `	if val != nil {
		if len(val) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError(` + "`" + `context` + "`" + `, val, len(val), 1, true))
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
)
