package dsl

import (
	"fmt"

	"github.com/raphael/goa/design"
)

// Attribute defines an attribute type, description and an optional validation DSL.
// When Attribute() is used in an action parameter definition all the arguments are optional and
// the corresponding attribute definition fields are inherited from the resource media type.
// Valid usage:
//
// * Attribute(name string, dataType DataType, description string, dsl func())
//
// * Attribute(name string, dataType DataType, dsl func())
//
// * Attribute(name string, dsl func())
//
// * Attribute(name string)
//
// The following all call this method:
//
//     Attribute("foo", func() {
//         Enum("one", "two")
//     })
//
//     Header("Authorization", String)
//
//     Param("AccountID", Integer, "Account ID")
//
func Attribute(name string, args ...interface{}) *design.AttributeDefinition {
	if parent, ok := attributeDefinition(); ok {
		if parent.Type == nil {
			parent.Type = &Object{}
		}
		var dataType DataType
		var description string
		var dsl func()
		var ok bool
		if len(args) == 1 {
			if dsl, ok = args[0].(func()); !ok {
				invalidArgError("func()", args[0])
			}
		} else if len(args) == 2 {
			if dataType, ok = args[0].(DataType); !ok {
				invalidArgError("DataType", args[0])
			}
			if dsl, ok = args[1].(func()); !ok {
				invalidArgError("func()", args[1])
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
		att := AttributeDefinition{
			Type:        dataType,
			Description: description,
		}
		if dsl != nil {
			executeDSL(dsl, &att)
		}
		parent.Type.(*Object)[name] = &att
	}
}

// Header is an alias to Attribute
func Header(args ...interface{}) *design.AttributeDefinition {
	return Attribute(args...)
}

// Member is an alias to Attribute
func Member(args ...interface{}) *design.AttributeDefinition {
	return Attribute(args...)
}

// Param is an alias to Attribute
func Param(args ...interface{}) *design.AttributeDefinition {
	return Attribute(args...)
}

/* Validation keywords for any instance type */

// http://json-schema.org/latest/json-schema-validation.html#anchor76
func (a *design.AttributeDefinition) Enum(val ...interface{}) *AttributeDefinition {
	a.Validations = append(a.Validations, NewEnumValidation(val))
	return a
}

// Set default value
func (a *design.AttributeDefinition) Default(def interface{}) *AttributeDefinition {
	a.DefaultValue = def
	return a
}

// Set string format
func (a *design.AttributeDefinition) Format(f string) *AttributeDefinition {
	a.Validations = append(a.Validations, NewFormatValidation(f))
	return a
}

// Minimum value validation
func (a *design.AttributeDefinition) Minimum(val int) *AttributeDefinition {
	a.Validations = append(a.Validations, NewMinimumValidation(val))
	return a
}

// Maximum value validation
func (a *design.AttributeDefinition) Maximum(val int) *AttributeDefinition {
	a.Validations = append(a.Validations, NewMaximumValidation(val))
	return a
}

// Minimum length validation
func (a *design.AttributeDefinition) MinLength(val int) *AttributeDefinition {
	a.Validations = append(a.Validations, NewMinLengthValidation(val))
	return a
}

// Maximum length validation
func (a *design.AttributeDefinition) MaxLength(val int) *AttributeDefinition {
	a.Validations = append(a.Validations, NewMaxLengthValidation(val))
	return a
}

// Required properties validation
func (a *design.AttributeDefinition) Required(names ...string) *AttributeDefinition {
	if a.Type.Kind() != ObjectType {
		panic("Required validation must be applied to object types")
	}
	a.Validations = append(a.Validations, NewRequiredValidation(names...))
	return a
}
