package dsl

// Type describes a user type.
//
// Type is a top level definition.
// Type takes two arguments: the type name and the defining DSL.
//
// Example:
//
//     var SumPayload = Type("SumPayload", func() {
//         Description("Type sent to add endpoint")
//
//         Field("a", String)                 // Defines string field "a"
//         Field("b", Int32, "'b' operand")   // Defines int32 field "b" with description
//         Field("operands", ArrayOf(Int32))  // Defines int32 array field
//         Field("ops", MapOf(String, Int32)) // Defines map<string, int32> field
//         Field("c", SumMod)                 // Defines field using user type
//
//         Required("a")                      // Required fields must be present
//         Required("b", "c")                 // in serialized data.
//     })
//
func Type(name string, dsl func()) {
}
