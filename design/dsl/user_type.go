package dsl

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
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
// The first argument of ArrayOf is the type of the array elements specified by
// name or by reference.
// The second argument of ArrayOf is an optional DSL that defines validations
// for the array elements.
//
// Examples:
//
//    var Names = ArrayOf(String, func() {
//        Pattern("[a-zA-Z]+") // Validates elements of the array
//    })
//
//    var Account = Type("Account", func() {
//        Attribute("bottles", ArrayOf(Bottle), "Account bottles", func() {
//            MinLength(1) // Validates array as a whole
//        })
//    })
//
// Note: CollectionOf and ArrayOf both return array types. CollectionOf returns
// a media type where ArrayOf returns a user type. In general you want to use
// CollectionOf if the argument is a media type and ArrayOf if it is a user
// type.
func ArrayOf(v interface{}, dsl ...func()) *design.Array {
	var t design.DataType
	var ok bool
	t, ok = v.(design.DataType)
	if !ok {
		if name, ok := v.(string); ok {
			t = design.Root.UserType(name)
		}
	}
	// never return nil to avoid panics, errors are reported after DSL execution
	res := &design.Array{ElemType: &design.AttributeExpr{Type: design.String}}
	if t == nil {
		eval.ReportError("invalid ArrayOf argument: not a type and not a known user type name")
		return res
	}
	if len(dsl) > 1 {
		eval.ReportError("ArrayOf: too many arguments")
		return res
	}
	at := design.AttributeExpr{Type: t}
	if len(dsl) == 1 {
		eval.Execute(dsl[0], &at)
	}
	return &design.Array{ElemType: &at}
}

// MapOf creates a map from its key and element types.
//
// MapOf may be used wherever types can.
// MapOf takes two arguments: the key and value types either by name of by reference.
//
// Example:
//
//    var ReviewByID = MapOf(Int64, String, func() {
//        Key(func() {
//            Minimum(1)           // Validates keys of the map
//        })
//        Value(func() {
//            Pattern("[a-zA-Z]+") // Validates values of the map
//        })
//    })
//
//    var Review = Type("Review", func() {
//        Attribute("ratings", MapOf(Bottle, Int32), "Bottle ratings")
//    })
//
func MapOf(k, v interface{}, dsl ...func()) *design.Map {
	var tk, tv design.DataType
	var ok bool
	tk, ok = k.(design.DataType)
	if !ok {
		if name, ok := k.(string); ok {
			tk = design.Root.UserType(name)
		}
	}
	tv, ok = v.(design.DataType)
	if !ok {
		if name, ok := v.(string); ok {
			tv = design.Root.UserType(name)
		}
	}
	// never return nil to avoid panics, errors are reported after DSL execution
	res := &design.Map{KeyType: &design.AttributeExpr{Type: design.String}, ElemType: &design.AttributeExpr{Type: design.String}}
	if tk == nil {
		eval.ReportError("invalid MapOf key argument: not a type and not a known user type name")
		return res
	}
	if tv == nil {
		eval.ReportError("invalid MapOf value argument: not a type and not a known user type name")
		return res
	}
	if len(dsl) > 1 {
		eval.ReportError("MapOf: too many arguments")
		return res
	}
	kat := design.AttributeExpr{Type: tk}
	vat := design.AttributeExpr{Type: tv}
	m := &design.Map{KeyType: &kat, ElemType: &vat}
	if len(dsl) == 1 {
		mat := design.AttributeExpr{Type: m}
		eval.Execute(dsl[0], &mat)
	}
	return m
}

// Key makes it possible to specify validations for map keys.
func Key(dsl func()) {
	at, ok := eval.Current().(*design.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	m := at.Type.(*design.Map)
	eval.Execute(dsl, m.KeyType)
}

// Value makes it possible to specify validations for map values.
func Value(dsl func()) {
	at, ok := eval.Current().(*design.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	m := at.Type.(*design.Map)
	eval.Execute(dsl, m.ElemType)
}
