package dsl

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/raphael/goa/design"
)

// Attribute implements the attribute definition DSL. An attribute describes a data structure
// recursively. Attributes are used for describing request headers, parameters and payloads -
// response bodies and headers - media types and types. An attribute definition is recursive:
// attributes may include other attributes. At the basic level an attribute has a name,
// a type and optionally a default value and validation rules. The type of an attribute can be one of:
//
// * The primitive types Boolean, Integer, Number or String.
//
// * A type defined via the Type function.
//
// * A media type defined via the MediaType function.
//
// * An object described recursively with child attributes.
//
// * An array defined using the ArrayOf function.
//
// * An hashmap defined using the HashOf function.
//
// Attributes can be defined using the Attribute, Param, Member or Header functions depending
// on where the definition appears. The syntax for all these DSL is the same.
// Here are some examples:
//
//	 Attribute("name")             //  Defines an attribute of type String
//
//	 Attribute("name", func() {
//	 	Pattern("^foo")        // Adds a validation rule to the attribute
//	 })
//
//	 Attribute("name", Integer)    // Defines an attribute of type Integer
//
//	 Attribute("name", Integer, func() {
//	 	Default(42)            // With a default value
//	 })
//
//	 Attribute("name", Integer, "description") // Specifies a description
//
//	 Attribute("name", Integer, "description", func() {
//	 	Enum(1, 2)             // And validation rules
//	 })
//
// Nested attributes:
//
//	 Attribute("nested", func() {
//	 	Description("description")
//	 	Attribute("child")
//	 	Attribute("child2", func() {
//	 		// ....
//	 	})
//              Required("child")
//	 })
//
// Here are all the valid usage of the Attribute function:
//
//	 Attribute(name string, dataType DataType, description string, dsl func())
//
//	 Attribute(name string, dataType DataType, description string)
//
//	 Attribute(name string, dataType DataType, dsl func())
//
//	 Attribute(name string, dataType DataType)
//
//	 Attribute(name string, dsl func()) /* dataType is String or Object (if DSL defines child attributes) */
//
//	 Attribute(name string) /* dataType is String */
func Attribute(name string, args ...interface{}) {
	var parent *design.AttributeDefinition
	if at, ok := attributeDefinition(false); ok {
		parent = at
	} else if mt, ok := mediaTypeDefinition(true); ok {
		parent = mt.AttributeDefinition
	}
	if parent != nil {
		if parent.Type == nil {
			parent.Type = design.Object{}
		} else if _, ok := parent.Type.(design.Object); !ok {
			ReportError("can't define child attributes on attribute of type %s", parent.Type.Name())
			return
		}
		var baseAttr *design.AttributeDefinition
		if parent.Reference != nil {
			for n, att := range parent.Reference.ToObject() {
				if n == name {
					baseAttr = att.Dup()
					break
				}
			}
		}
		var dataType design.DataType
		var description string
		var dsl func()
		var ok bool
		if len(args) == 0 {
			if baseAttr != nil {
				dataType = baseAttr.Type
			} else {
				dataType = design.String
			}
		} else if len(args) == 1 {
			if dsl, ok = args[0].(func()); !ok {
				if dataType, ok = args[0].(design.DataType); !ok {
					invalidArgError("DataType or func()", args[0])
				}
			} else if baseAttr != nil {
				dataType = baseAttr.Type
			}
		} else if len(args) == 2 {
			if dataType, ok = args[0].(design.DataType); !ok {
				invalidArgError("DataType", args[0])
			}
			if dsl, ok = args[1].(func()); !ok {
				if description, ok = args[1].(string); !ok {
					invalidArgError("string or func()", args[1])
				}
			}
		} else if len(args) == 3 {
			if dataType, ok = args[0].(design.DataType); !ok {
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
		var att *design.AttributeDefinition
		if baseAttr != nil {
			att = baseAttr
			if description != "" {
				att.Description = description
			}
			if dataType != nil {
				att.Type = dataType
			}
		} else {
			att = &design.AttributeDefinition{
				Type:        dataType,
				Description: description,
			}
		}
		att.Reference = parent.Reference
		if dsl != nil {
			executeDSL(dsl, att)
		}
		if att.Type == nil {
			// DSL did not contain an "Attribute" declaration
			att.Type = design.String
		}
		parent.Type.(design.Object)[name] = att
	}
}

// Header is an alias of Attribute.
func Header(name string, args ...interface{}) {
	Attribute(name, args...)
}

// Member is an alias of Attribute.
func Member(name string, args ...interface{}) {
	Attribute(name, args...)
}

// Param is an alias of Attribute.
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

// Enum adds a "enum" validation to the attribute.
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
			a.Validations = append(a.Validations, &design.EnumValidationDefinition{Values: val})
		}
	}
}

// SupportedValidationFormats lists the supported formats for use with the
// Format DSL.
var SupportedValidationFormats = []string{
	"cidr",
	"date-time",
	"email",
	"hostname",
	"ipv4",
	"ipv6",
	"mac",
	"regexp",
	"uri",
}

