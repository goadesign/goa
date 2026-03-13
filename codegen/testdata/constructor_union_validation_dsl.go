package testdata

import . "goa.design/goa/v3/dsl"

// ConstructorUnionValidationDSL defines constructor-form unions with
// branch-specific validations.
var ConstructorUnionValidationDSL = func() {
	TextPayload := Type("ConstructorUnionValidationText", func() {
		Attribute("text", String, func() {
			MinLength(3)
		})
		Required("text")
	})
	JSONPayload := Type("ConstructorUnionValidationJSON", func() {
		Attribute("message", String, func() {
			MaxLength(10)
		})
		Required("message")
	})

	_ = Type("ConstructorUnionValidation", func() {
		Attribute("choice", OneOf(TextPayload, JSONPayload))
		Required("choice")
	})
}
