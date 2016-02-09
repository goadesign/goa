package codegen_test

import (
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
			var validations []dslengine.ValidationDefinition

			att := new(design.AttributeDefinition)
			target := "val"
			context := "context"
			var code string // generated code

			JustBeforeEach(func() {
				att.Type = attType
				att.Validations = validations
				code = codegen.RecursiveChecker(att, false, false, target, context, 1)
			})

			Context("of enum", func() {
				BeforeEach(func() {
					attType = design.Integer
					enumVal := &dslengine.EnumValidationDefinition{
						Values: []interface{}{1, 2, 3},
					}
					validations = []dslengine.ValidationDefinition{enumVal}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(enumValCode))
				})
			})

			Context("of pattern", func() {
				BeforeEach(func() {
					attType = design.String
					patternVal := &dslengine.PatternValidationDefinition{
						Pattern: ".*",
					}
					validations = []dslengine.ValidationDefinition{patternVal}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(patternValCode))
				})
			})

			Context("of min value 0", func() {
				BeforeEach(func() {
					attType = design.Integer
					minVal := &dslengine.MinimumValidationDefinition{
						Min: 0,
					}
					validations = []dslengine.ValidationDefinition{minVal}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(minValCode))
				})
			})

			Context("of embedded object", func() {
				BeforeEach(func() {
					enumVal := &dslengine.EnumValidationDefinition{
						Values: []interface{}{1, 2, 3},
					}
					ccatt := &design.AttributeDefinition{
						Type:        design.Integer,
						Validations: []dslengine.ValidationDefinition{enumVal},
					}
					catt := &design.AttributeDefinition{
						Type: design.Object{"bar": ccatt},
					}

					attType = design.Object{"foo": catt}
				})
				Context("and the parent is optional", func() {
					BeforeEach(func() {
						validations = nil
					})
					It("checks the child & parent object are not nil", func() {
						Ω(code).Should(Equal(embeddedValCode))
					})
				})
				Context("and the parent is required", func() {
					BeforeEach(func() {
						validations = []dslengine.ValidationDefinition{&dslengine.RequiredValidationDefinition{
							Names: []string{"foo"},
						}}
					})
					It("checks the child & parent object are not nil", func() {
						Ω(code).Should(Equal(embeddedRequiredValCode))
					})
				})
			})

		})
	})
})

const (
	enumValCode = `	if val != nil {
		if !(*val == 1 || *val == 2 || *val == 3) {
			err = goa.InvalidEnumValueError(` + "`context`" + `, *val, []interface{}{1, 2, 3}, err)
		}
	}`

	patternValCode = `	if val != nil {
		if ok := goa.ValidatePattern(` + "`.*`" + `, *val); !ok {
			err = goa.InvalidPatternError(` + "`context`" + `, *val, ` + "`.*`" + `, err)
		}
	}`

	minValCode = `	if val != nil {
		if *val < 0 {
			err = goa.InvalidRangeError(` + "`" + `context` + "`" + `, *val, 0, true, err)
		}
	}`

	embeddedValCode = `	if val.Foo != nil {
		if val.Foo.Bar != nil {
			if !(*val.Foo.Bar == 1 || *val.Foo.Bar == 2 || *val.Foo.Bar == 3) {
				err = goa.InvalidEnumValueError(` + "`" + `context.foo.bar` + "`" + `, *val.Foo.Bar, []interface{}{1, 2, 3}, err)
			}
		}
	}`

	embeddedRequiredValCode = `	if val.Foo == nil {
		err = goa.MissingAttributeError(` + "`context`" + `, "foo", err)
	}

	if val.Foo != nil {
	if val.Foo.Bar != nil {
		if !(*val.Foo.Bar == 1 || *val.Foo.Bar == 2 || *val.Foo.Bar == 3) {
			err = goa.InvalidEnumValueError(` + "`" + `context.foo.bar` + "`" + `, *val.Foo.Bar, []interface{}{1, 2, 3}, err)
		}
	}
	}`
)
