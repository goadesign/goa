package dsl

import "github.com/goadesign/goa/design"

// ArrayOf creates a repeated type from its element type.
//
// ArrayOf may be used wherever types can.
// ArrayOf takes one argument: the type being repeated.
//
// Example:
//
//     var SumPayload = Message("SumPayload", func() {
//        Description("Message sent to add endpoint")
//        Field(1, "a", ArrayOf(Int32)) // Defines int32 repeated field "a"
//     })
//
func ArrayOf(t design.DataType) *design.Array {
}

// MapOf creates a map from its key and element types.
//
// MapOf may be used wherever types can.
// MapOf takes two arguments: the key and value types.
//
// Example:
//
//     var SumPayload = Message("SumPayload", func() {
//        Description("Message sent to add endpoint")
//        Field(1, "a", MapOf(String, Int32)) // Defines map<string, int32> field "a"
//     })
//
func MapOf(k, v design.DataType) *design.Hash {
	return &design.Hash{KeyType: &kat, ElemType: &vat}
}
