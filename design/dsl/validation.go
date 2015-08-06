package dsl

import . "github.com/raphael/goa/design"

// NewEnumValidation creates a definition for an enum validation.
func NewEnumValidation(val ...interface{}) ValidationDefinition {
	return &EnumValidationDefinition{Values: val}
}

// NewFormatValidation creates a definition for a format validation.
func NewFormatValidation(f string) ValidationDefinition {
	return &FormatValidationDefinition{Format: f}
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
