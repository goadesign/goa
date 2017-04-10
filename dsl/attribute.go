package dsl

import (
	"fmt"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Attribute describes a field of an object.
//
// An attribute has a name, a type and optionally a default value, an example
// value and validation rules.
//
// The type of an attribute can be one of:
//
// * The primitive types Boolean, Float32, Float64, Int, Int32, Int64, UInt,
//   UInt32, UInt64, String or Bytes.
//
// * A user type defined via the Type function.
//
// * An array defined using the ArrayOf function.
//
// * An map defined using the MapOf function.
//
// * An object defined inline using Attribute to define the type fields
//   recursively.
//
// * The special type Any to indicate that the attribute may take any of the
//   types listed above.
//
// Attribute may appear in MediaType, Type, Attribute or Attributes.
//
// Attribute accepts one to four arguments, the valid usages of the function
// are:
//
//    Attribute(name)       // Attribute of type String with no description, no
//                          // validation, default or example value
//
//    Attribute(name, fn)   // Attribute of type object with inline field
//                          // definitions, description, validations, default
//                          // and/or example value
//
//    Attribute(name, type) // Attribute with no description, no validation,
//                          // no default or example value
//
//    Attribute(name, type, fn) // Attribute with description, validations,
//                              // default and/or example value
//
//    Attribute(name, type, description)     // Attribute with no validation,
//                                           // default or example value
//
//    Attribute(name, type, description, fn) // Attribute with description,
//                                           // validations, default and/or
//                                           // example value
//
// Where name is a string indicating the name of the attribute, type specifies
// the attribute type (see above for the possible values), description a string
// providing a human description of the attribute and fn the defining DSL if
// any.
//
// When defining the type inline using Attribute recursively the function takes
// the second form (name and DSL defining the type). The description can be
// provided using the Description function in this case.
//
// Examples:
//
//    Attribute("name")
//
//    Attribute("driver", Person)         // Use type defined with Type function
//
//    Attribute("driver", "Person")       // May also use the type name
//
//    Attribute("name", String, func() {
//        Pattern("^foo")                 // Adds a validation rule
//    })
//
//    Attribute("driver", Person, func() {
//        Required("name")                // Add required field to list of
//    })                                  // fields already required in Person
//
//    Attribute("name", String, func() {
//        Default("bob")                  // Sets a default value
//    })
//
//    Attribute("name", String, "name of driver") // Sets a description
//
//    Attribute("age", Int32, "description", func() {
//        Minimum(2)                       // Sets both a description and
//                                         // validations
//    })
//
// The definition below defines an attribute inline. The resulting type
// is an object with three attributes "name", "age" and "child". The "child"
// attribute is itself defined inline and has one child attribute "name".
//
//    Attribute("driver", func() {           // Define type inline
//        Description("Composite attribute") // Set description
//
//        Attribute("name", String)          // Child attribute
//        Attribute("age", Int32, func() {   // Another child attribute
//            Description("Age of driver")
//            Default(42)
//            Minimum(2)
//        })
//        Attribute("child", func() {        // Defines a child attribute
//            Attribute("name", String)      // Grand-child attribute
//            Required("name")
//        })
//
//        Required("name", "age")            // List required attributes
//    })
//
func Attribute(name string, args ...interface{}) {
	var parent *design.AttributeExpr
	{
		switch def := eval.Current().(type) {
		case *design.AttributeExpr:
			parent = def
		case design.CompositeExpr:
			parent = def.Attribute()
		default:
			eval.IncompatibleDSL()
			return
		}
		if parent == nil {
			eval.ReportError("invalid syntax, attribute %#v has no parent", name)
			return
		}
		if parent.Type == nil {
			parent.Type = make(design.Object)
		}
		if _, ok := parent.Type.(design.Object); !ok {
			eval.ReportError("can't define child attribute %#v on attribute of type %s", name, parent.Type.Name())
			return
		}
	}

	var attr *design.AttributeExpr
	if parent.Reference != nil {
		if att, ok := design.AsObject(parent.Reference)[name]; ok {
			attr = design.DupAtt(att)
		}
	}

	dataType, description, fn := parseAttributeArgs(attr, args...)
	if attr != nil {
		if description != "" {
			attr.Description = description
		}
		if dataType != nil {
			attr.Type = dataType
		}
	} else {
		attr = &design.AttributeExpr{
			Type:        dataType,
			Description: description,
		}
	}
	attr.Reference = parent.Reference
	if fn != nil {
		eval.Execute(fn, attr)
	}
	if attr.Type == nil {
		// DSL did not contain an "Attribute" declaration
		attr.Type = design.String
	}
	design.AsObject(parent.Type)[name] = attr
}

// Field is syntactic sugar to define an attribute with the "rpc:tag" metadata
// set with the value of the first argument.
//
// Field may appear wherever Attribute can.
// Field takes the same arguments as Attribute with the addition of the tag
// value as first argument.
//
// Example:
//
//     Field(1, "ID", String, func() {
//         Pattern("[0-9]+")
//     })
//
func Field(tag interface{}, name string, args ...interface{}) {
	fn := func() { Metadata("rpc:tag", fmt.Sprintf("%v", tag)) }
	if d, ok := args[len(args)-1].(func()); ok {
		old := fn
		fn = func() { d(); old() }
		args = args[:len(args)-1]
	}
	Attribute(name, append(args, fn)...)
}

// Default sets the default value for an attribute.
func Default(def interface{}) {
	a, ok := eval.Current().(*design.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if a.Type != nil && !a.Type.IsCompatible(def) {
		eval.ReportError("default value %#v is incompatible with attribute of type %s",
			def, design.QualifiedTypeName(a.Type))
		return
	}
	a.SetDefault(def)
}

// Example sets the example of an attribute to be used for the documentation.
// If no example is explicitly provided then a random example is generated
// unless the "swagger:example" metadata is set to "false". See Metadata.
//
// Example may appear in a Attribute expression.
// Example takes one argument: the example value.
//
// Example:
//
//	Attributes(func() {
//		Attribute("ID", Int64, func() {
//			Example(1)
//		})
//	})
//
func Example(exp interface{}) {
	if a, ok := eval.Current().(*design.AttributeExpr); ok {
		if !a.Type.IsCompatible(exp) {
			eval.ReportError("example value %#v is incompatible with attribute of type %s",
				exp, a.Type.Name())
			return
		}
		a.UserExample = exp
	}
}

func parseAttributeArgs(baseAttr *design.AttributeExpr, args ...interface{}) (design.DataType, string, func()) {
	var (
		dataType    design.DataType
		description string
		fn          func()
		ok          bool
	)

	parseDataType := func(expected string, index int) {
		if name, ok2 := args[index].(string); ok2 {
			// Lookup type by name
			if dataType = design.Root.UserType(name); dataType == nil {
				eval.InvalidArgError(expected, args[index])
			}
			return
		}
		if dataType, ok = args[index].(design.DataType); !ok {
			eval.InvalidArgError(expected, args[index])
		}
	}
	parseDescription := func(expected string, index int) {
		if description, ok = args[index].(string); !ok {
			eval.InvalidArgError(expected, args[index])
		}
	}
	parseDSL := func(index int, success, failure func()) {
		if fn, ok = args[index].(func()); ok {
			success()
			return
		}
		failure()
	}

	success := func() {}

	switch len(args) {
	case 0:
		if baseAttr != nil {
			dataType = baseAttr.Type
		} else {
			dataType = design.String
		}
	case 1:
		success = func() {
			if baseAttr != nil {
				dataType = baseAttr.Type
			}
		}
		parseDSL(0, success, func() { parseDataType("type, type name or func()", 0) })
	case 2:
		parseDataType("type or type name", 0)
		parseDSL(1, success, func() { parseDescription("string or func()", 1) })
	case 3:
		parseDataType("type or type name", 0)
		parseDescription("string", 1)
		parseDSL(2, success, func() { eval.InvalidArgError("func()", args[2]) })
	default:
		eval.ReportError("too many arguments in call to Attribute")
	}

	return dataType, description, fn
}
