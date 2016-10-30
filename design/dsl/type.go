package dsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/eval"
)

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
//         Attribute("a", String)                 // string field "a"
//         Attribute("b", Int32, "operand")       // field with description
//         Attribute("operands", ArrayOf(Int32))  // array field
//         Attribute("ops", MapOf(String, Int32)) // map field
//         Attribute("c", SumMod)                 // field using user type
//         Attribute("len", Int64, func() {       // field with validation
//             Minimum(1)
//         })
//
//         Required("a")                          // Required fields
//         Required("b", "c")
//     })
//
func Type(name string, dsl func()) design.UserType {
	if t := design.Root.UserType(name); t != nil {
		eval.ReportError("type %#v defined twice", name)
		return nil
	}

	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	t := &design.UserTypeExpr{
		TypeName:      name,
		AttributeExpr: &design.AttributeExpr{DSLFunc: dsl},
	}
	if dsl == nil {
		t.Type = design.String
	}
	design.Root.Types = append(design.Root.Types, t)
	return t
}

// ArrayOf creates an array type from its element type.
//
// ArrayOf may be used wherever types can.
// ArrayOf takes one argument: the type of the array elements either by name or by reference.
//
// Example:
//
//    var Bottle = Type("Bottle", func() {
//        Attribute("name")
//    })
//
//    var Account = Type("Account", func() {
//        Attribute("bottles", ArrayOf(Bottle), "Account bottles", func() {
//            MinLength(1)
//        })
//    })
//
func ArrayOf(t design.DataType) *design.Array {
	at := design.AttributeExpr{Type: t}
	return &design.Array{ElemType: &at}
}

// MapOf creates a map from its key and element types.
//
// MapOf may be used wherever types can.
// MapOf takes two arguments: the key and value types either by name of by reference.
//
// Example:
//
//    var Bottle = Type("Bottle", func() {
//        Attribute("name")
//    })
//
//    var Review = Type("Review", func() {
//        Attribute("ratings", MapOf(Bottle, Int32), "Bottle ratings")
//    })
//
func MapOf(k, v design.DataType) *design.Map {
	kat := design.AttributeExpr{Type: k}
	vat := design.AttributeExpr{Type: v}
	return &design.Map{KeyType: &kat, ElemType: &vat}
}
