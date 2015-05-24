package design

// ValidationDefinition is the common interface for all validation data structures.
// It doesn't expose any method and simply exists to help with documentation.
type ValidationDefinition interface{}

// EnumValidationDefinition represents an enum validation as described at
// http://json-schema.org/latest/json-schema-validation.html#anchor76.
type EnumValidationDefinition struct {
	Values []interface{}
}

// NewEnumValidation creates a definition for an enum validation.
func NewEnumValidation(val ...interface{}) ValidationDefinition {
	return &EnumValidationDefinition{Values: val}
}

// FormatValidationDefinition represents a format validation as described at
// http://json-schema.org/latest/json-schema-validation.html#anchor104.
type FormatValidationDefinition struct {
	Format string
}

// NewFormatValidation creates a definition for a format validation.
func NewFormatValidation(f string) ValidationDefinition {
	return &FormatValidationDefinition{Format: f}
}

// MinimumValidationDefinition represents an minimum value validation as described at
// http://json-schema.org/latest/json-schema-validation.html#anchor21.
type MinimumValidationDefinition struct {
	Min int
}

// NewMinimumValidation creates a definition for a minimum value validation.
func NewMinimumValidation(min int) ValidationDefinition {
	return &MinimumValidationDefinition{Min: min}
}

// MaximumValidationDefinition represents a maximum value validation as described at
// http://json-schema.org/latest/json-schema-validation.html#anchor17.
type MaximumValidationDefinition struct {
	Max int
}

// NewMaximumValidation creates a definition for a maximum value validation.
func NewMaximumValidation(max int) ValidationDefinition {
	return &MaximumValidationDefinition{Max: max}
}

// MinLengthValidationDefinition represents an minimum length validation as described at
// http://json-schema.org/latest/json-schema-validation.html#anchor29.
type MinLengthValidationDefinition struct {
	MinLength int
}

// NewMinLengthValidation creates a definition for a minimum length validation.
func NewMinLengthValidation(minLength int) ValidationDefinition {
	return &MinLengthValidationDefinition{MinLength: minLength}
}

// MaxLengthValidationDefinition represents an maximum length validation as described at
// http://json-schema.org/latest/json-schema-validation.html#anchor26.
type MaxLengthValidationDefinition struct {
	MaxLength int
}

// NewMaxLengthValidation creates a definition for a maximum length validation.
func NewMaxLengthValidation(maxLength int) ValidationDefinition {
	return &MaxLengthValidationDefinition{MaxLength: maxLength}
}

// RequiredValidationDefinition represents a required validation as described at
// http://json-schema.org/latest/json-schema-validation.html#anchor61.
type RequiredValidationDefinition struct {
	Names []string
}

// NewRequiredValidation creates a definition for a required fields validation.
func NewRequiredValidation(names ...string) ValidationDefinition {
	return &RequiredValidationDefinition{Names: names}
}
