package dsl

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Result describes and endpoint result type.
//
// Result may appear in a Endpoint expression.
//
// Result accepts a type as first argument. This argument is optional in which
// case the type must be described inline (see below).
//
// Result accepts an optional DSL function as second argument. This function may
// define the result type inline using Attribute or may further specialize the
// type passed as first argument e.g. by providing additional validations (e.g.
// list of required attributes). The DSL may also specify a view when the first
// argument is a media type corresponding to the view rendered by this endpoint.
// Note that specifying a view when the result type is a media type is optional
// and only useful in cases the endpoint renders a single view.
//
// The valid syntax for Result is thus:
//
//    Result(dsltype)
//
//    Result(func())
//
//    Result(dsltype, func())
//
// Examples:
//
//    // Define result using primitive type
//    Endpoint("add", func() {
//        Result(Int32)
//    })
//
//    // Define result using object defined inline
//    Endpoint("add", func() {
//        Result(func() {
//            Attribute("value", Int32, "Resulting sum")
//            Required("value")
//        })
//    })
//
//    // Define result type using user type
//    Endpoint("add", func() {
//        Result(Sum)
//    })
//
//    // Specify view and required attributes on media type
//    Endpoint("add", func() {
//        Result(Sum, func() {
//            View("default")
//            Required("value")
//        })
//    })
//
func Result(val interface{}, fns ...func()) {
	if len(fns) > 1 {
		eval.ReportError("too many arguments")
		return
	}
	e, ok := eval.Current().(*design.EndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	e.Result = endpointDSL("Result", val, fns...)
}
