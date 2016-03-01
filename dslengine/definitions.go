package dslengine

import "fmt"

type (

	// Definition is the common interface implemented by all definitions.
	Definition interface {
		// Context is used to build error messages that refer to the definition.
		Context() string
	}

	// DefinitionSet contains DSL definitions that are executed as one unit.
	// The slice elements may implement the Validate an, Source interfaces to enable the
	// corresponding behaviors during DSL execution.
	DefinitionSet []Definition

	// Root is the interface implemented by the DSL root objects held by
	// RootDefinitions.
	// These objects contains all the definition sets created by the DSL and can
	// be passed to the dsl for execution.
	Root interface {
		// IterateSets calls the given iterator passing in each definition set
		// sorted in execution order.
		IterateSets(SetIterator)
	}

	// RootDefinitions is the interface for the object containging all the Roots
	// registered by DSLs and can be passed to the dsl for execution.
	RootDefinitions interface {
		// Register a new root into the list of definitions.
		Register(Root)
		// IterateRoots takes a handler function that will be called with each of the
		// registered Roots. If the handler returns an error the walk will be
		// stopped and the error will be returned by IterateRoots.
		IterateRoots(func(Root) error) error
	}

	// Validate is the interface implemented by definitions that can be validated.
	// Validation is done by the DSL dsl post execution.
	Validate interface {
		Definition
		// Validate returns nil if the definition contains no validation error.
		// The Validate implementation may take advantage of ValidationErrors to report
		// more than one errors at a time.
		Validate() error
	}

	// Source is the interface implemented by definitions that can be initialized via DSL.
	Source interface {
		Definition
		// DSL returns the DSL used to initialize the definition if any.
		DSL() func()
	}

	// Finalize is the interface implemented by definitions that require an additional pass
	// after the DSL has executed (e.g. to merge generated definitions or initialize default
	// values)
	Finalize interface {
		Definition
		// Finalize is run by the DSL runner once the definition DSL has executed and the
		// definition has been validated.
		Finalize()
	}

	// SetIterator is the function signature used to iterate over definition sets with
	// IterateSets.
	SetIterator func(s DefinitionSet) error

	// MetadataDefinition is a set of key/value pairs
	MetadataDefinition map[string][]string

	// TraitDefinition defines a set of reusable properties.
	TraitDefinition struct {
		// Trait name
		Name string
		// Trait DSL
		DSLFunc func()
	}

	// ValidationDefinition contains validation rules for an attribute.
	ValidationDefinition struct {
		// Values represents an enum validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor76.
		Values []interface{}
		// Format represents a format validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor104.
		Format string
		// PatternValidationDefinition represents a pattern validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor33
		Pattern string
		// Minimum represents an minimum value validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor21.
		Minimum *float64
		// Maximum represents a maximum value validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor17.
		Maximum *float64
		// MinLength represents an minimum length validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor29.
		MinLength *int
		// MaxLength represents an maximum length validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor26.
		MaxLength *int
		// Required list the required fields of object attributes as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor61.
		Required []string
	}
)

// Context returns the generic definition name used in error messages.
func (t *TraitDefinition) Context() string {
	if t.Name != "" {
		return fmt.Sprintf("trait %#v", t.Name)
	}
	return "unnamed trait"
}

// DSL returns the initialization DSL.
func (t *TraitDefinition) DSL() func() {
	return t.DSLFunc
}

// Context returns the generic definition name used in error messages.
func (v *ValidationDefinition) Context() string {
	return "validation"
}

// Merge merges other into v.
func (v *ValidationDefinition) Merge(other *ValidationDefinition) {
	if v.Values == nil {
		v.Values = other.Values
	}
	if v.Format == "" {
		v.Format = other.Format
	}
	if v.Pattern == "" {
		v.Pattern = other.Pattern
	}
	if v.Minimum == nil || (other.Minimum != nil && *v.Minimum > *other.Minimum) {
		v.Minimum = other.Minimum
	}
	if v.Maximum == nil || (other.Maximum != nil && *v.Maximum < *other.Maximum) {
		v.Maximum = other.Maximum
	}
	if v.MinLength == nil || (other.MinLength != nil && *v.MinLength > *other.MinLength) {
		v.MinLength = other.MinLength
	}
	if v.MaxLength == nil || (other.MaxLength != nil && *v.MaxLength < *other.MaxLength) {
		v.MaxLength = other.MaxLength
	}
	v.AddRequired(other.Required)
}

// AddRequired merges the required fields from other into v
func (v *ValidationDefinition) AddRequired(required []string) {
	for _, r := range required {
		found := false
		for _, rr := range v.Required {
			if r == rr {
				found = true
				break
			}
		}
		if !found {
			v.Required = append(v.Required, r)
		}
	}
}

// Dup makes a shallow dup of the validation.
func (v *ValidationDefinition) Dup() *ValidationDefinition {
	return &ValidationDefinition{
		Values:    v.Values,
		Format:    v.Format,
		Pattern:   v.Pattern,
		Minimum:   v.Minimum,
		Maximum:   v.Maximum,
		MinLength: v.MinLength,
		MaxLength: v.MaxLength,
		Required:  v.Required,
	}
}
