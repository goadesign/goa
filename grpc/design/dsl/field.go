package dsl

// Field defines the fields of messages.
// A field has a name, a type, a tag and optionally a default value and validation rules.
//
// The type of an field can be one of:
//
// * The primitive types Double, Float, Int32, Int64, UInt32, UInt64, Bool, String or Bytes.
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
// Fields may appear in Message definitions.
//
// Examples:
//
//    Field("name", String, 1)               // Defines a field of type String and tag 1
//
//    Field("name", String, 1, func() {
//        Pattern("^foo")                    // Adds a validation rule to the field
//    })
//
//    Field("name", Operand, 1)              // Defines a field using user type Operand
//
//    Field("name", Int32, 1, func() {
//        Default(42)                        // Defines a default value
//    })
//
//    Field("name", Int32, 1, func() {       // DSL below only applies to Int32 and Int64 types
//        Signed()                           // Generate a signed type
//        Fixed()                            // Generate a fixed protobuf type
//    })                                     // This example generates sfixed32
//
//    Field("name", Int32, 1, "description") // Specifies a description
//
//    Field("name", Int32, 1, "description", func() {
//        Minimum(2)                         // And validation rules
//    })
//
// The valid usages of the Field function are:
//
//    Field(name, type, tag)
//
//    Field(name, type, description, tag)
//
//    Field(name, type, description, tag, dsl)
//
func Field(name string, t DataType, args ...interface{}) {
}
