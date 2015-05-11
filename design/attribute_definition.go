package design

// AttributeDefinition defines an object member with optional description, default value and
// validations.
type AttributeDefinition struct {
	Type         DataType     // Attribute type
	Description  string       // Optional description
	Validations  []Validation // Optional validation functions
	DefaultValue interface{}  // Optional member default value
}

// Load calls Load on the attibute type then runs any member validation.
func (a *AttributeDefinition) Load(name string, value interface{}) (interface{}, error) {
	res, err := a.Type.Load(value)
	if err != nil {
		return nil, err
	}
	for _, validation := range a.Validations {
		if err := validation(name, res); err != nil {
			return nil, err
		}
	}
	return res, nil
}
