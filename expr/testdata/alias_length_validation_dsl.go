package testdata

import . "goa.design/goa/v3/dsl"

// AliasLengthValidationDSL defines an alias type with length validation for testing.
var AliasLengthValidationDSL = func() {
	var _ = Type("ValidatedString", String, func() {
		MinLength(5)
		MaxLength(10)
	})
}

// AliasArrayLengthValidationDSL defines an alias array type with length validation for testing.
var AliasArrayLengthValidationDSL = func() {
	var _ = Type("StringArray", ArrayOf(String), func() {
		MinLength(2)
		MaxLength(5)
	})
}
