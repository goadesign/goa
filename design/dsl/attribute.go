package dsl

import (
	"github.com/goadesign/goa/eval"
)

// Attribute defines the field of a composite type.
// An attribute has a name, a type and optionally a default value and validation rules.
//
// The type of an attribute can be one of:
//
// * The primitive types Double, Float, Int32, Int64, UInt32, UInt64, Bool, String or Bytes.
//
// * A user type defined via the Type function.
//
// * An array defined using the ArrayOf function.
//
// * An map defined using the MapOf function.
//
// * The special type Any to indicate that the attribute may take any of the types listed above.
//
// The type may also be defined inline using Attribute to define the type fields recursively.
//
// Attribute may appear in Type, Attribute or Attributes.
//
// Attribute accepts two to four arguments, the valid usages of the function are:
//
//    Attribute(name, dsl)  // Defines type inline
//                          // Description and/or validations also in DSL
//
//    Attribute(name, type) // No description and no validation
//
//    Attribute(name, type, dsl) // Description and/or validations in DSL
//
//    Attribute(name, type, description)      // No validations
//
//    Attribute(name, type, description, dsl) // Validations in DSL
//
// Where name is a string indicating the name of the attribute, type specifies the attribute type
// (see above for the possible values), description a string providing a human description of the
// attribute and dsl the defining DSL if any.
//
// When defining the type inline using Attribute recursively the function takes the first form (name
// and DSL defining the type). The description can be provided using the Description function in
// this case.
//
// Examples:
//
//    Attribute("name", String)           // Defines a attribute of type String
//                                        // with no description and no validation
//    Attribute("driver", Person)         // Use type defined with Type function
//
//    Attribute("driver", "Person")       // May also use the type name
//
//    Attribute("name", String, func() {
//        Pattern("^foo")                 // Adds a validation rule
//    })
//
//    Attribute("driver", Person, func() {
//        Required("name")                // Add required field to list of required
//    })                                  // fields already defined in Person
//
//    Attribute("name", String, func() {
//        Default("bob")                  // Sets a default value
//    })
//
//    Attribute("name", String, "name of driver") // Sets a description
//
//    Attribute("age", Int32, "description", func() {
//        Minimum(2)                       // Sets both a description and validations
//    })
//
//    // The definition below defines a composite attribute inline.
//    // The resulting type is an object with three attributes
//    // "name", "age" and "child". The "child" attribute is itself
//    // defined inline and has one child attribute "name".
//    //
//    Attribute("driver", func() {           // Define type inline
//        Description("Composite attribute") // Set description
//
//        Attribute("name", String)          // Child attribute
//        Attribute("age", Int32, func() {   // Another child attribute
//            Description("Age of driver")
//            Default(42)
//            Minimum(2)
//        })
//        Attribute("child", func() {        // Defines a composite child
//            Attribute("name", String)      // Grand-child attribute
//            Required("name")
//        })
//
//        Required("name", "age")            // List attributes that are required
//    })
//
func Attribute(name string, args ...interface{}) {
	var parent *design.AttributeExpr

	switch def := eval.Current().(type) {
	case *design.AttributeExpr:
		parent = def
	case design.CompositeExp:
		parent = def.Attribute()
	default:
		eval.IncompatibleDSL()
		return
	}

	if parent != nil {
		if parent.Type == nil {
			parent.Type = make(design.Object)
		}
		if _, ok := parent.Type.(design.Object); !ok {
			eval.ReportError("can't define child attributes on attribute of type %s", parent.Type.Name())
			return
		}

		var baseAttr *design.AttributeExpr
		if parent.Reference != nil {
			if att, ok := parent.Reference.ToObject()[name]; ok {
				baseAttr = design.DupAtt(att)
			}
		}

		dataType, description, dsl := parseAttributeArgs(baseAttr, args...)
		if baseAttr != nil {
			if description != "" {
				baseAttr.Description = description
			}
			if dataType != nil {
				baseAttr.Type = dataType
			}
		} else {
			baseAttr = &design.AttributeExpr{
				Type:        dataType,
				Description: description,
			}
		}
		baseAttr.Reference = parent.Reference
		if dsl != nil {
			eval.Execute(dsl, baseAttr)
		}
		if baseAttr.Type == nil {
			// DSL did not contain an "Attribute" declaration
			baseAttr.Type = design.String
		}
		parent.Type.(design.Object)[name] = baseAttr
	}
}

func parseAttributeArgs(baseAttr *design.AttributeExpr, args ...interface{}) (design.DataType, string, func()) {
	var (
		dataType    design.DataType
		description string
		dsl         func()
		ok          bool
	)

	parseDataType := func(expected string, index int) {
		if name, ok2 := args[index].(string); ok2 {
			// Lookup type by name
			if dataType, ok = design.Design.Types[name]; !ok {
				if dataType = design.Design.MediaTypeWithIdentifier(name); dataType == nil {
					dslengine.InvalidArgError(expected, args[index])
				}
			}
			return
		}
		if dataType, ok = args[index].(design.DataType); !ok {
			dslengine.InvalidArgError(expected, args[index])
		}
	}
	parseDescription := func(expected string, index int) {
		if description, ok = args[index].(string); !ok {
			dslengine.InvalidArgError(expected, args[index])
		}
	}
	parseDSL := func(index int, success, failure func()) {
		if dsl, ok = args[index].(func()); ok {
			success()
		} else {
			failure()
		}
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
		parseDSL(2, success, func() { dslengine.InvalidArgError("func()", args[2]) })
	default:
		dslengine.ReportError("too many arguments in call to Attribute")
	}

	return dataType, description, dsl
}
