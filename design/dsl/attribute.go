package dsl

import (
	"fmt"

	. "github.com/raphael/goa/design"
)

// Attribute defines an attribute type, description and an optional validation DSL.
// When Attribute() is used in an action parameter definition all the arguments are optional and
// the corresponding attribute definition fields are inherited from the resource media type
// attribute of the same name.
// Valid usage:
//
// * Attribute(name string, dataType DataType, description string, dsl func())
//
// * Attribute(name string, dataType DataType, description string)
//
// * Attribute(name string, dataType DataType, dsl func())
//
// * Attribute(name string, dataType DataType)
//
// * Attribute(name string, dsl func()) /* dataType is String */
//
// * Attribute(name string) /* dataType is String */
//
// The following all call this method:
//
//     Attribute("foo", func() {
//         Enum("one", "two")
//     })
//
//     Header("Authorization")
//
//     Param("AccountID", Integer, "Account ID")
//
func Attribute(name string, args ...interface{}) {
	var parent *AttributeDefinition
	if at, ok := attributeDefinition(false); ok {
		parent = at
	} else {
		if mt, ok := mediaTypeDefinition(true); ok {
			parent = mt.AttributeDefinition
		}
	}
	if parent != nil {
		if parent.Type == nil {
			parent.Type = Object{}
		}
		var dataType DataType
		var description string
		var dsl func()
		var ok bool
		if len(args) == 1 {
			if dsl, ok = args[0].(func()); !ok {
				if dataType, ok = args[0].(DataType); !ok {
					invalidArgError("DataType or func()", args[0])
				}
			}
		} else if len(args) == 2 {
			if dataType, ok = args[0].(DataType); !ok {
				invalidArgError("DataType", args[0])
			}
			if dsl, ok = args[1].(func()); !ok {
				if description, ok = args[1].(string); !ok {
					invalidArgError("string or func()", args[1])
				}
			}
		} else if len(args) == 3 {
			if dataType, ok = args[0].(DataType); !ok {
				invalidArgError("DataType", args[0])
			}
			if description, ok = args[1].(string); !ok {
				invalidArgError("string", args[1])
			}
			if dsl, ok = args[2].(func()); !ok {
				invalidArgError("func()", args[2])
			}
		} else if len(args) != 0 {
			appendError(fmt.Errorf("too many arguments in call to Attribute"))
		}
		if dataType == nil {
			dataType = String
		}
		att := AttributeDefinition{
			Type:        dataType,
			Description: description,
		}
		if dsl != nil {
			executeDSL(dsl, &att)
		}
		parent.Type.(Object)[name] = &att
	}
}

// Header is an alias of Attribute
func Header(name string, args ...interface{}) {
	Attribute(name, args...)
}

// Member is an alias of Attribute
func Member(name string, args ...interface{}) {
	Attribute(name, args...)
}

// Param is an alias of Attribute
func Param(name string, args ...interface{}) {
	Attribute(name, args...)
}

// Enum defines the possible values for an attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor76.
func Enum(val ...interface{}) {
	if a, ok := attributeDefinition(true); ok {
		a.Validations = append(a.Validations, NewEnumValidation(val))
	}
}

// Default sets the default value for an attribute.
func Default(def interface{}) {
	if a, ok := attributeDefinition(true); ok {
		a.DefaultValue = def
	}
}

// Format sets the string format for an attribute.
func Format(f string) {
	if a, ok := attributeDefinition(true); ok {
		a.Validations = append(a.Validations, NewFormatValidation(f))
	}
}

// Minimum value validation
func Minimum(val int) {
	if a, ok := attributeDefinition(true); ok {
		a.Validations = append(a.Validations, NewMinimumValidation(val))
	}
}

// Maximum value validation
func Maximum(val int) {
	if a, ok := attributeDefinition(true); ok {
		a.Validations = append(a.Validations, NewMaximumValidation(val))
	}
}

// MinLength validation
func MinLength(val int) {
	if a, ok := attributeDefinition(true); ok {
		a.Validations = append(a.Validations, NewMinLengthValidation(val))
	}
}

// MaxLength validation
func MaxLength(val int) {
	if a, ok := attributeDefinition(true); ok {
		a.Validations = append(a.Validations, NewMaxLengthValidation(val))
	}
}

// Required properties validation
func Required(names ...string) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type.Kind() != ObjectKind {
			a.Validations = append(a.Validations, NewRequiredValidation(names...))
		}
	}
}
