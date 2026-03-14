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

// ConstructorUnionNestedValidationDSL defines nested constructor-form unions
// with validations on both the outer branch and the nested active branch.
var ConstructorUnionNestedValidationDSL = func() {
	InnerText := Type("ConstructorUnionNestedValidationInnerText", func() {
		Attribute("text", String, func() {
			MinLength(5)
		})
		Required("text")
	})
	InnerJSON := Type("ConstructorUnionNestedValidationInnerJSON", func() {
		Attribute("message", String, func() {
			MaxLength(8)
		})
		Required("message")
	})
	OuterText := Type("ConstructorUnionNestedValidationOuterText", func() {
		Attribute("label", String, func() {
			MinLength(3)
		})
		Attribute("inner", OneOf(InnerText, InnerJSON))
		Required("label", "inner")
	})
	OuterJSON := Type("ConstructorUnionNestedValidationOuterJSON", func() {
		Attribute("count", Int, func() {
			Minimum(1)
		})
		Required("count")
	})

	_ = Type("ConstructorUnionNestedValidation", func() {
		Attribute("choice", OneOf(OuterText, OuterJSON))
		Required("choice")
	})
}
