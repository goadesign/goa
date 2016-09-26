package dsl

import (
	"strconv"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/eval"
)

// Field defines the fields of messages.
// A field has a name, a type, a tag and optionally a default value and validation rules.
//
// The type of an field can be one of:
//
// * The primitive types Int32, Int64, UInt32, UInt64, Float32, Float64, Bool, String or Bytes.
//   See Signed() and Fixed() for defining fields with signed and/or fixed integer types.
//
// * A message defined via the Message function.
//
// * An enum defined via the Enum function.
//
// * An array defined using the ArrayOf function for repeated fields.
//
// * An map defined using the MapOf function.
//
// * The special type Any to indicate that the field may take any of the types listed above.
//
// The valid usages of the Field function are:
//
//    Field(tag, name, type)
//
//    Field(tag, name, type, description)
//
//    Field(tag, name, type, description, dsl)
//
// Fields may appear in Message definitions.
//
// Examples:
//
//    Field(1, "name", String)               // Defines a field of type String and tag 1
//
//    Field(1, "name", String, func() {
//        Pattern("^foo")                    // Adds a validation rule to the field
//    })
//
//    Field(1, "name", Operand)              // Defines a field using user type Operand
//
//    Field(1, "name", Int32, func() {
//        Default(42)                        // Defines a default value
//    })
//
//    Field(1, "name", Int32, func() {       // DSL below only applies to Int32 and Int64 types
//        Signed()                           // Generate a signed type
//        Fixed()                            // Generate a fixed protobuf type
//    })                                     // This example generates sfixed32
//
//    Field(1, "name", Int32, "description") // Specifies a description
//
//    Field(1, "name", Int32, "description", func() {
//        Minimum(2)                         // And validation rules
//    })
//
func Field(tag uint, name string, t design.DataType, args ...interface{}) {
	if tag == 0 {
		eval.ReportError("field tag must be at least 1")
		return
	}
	ds, ok := args[len(args)-1].(func())
	args[len(args)-1] = func() {
		if ok {
			ds()
		}
		dsl.Metadata("grpc:tag", strconv.Itoa(int(tag)))
	}
	dsl.Attribute(name, append([]interface{}{t}, args...)...)
}
