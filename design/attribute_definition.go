package design

// AttributeDefinition defines an object member with optional description, default value and
// validations.
type AttributeDefinition struct {
	Type         DataType                // Attribute type
	Description  string                  // Optional description
	Validations  []*ValidationDefinition // Optional validation functions
	DefaultValue interface{}             // Optional member default value
}
