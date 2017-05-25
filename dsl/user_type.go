package dsl

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Type defines a user type. A user type has a unique name and may be an alias
// to an existing type or may describe a completely new type using a list of
// attributes (object fields). Attribute types may themselves be user type.
// When a user type is defined as an alias to another type it may define
// additional validations - for example it a user type which is an alias of
// String may define a validation pattern that all instances of the type
// must match.
//
// Type is a top level definition.
//
// Type takes two or three arguments: the first argument is the name of the type.
// The name must be unique. The second argument is either another type or a
// function. If the second argument is a type then there may be a function passed
// as third argument.
//
// Example:
//
//     // simple alias
//     var MyString = Type("MyString", String)
//
//     // alias with description and additional validation
//     var Hostname = Type("Hostname", String, func() {
//         Description("A host name")
//         Format(FormatHostname)
//     })
//
//     // new type
//     var SumPayload = Type("SumPayload", func() {
//         Description("Type sent to add endpoint")
//
//         Attribute("a", String)                 // string attribute "a"
//         Attribute("b", Int32, "operand")       // attribute with description
//         Attribute("operands", ArrayOf(Int32))  // array attribute
//         Attribute("ops", MapOf(String, Int32)) // map attribute
//         Attribute("c", SumMod)                 // attribute using user type
//         Attribute("len", Int64, func() {       // attribute with validation
//             Minimum(1)
//         })
//
//         Required("a")                          // Required attributes
//         Required("b", "c")
//     })
//
func Type(name string, args ...interface{}) design.UserType {
	if len(args) > 2 {
		eval.ReportError("too many arguments")
		return nil
	}
	if t := design.Root.UserType(name); t != nil {
		eval.ReportError("type %#v defined twice", name)
		return nil
	}

	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}

	var (
		base design.DataType
		fn   func()
	)
	if len(args) == 0 {
		// Make Type behave like Attribute
		args = []interface{}{design.String}
	}
	switch a := args[0].(type) {
	case design.DataType:
		base = a
		if len(args) == 2 {
			d, ok := args[1].(func())
			if !ok {
				eval.ReportError("third argument must be a function")
				return nil
			}
			fn = d
		}
	case func():
		base = design.Object{}
		fn = a
		if len(args) == 2 {
			eval.ReportError("only one argument allowed when it is a function")
			return nil
		}
	default:
		eval.InvalidArgError("type or function", args[0])
		return nil
	}

	t := &design.UserTypeExpr{
		TypeName:      name,
		AttributeExpr: &design.AttributeExpr{Type: base, DSLFunc: fn},
	}
	design.Root.Types = append(design.Root.Types, t)
	return t
}

// ArrayOf creates an array type from its element type.
//
// ArrayOf may be used wherever types can.
// The first argument of ArrayOf is the type of the array elements specified by
// name or by reference.
// The second argument of ArrayOf is an optional function that defines
// validations for the array elements.
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
func ArrayOf(v interface{}, fn ...func()) *design.Array {
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
	if len(fn) > 1 {
		eval.ReportError("ArrayOf: too many arguments")
		return res
	}
	at := design.AttributeExpr{Type: t}
	if len(fn) == 1 {
		eval.Execute(fn[0], &at)
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
func MapOf(k, v interface{}, fn ...func()) *design.Map {
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
	if len(fn) > 1 {
		eval.ReportError("MapOf: too many arguments")
		return res
	}
	kat := design.AttributeExpr{Type: tk}
	vat := design.AttributeExpr{Type: tv}
	m := &design.Map{KeyType: &kat, ElemType: &vat}
	if len(fn) == 1 {
		mat := design.AttributeExpr{Type: m}
		eval.Execute(fn[0], &mat)
	}
	return m
}

// Key makes it possible to specify validations for map keys.
func Key(fn func()) {
	at, ok := eval.Current().(*design.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if m, ok := at.Type.(*design.Map); ok {
		eval.Execute(fn, m.KeyType)
		return
	}
	eval.IncompatibleDSL()
}

// Elem makes it possible to specify validations for array and map values.
func Elem(fn func()) {
	at, ok := eval.Current().(*design.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	switch e := at.Type.(type) {
	case *design.Array:
		eval.Execute(fn, e.ElemType)
	case *design.Map:
		eval.Execute(fn, e.ElemType)
	default:
		eval.IncompatibleDSL()
	}
}
