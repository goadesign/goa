package dsl

import "github.com/raphael/goa/design"

// NewEnumValidation creates a definition for an enum validation.
func NewEnumValidation(val ...interface{}) design.ValidationDefinition {
	return &design.EnumValidationDefinition{Values: val}
}

// NewFormatValidation creates a definition for a format validation.
func NewFormatValidation(f string) design.ValidationDefinition {
	return &design.FormatValidationDefinition{Format: f}
}

// NewMinimumValidation creates a definition for a minimum value validation.
func NewMinimumValidation(min int) design.ValidationDefinition {
	return &design.MinimumValidationDefinition{Min: min}
}

// NewMaximumValidation creates a definition for a maximum value validation.
func NewMaximumValidation(max int) design.ValidationDefinition {
	return &design.MaximumValidationDefinition{Max: max}
}

// NewMinLengthValidation creates a definition for a minimum length validation.
func NewMinLengthValidation(minLength int) design.ValidationDefinition {
	return &design.MinLengthValidationDefinition{MinLength: minLength}
}

// NewMaxLengthValidation creates a definition for a maximum length validation.
func NewMaxLengthValidation(maxLength int) design.ValidationDefinition {
	return &design.MaxLengthValidationDefinition{MaxLength: maxLength}
}

// NewRequiredValidation creates a definition for a required fields validation.
func NewRequiredValidation(names ...string) design.ValidationDefinition {
	return &design.RequiredValidationDefinition{Names: names}
}
