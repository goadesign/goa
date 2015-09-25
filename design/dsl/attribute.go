package dsl

import (
	"strings"

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
// * Attribute(name string, dsl func()) /* dataType is String or Object (if DSL defines child attributes) */
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
	} else if mt, ok := mediaTypeDefinition(true); ok {
		parent = mt.AttributeDefinition
	}
	if parent != nil {
		if parent.Type == nil {
			parent.Type = Object{}
		} else if _, ok := parent.Type.(Object); !ok {
			ReportError("can't define child attributes on attribute of type %s", parent.Type.Name())
			return
		}
		var dataType DataType
		var description string
		var dsl func()
		var ok bool
		if len(args) == 0 {
			dataType = String
		} else if len(args) == 1 {
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
		} else {
			ReportError("too many arguments in call to Attribute")
		}
		att := AttributeDefinition{
			Type:        dataType,
			Description: description,
		}
		if dsl != nil {
			executeDSL(dsl, &att)
		}
		if att.Type == nil {
			// DSL did not contain an "Attribute" declaration
			att.Type = String
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

// Default sets the default value for an attribute.
func Default(def interface{}) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && !a.Type.IsCompatible(def) {
			ReportError("default value %#v is incompatible with attribute of type %s",
				def, a.Type.Name())
		} else {
			a.DefaultValue = def
		}
	}
}

// Enum defines the possible values for an attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor76.
func Enum(val ...interface{}) {
	if a, ok := attributeDefinition(true); ok {
		ok := true
		for i, v := range val {
			// When can a.Type be nil? glad you asked
			// There are two ways to write an Attribute declaration with the DSL that
			// don't set the type: with one argument - just the name - in which case the type
			// is set to String or with two arguments - the name and DSL. In this latter form
			// the type can end up being either String - if the DSL does not define any
			// attribute - or object if it does.
			// Why allowing this? because it's not always possible to specify the type of an
			// object - an object may just be declared inline to represent a substructure.
			// OK then why not assuming object and not allowing for string? because the DSL
			// where there's only one argument and the type is string implicitely is very
			// useful and common, for example to list attributes that refer to other attributes
			// such as responses that refer to responses defined at the API level or links that
			// refer to the media type attributes. So if the form that takes a DSL always ended
			// up defining an object we'd have a weird situation where one arg is string and
			// two args is object. Breaks the least surprise principle. Soooo long story
			// short the lesser evil seems to be to allow the ambiguity. Also tests like the
			// one below are really a convenience to the user and not a fundamental feature
			// - not checking in the case the type is not known yet is OK.
			if a.Type != nil && !a.Type.IsCompatible(v) {
				ReportError("value %#v at index #d is incompatible with attribute of type %s",
					v, i, a.Type.Name())
				ok = false
			}
		}
		if ok {
			a.Validations = append(a.Validations, NewEnumValidation(val...))
		}
	}
}

// Format sets the string format for an attribute.
func Format(f string) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != StringKind {
			incompatibleAttributeType("format", a.Type.Name(), "a string")
		} else {
			supported := false
			for _, s := range SupportedValidationFormats {
				if s == f {
					supported = true
					break
				}
			}
			if !supported {
				ReportError("unsupported format %#v, supported formats are: %s",
					f, strings.Join(SupportedValidationFormats, ", "))
			} else {
				a.Validations = append(a.Validations, NewFormatValidation(f))
			}
		}
	}
}

// Minimum value validation
func Minimum(val int) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != IntegerKind && a.Type.Kind() != NumberKind {
			incompatibleAttributeType("minimum", a.Type.Name(), "an integer or a number")
		} else {
			a.Validations = append(a.Validations, NewMinimumValidation(val))
		}
	}
}

// Maximum value validation
func Maximum(val int) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != IntegerKind && a.Type.Kind() != NumberKind {
			incompatibleAttributeType("maximum", a.Type.Name(), "an integer or a number")
		} else {
			a.Validations = append(a.Validations, NewMaximumValidation(val))
		}
	}
}

// MinLength validation
func MinLength(val int) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != StringKind && a.Type.Kind() != ArrayKind {
			incompatibleAttributeType("minimum length", a.Type.Name(), "a string or an array")
		} else {
			a.Validations = append(a.Validations, NewMinLengthValidation(val))
		}
	}
}

// MaxLength validation
func MaxLength(val int) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != StringKind && a.Type.Kind() != ArrayKind {
			incompatibleAttributeType("maximum length", a.Type.Name(), "a string or an array")
		} else {
			a.Validations = append(a.Validations, NewMaxLengthValidation(val))
		}
	}
}

// Required properties validation
func Required(names ...string) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != ObjectKind {
			incompatibleAttributeType("required", a.Type.Name(), "an object")
		} else {
			a.Validations = append(a.Validations, NewRequiredValidation(names...))
		}
	}
}

// incompatibleAttributeType reports an error for validations defined on
// incompatible attributes (e.g. max value on string).
func incompatibleAttributeType(validation, actual, expected string) {
	ReportError("invalid %s validation definition: attribute must be %s",
		validation, expected, actual)
}
