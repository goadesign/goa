package dsl

// Message describes a gRPC message.
//
// Message is a top level definition.
// Message takes two arguments: the message name and the defining DSL.
//
// Examples:
//
//     var SumPayload = Message("SumPayload", func() {
//        Description("Message sent to add endpoint")
//        Field("a", Int32, "'a' operand", 1)   // Defines int32 field "a" with tag 1 with
//                                              // optional description
//        Field("b", Int32, 2)
//        Field("ops", MapOf(String, Int32), 3) // Defines map<string, int32> field "ops"
//        Field("c", SumMod, 4)                 // Defines fields using embedded message
//        Field("operands", ArrayOf(Int32), 5)  // Defines repeated int32 field
//        Reserved(6,11,To(15,90))              // Defines reserved tags
//        Reserved("c", "C")                    // Defines reserved field names
//        Message("Inner", func() {             // Defines nested message
//            Field("inner", Int32, 7)
//        })
//        Field("in", "SubPayload.Inner", 8)    // Use nested message
//        OneOf("alternatives", func() {        // Defines a oneof field
//            Field("first", Int32, 9)
//            Field("second", String, 10)
//        })
//     })
//
//     var SumPayload = Message("SumPayload", func() {
//        Description("Message sent to add endpoint")
//        Proto("adder.proto", "adder.SumPayload")  // Use message definition "adder.SumPayload"
//     })                                           // from protobuf file "adder.proto"
//
func Message(name string, args ...interface{}) {
}

// Reserved defines one or more reserved tags, or one or more reserved field names.
//
// Reserved may be used in a Message definition.
// Reserved takes one or more arguments: either a list of integer field tags and To calls or a list
// of field names.
//
// Example:
//
//     var SumPayload = Message("SumPayload", func() {
//        Description("Message sent to add endpoint")
//        Field(1, "a", Int32)
//        Reserved(6,11,To(15,90)) // Defines reserved tags
//        Reserved("c", "C")       // Defines reserved field names
//     })
//
func Reserved(args ...interface{}) {
}

// OneOf defines a oneof message field.
//
// OneOf may be used in a Message definition.
// OneOf takes two arguments: the name of the oneof and the defining DSL.
//
// Example:
//
//     var SumPayload = Message("SumPayload", func() {
//        Description("Message sent to add endpoint")
//        Field(1, "a", Int32)
//        OneOf("operands", func() {
//            Field(2, "added", Int32)
//            Field(3, "substracted", Int32)
//        })
//     })
//
func OneOf(name string, dsl func()) {
}

// To creates a range of reserved message tags.
//
// To may be used as an argument of the Reserved function.
// To takes two int arguments: the starting reserved tag and the ending reserved tag which must have
// a greater value.
//
// Example:
//
//     var SumPayload = Message("SumPayload", func() {
//        Description("Message sent to add endpoint")
//        Field(1, "a", Int32)
//        Reserved(6,11,To(15,90)) // Defines reserved tags
//     })
//
func To(from, to int) {
}
