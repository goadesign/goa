package dsl

import (
	"fmt"

	"goa.design/goa/design"
	"goa.design/goa/eval"
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
// Attribute must appear in ResultType, Type, Attribute or Attributes.
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
			parent.Type = &design.Object{}
		}
		if _, ok := parent.Type.(*design.Object); !ok {
			eval.ReportError("can't define child attribute %#v on attribute of type %s", name, parent.Type.Name())
			return
		}
	}

	var attr *design.AttributeExpr
	{
		for _, ref := range parent.References {
			if att := design.AsObject(ref).Attribute(name); att != nil {
				attr = design.DupAtt(att)
				break
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
		attr.References = parent.References
		attr.Bases = parent.Bases
		if fn != nil {
			eval.Execute(fn, attr)
		}
		if attr.Type == nil {
			// DSL did not contain an "Attribute" declaration
			attr.Type = design.String
		}
	}

	parent.Type.(*design.Object).Set(name, attr)
}

// Field is syntactic sugar to define an attribute with the "rpc:tag" metadata
// set with the value of the first argument.
//
// Field must appear wherever Attribute can.
//
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

// Example provides an example value for a type, a parameter, a header or any
// attribute. Example supports two syntaxes: one syntax accepts two arguments
// where the first argument is a summary describing the example and the second a
// value provided directly or via a DSL which may also specify a long
// description. The other syntax accepts a single argument and is equivalent to
// using the first syntax where the summary is the string "default".
//
// If no example is explicitly provided in an attribute expression then a random
// example is generated unless the "swagger:example" metadata is set to "false".
// See Metadata.
//
// Example must appear in a Attributes or Attribute expression DSL.
//
// Example takes one or two arguments: an optional summary and the example value
// or defining DSL.
//
// Examples:
//
//	Params(func() {
//		Param("ZipCode:zip-code", String, "Zip code filter", func() {
//			Example("Santa Barbara", "93111")
//			Example("93117") // same as Example("default", "93117")
//		})
//	})
//
//	Attributes(func() {
//		Attribute("ID", Int64, "ID is the unique bottle identifier")
//		Example("The first bottle", func() {
//			Description("This bottle has an ID set to 1")
//			Value(Val{"ID": 1})
//		})
//		Example("Another bottle", func() {
//			Description("This bottle has an ID set to 5")
//			Value(Val{"ID": 5})
//		})
//	})
//
func Example(args ...interface{}) {
	if len(args) == 0 {
		eval.ReportError("not enough arguments")
		return
	}
	if len(args) > 2 {
		eval.ReportError("too many arguments")
		return
	}
	var (
		summary string
		arg     interface{}
	)
	if len(args) == 1 {
		summary = "default"
		arg = args[0]
	} else {
		var ok bool
		summary, ok = args[0].(string)
		if !ok {
			eval.InvalidArgError("summary (string)", summary)
			return
		}
		arg = args[1]
	}
	if a, ok := eval.Current().(*design.AttributeExpr); ok {
		ex := &design.ExampleExpr{Summary: summary}
		if dsl, ok := arg.(func()); ok {
			eval.Execute(dsl, ex)
		} else {
			ex.Value = arg
		}
		if ex.Value == nil {
			eval.ReportError("example value is missing")
			return
		}
		if a.Type != nil && !a.Type.IsCompatible(ex.Value) {
			eval.ReportError("example value %#v is incompatible with attribute of type %s",
				ex.Value, a.Type.Name())
			return
		}
		a.UserExamples = append(a.UserExamples, ex)
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
