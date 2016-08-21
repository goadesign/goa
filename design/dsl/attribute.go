package dsl

// Attribute defines the fields of composite types.
// A field has a name, a type and optionally a default value and validation rules.
//
// The type of an field can be one of:
//
// * The primitive types Double, Float, Int32, Int64, UInt32, UInt64, Bool, String or Bytes.
//
// * A user type defined via the Type function.
//
// * An array defined using the ArrayOf function.
//
// * An map defined using the MapOf function.
//
// * The special type Any to indicate that the field may take any of the types listed above.
//
// Attributes may appear in Message definitions.
//
// Examples:
//
//    Attribute("name", String)           // Defines a field of type String
//
//    Attribute("name", String, func() {
//        Pattern("^foo")                 // Adds a validation rule to the field
//    })
//
//    Attribute("name", Int32)            // Defines a field of type Int32
//
//    Attribute("name", Int32, func() {
//        Default(42)                     // With a default value
//    })
//
//    Attribute("name", Int32, "description") // Specifies a description
//
//    Attribute("name", Int32, "description", func() {
//        Minimum(2)                      // And validation rules
//    })
//
// The valid usages of the Attribute function are:
//
//    Attribute(name, type)
//
//    Attribute(name, type, description)
//
//    Attribute(name, type, description, dsl)
//
func Attribute(name string, t DataType, args ...interface{}) {
}