// Format adds a "format" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor104.
// The formats supported by goa are:
//
// "date-time": RFC3339 date time
//
// "email": RFC5322 email address
//
// "hostname": RFC1035 internet host name
//
// "ipv4" and "ipv6": RFC2373 IPv4 and IPv6 address
//
// "uri": RFC3986 URI
//
// "mac": IEEE 802 MAC-48, EUI-48 or EUI-64 MAC address
//
// "cidr": RFC4632 or RFC4291 CIDR notation IP address
//
// "regexp": RE2 regular expression
func Format(f string) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != design.StringKind {
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
				a.Validations = append(a.Validations, &design.FormatValidationDefinition{Format: f})
			}
		}
	}
}

// Pattern adds a "pattern" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor33.
func Pattern(p string) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != design.StringKind {
			incompatibleAttributeType("pattern", a.Type.Name(), "a string")
		} else {
			_, err := regexp.Compile(p)
			if err != nil {
				ReportError("invalid pattern %#v, %s", p, err)
			} else {
				a.Validations = append(a.Validations, &design.PatternValidationDefinition{Pattern: p})
			}
		}
	}
}

// Minimum adds a "minimum" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor21.
func Minimum(val interface{}) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != design.IntegerKind && a.Type.Kind() != design.NumberKind {
			incompatibleAttributeType("minimum", a.Type.Name(), "an integer or a number")
		} else {
			var f float64
			switch v := val.(type) {
			case float32:
				f = float64(v)
			case float64:
				f = v
			case int:
				f = float64(v)
			case int8:
				f = float64(v)
			case int16:
				f = float64(v)
			case int32:
				f = float64(v)
			case int64:
				f = float64(v)
			case uint8:
				f = float64(v)
			case uint16:
				f = float64(v)
			case uint32:
				f = float64(v)
			case uint64:
				f = float64(v)
			case string:
				var err error
				f, err = strconv.ParseFloat(v, 64)
				if err != nil {
					ReportError("invalid number value %#v", v)
					return
				}
			default:
				ReportError("invalid number value %#v", v)
				return
			}
			a.Validations = append(a.Validations, &design.MinimumValidationDefinition{Min: f})
		}
	}
}

// Maximum adds a "maximum" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor17.
func Maximum(val interface{}) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != design.IntegerKind && a.Type.Kind() != design.NumberKind {
			incompatibleAttributeType("maximum", a.Type.Name(), "an integer or a number")
		} else {
			var f float64
			switch v := val.(type) {
			case float32:
				f = float64(v)
			case float64:
				f = v
			case int:
				f = float64(v)
			case int8:
				f = float64(v)
			case int16:
				f = float64(v)
			case int32:
				f = float64(v)
			case int64:
				f = float64(v)
			case uint8:
				f = float64(v)
			case uint16:
				f = float64(v)
			case uint32:
				f = float64(v)
			case uint64:
				f = float64(v)
			case string:
				var err error
				f, err = strconv.ParseFloat(v, 64)
				if err != nil {
					ReportError("invalid number value %#v", v)
					return
				}
			default:
				ReportError("invalid number value %#v", v)
				return
			}
			a.Validations = append(a.Validations, &design.MaximumValidationDefinition{Max: f})
		}
	}
}

// MinLength adss a "minItems" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor45.
func MinLength(val int) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != design.StringKind && a.Type.Kind() != design.ArrayKind {
			incompatibleAttributeType("minimum length", a.Type.Name(), "a string or an array")
		} else {
			a.Validations = append(a.Validations, &design.MinLengthValidationDefinition{MinLength: val})
		}
	}
}

// MaxLength adss a "maxItems" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor42.
func MaxLength(val int) {
	if a, ok := attributeDefinition(true); ok {
		if a.Type != nil && a.Type.Kind() != design.StringKind && a.Type.Kind() != design.ArrayKind {
			incompatibleAttributeType("maximum length", a.Type.Name(), "a string or an array")
		} else {
			a.Validations = append(a.Validations, &design.MaxLengthValidationDefinition{MaxLength: val})
		}
	}
}

// Required adds a "required" validation to the attribute.
// See http://json-schema.org/latest/json-schema-validation.html#anchor61.
func Required(names ...string) {
	var at *design.AttributeDefinition
	if a, ok := attributeDefinition(false); ok {
		at = a
	} else if mt, ok := mediaTypeDefinition(true); ok {
		at = mt.AttributeDefinition
	} else {
		return
	}
	if at.Type != nil && at.Type.Kind() != design.ObjectKind {
		incompatibleAttributeType("required", at.Type.Name(), "an object")
	} else {
		at.Validations = append(at.Validations, &design.RequiredValidationDefinition{Names: names})
	}
}

// incompatibleAttributeType reports an error for validations defined on
// incompatible attributes (e.g. max value on string).
func incompatibleAttributeType(validation, actual, expected string) {
	ReportError("invalid %s validation definition: attribute must be %s (but type is %s)",
		validation, expected, actual)
}
