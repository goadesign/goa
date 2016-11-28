package dsl

// Signed signifies that the Int32 or Int64 type should be mapped to sint32 or sint64 unless Fixed()
// is also called in which case they map to sfixed32 and sfixed64 respectively.
//
// Signed may be used in a Field definition.
// Signed takes no argument.
//
// Example:
//
//    Field(1, "name", Int32, func() {
//        Signed()
//    }) // Produces a field of type sint32
//
func Signed() {
}

// Fixed signifies that the Int32 or Int64 type should be mapped to fixed32 or fixed64 unless
// Fixed() is also called in which case they map to sfixed32 and sfixed64 respectively.
//
// Fixed may be used in a Field definition.
// Fixed takes no argument.
//
// Example:
//
//    Field(1, "name", Int32, func() {
//        Fixed()
//    }) // Produces a field of type fixed32
//
func Fixed() {
}

// Enum defines an enum type.
//
// Enum is a top level definition.
// Enum takes two arguments: the enum name and the defining DSL.
//
// Example:
//
//     var OperationEnum = Enum("operations", func() {
//         Description("All possible operations")
//         Value("UNKNOWN", 0)
//         Value("ADDITION", 1)
//         Value("SUBSTRACTION", 2)
//     })
//
//     var OperationEnum = Enum("operations", func() {
//         Description("All possible operations")
//         Proto("adder.proto", "adder.OperationEnum") // Use enum definition "adder.OperationEnum"
//                                                     // from protobuf file "adder.proto"
//     })
//
func Enum(name string, dsl func()) {
}
