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

	// Versioned is implemented by potentially versioned definitions such as API resources.
	Versioned interface {
		Definition
		// Versions returns an array of supported versions if the object is versioned, nil
		// othewise.
		Versions() []string
		// SupportsVersion returns true if the object supports the given version.
		SupportsVersion(ver string) bool
		// SupportsNoVersion returns true if the object is unversioned.
		SupportsNoVersion() bool
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

	// ValidationDefinition is the common interface for all validation data structures.
	// It doesn't expose any method and simply exists to help with documentation.
	ValidationDefinition interface {
		Definition
	}

	// EnumValidationDefinition represents an enum validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor76.
	EnumValidationDefinition struct {
		Values []interface{}
	}

	// FormatValidationDefinition represents a format validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor104.
	FormatValidationDefinition struct {
		Format string
	}

	// PatternValidationDefinition represents a pattern validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor33
	PatternValidationDefinition struct {
		Pattern string
	}

	// MinimumValidationDefinition represents an minimum value validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor21.
	MinimumValidationDefinition struct {
		Min float64
	}

	// MaximumValidationDefinition represents a maximum value validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor17.
	MaximumValidationDefinition struct {
		Max float64
	}

	// MinLengthValidationDefinition represents an minimum length validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor29.
	MinLengthValidationDefinition struct {
		MinLength int
	}

	// MaxLengthValidationDefinition represents an maximum length validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor26.
	MaxLengthValidationDefinition struct {
		MaxLength int
	}

	// RequiredValidationDefinition represents a required validation as described at
	// http://json-schema.org/latest/json-schema-validation.html#anchor61.
	RequiredValidationDefinition struct {
		Names []string
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
func (v *EnumValidationDefinition) Context() string {
	return "enum validation"
}

// Context returns the generic definition name used in error messages.
func (f *FormatValidationDefinition) Context() string {
	return "format validation"
}

// Context returns the generic definition name used in error messages.
func (f *PatternValidationDefinition) Context() string {
	return "pattern validation"
}

// Context returns the generic definition name used in error messages.
func (m *MinimumValidationDefinition) Context() string {
	return "min value validation"
}

// Context returns the generic definition name used in error messages.
func (m *MaximumValidationDefinition) Context() string {
	return "max value validation"
}

// Context returns the generic definition name used in error messages.
func (m *MinLengthValidationDefinition) Context() string {
	return "min length validation"
}

// Context returns the generic definition name used in error messages.
func (m *MaxLengthValidationDefinition) Context() string {
	return "max length validation"
}

// Context returns the generic definition name used in error messages.
func (r *RequiredValidationDefinition) Context() string {
	return "required field validation"
}
