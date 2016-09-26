package dsl

// Proto makes it possible to import definitions from protobuf files.
//
// Proto may be used in a Service, Endpoint or Message definition.
// Proto takes two arguments: the filename of the protobuf file relative to the design file calling
// Proto and the defining DSL.
//
// Example:
//
//     var SumPayload = Message("SumPayload", func() {
//        Description("Message sent to add endpoint")
//        Proto("adder.proto", "adder.SumPayload")  // Use message definition "adder.SumPayload"
//     })                                           // from protobuf file "adder.proto"
//
func Proto(filename string, dsl func()) {
}
