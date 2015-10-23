package dsl

import . "github.com/raphael/goa/design"

// SupportedValidationFormats lists the supported formats for use with the
// Format DSL.
var SupportedValidationFormats = []string{
	"cidr",
	"date-time",
	"email",
	"hostname",
	"ipv4",
	"ipv6",
	"mac",
	"regexp",
	"uri",
}

// NewEnumValidation creates a definition for an enum validation.
func NewEnumValidation(val ...interface{}) ValidationDefinition {
	return &EnumValidationDefinition{Values: val}
}

// NewFormatValidation creates a definition for a format validation.
func NewFormatValidation(f string) ValidationDefinition {
	return &FormatValidationDefinition{Format: f}
}

// NewPatternValidation creates a definition for a pattern validation.
func NewPatternValidation(p string) ValidationDefinition {
	return &PatternValidationDefinition{Pattern: p}
}

// NewMinimumValidation creates a definition for a minimum value validation.
func NewMinimumValidation(min int) ValidationDefinition {
	return &MinimumValidationDefinition{Min: min}
}

// NewMaximumValidation creates a definition for a maximum value validation.
func NewMaximumValidation(max int) ValidationDefinition {
	return &MaximumValidationDefinition{Max: max}
}

// NewMinLengthValidation creates a definition for a minimum length validation.
func NewMinLengthValidation(minLength int) ValidationDefinition {
	return &MinLengthValidationDefinition{MinLength: minLength}
}

// NewMaxLengthValidation creates a definition for a maximum length validation.
func NewMaxLengthValidation(maxLength int) ValidationDefinition {
	return &MaxLengthValidationDefinition{MaxLength: maxLength}
}

// NewRequiredValidation creates a definition for a required fields validation.
func NewRequiredValidation(names ...string) ValidationDefinition {
	return &RequiredValidationDefinition{Names: names}
}
