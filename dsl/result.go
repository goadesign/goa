package dsl

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Result describes and method result type.
//
// Result may appear in a Method expression.
//
// Result accepts a type as first argument. This argument is optional in which
// case the type must be described inline (see below).
//
// Result accepts an optional DSL function as second argument. This function may
// define the result type inline using Attribute or may further specialize the
// type passed as first argument e.g. by providing additional validations (e.g.
// list of required attributes). The DSL may also specify a view when the first
// argument is a result type corresponding to the view rendered by this method.
// Note that specifying a view when the result type is a result type is optional
// and only useful in cases the method renders a single view.
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
//    Method("add", func() {
//        Result(Int32)
//    })
//
//    // Define result using object defined inline
//    Method("add", func() {
//        Result(func() {
//            Attribute("value", Int32, "Resulting sum")
//            Required("value")
//        })
//    })
//
//    // Define result type using user type
//    Method("add", func() {
//        Result(Sum)
//    })
//
//    // Specify view and required attributes on result type
//    Method("add", func() {
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
	e, ok := eval.Current().(*design.MethodExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	e.Result = methodDSL("Result", val, fns...)
}
