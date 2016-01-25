package codegen_test

import (
	"github.com/goadesign/goa/design"
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
			var validations []design.ValidationDefinition

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
					enumVal := &design.EnumValidationDefinition{
						Values: []interface{}{1, 2, 3},
					}
					validations = []design.ValidationDefinition{enumVal}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(enumValCode))
				})
			})

			Context("of pattern", func() {
				BeforeEach(func() {
					attType = design.String
					patternVal := &design.PatternValidationDefinition{
						Pattern: ".*",
					}
					validations = []design.ValidationDefinition{patternVal}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(patternValCode))
				})
			})

			Context("of min value 0", func() {
				BeforeEach(func() {
					attType = design.Integer
					minVal := &design.MinimumValidationDefinition{
						Min: 0,
					}
					validations = []design.ValidationDefinition{minVal}
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(minValCode))
				})
			})

			Context("of embedded object", func() {
				BeforeEach(func() {
					enumVal := &design.EnumValidationDefinition{
						Values: []interface{}{1, 2, 3},
					}
					ccatt := &design.AttributeDefinition{
						Type:        design.Integer,
						Validations: []design.ValidationDefinition{enumVal},
					}
					catt := &design.AttributeDefinition{
						Type: design.Object{"bar": ccatt},
					}
					attType = design.Object{"foo": catt}
					validations = nil
				})

				It("checks the child object is not nil", func() {
					Ω(code).Should(Equal(embeddedValCode))
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
)
