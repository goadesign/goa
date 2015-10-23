package codegen_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/codegen"
	"github.com/raphael/goa/design"
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
			var code string // generated code

			JustBeforeEach(func() {
				att.Type = attType
				att.Validations = validations
				code = codegen.ValidationChecker(att, target)
			})

			Context("of enum", func() {
				BeforeEach(func() {
					attType = design.Integer
					enumVal := &design.EnumValidationDefinition{
						Values: []interface{}{1, 2, 3},
					}
					validations = append(validations, enumVal)
				})

				It("produces the validation go code", func() {
					Ω(code).Should(Equal(enumValCode))
				})
			})
		})
	})
})

const (
	enumValCode = `	if !(val == 1 || val == 2 || val == 3) {
		err = support.InvalidEnumValueError(` + "``" + `, val, []interface{}{1, 2, 3}, err)
	}`
)
