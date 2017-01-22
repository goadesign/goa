package dsl

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Result describes and endpoint result type.
//
// Result may appear in a Endpoint expression.
//
// Result takes one or two arguments. The first argument is either a type or a
// DSL function. If the first argument is a type then an optional DSL may be
// passed as second argument that further specializes the type by providing
// additional validations (e.g. list of required attributes)
//
// Examples:
//
// Endpoint("add", func() {
//     // Define result using primitive type
//     Result(Int32)
// })
//
// Endpoint("add", func() {
//     // Define result using object defined inline
//     Result(func() {
//         Attribute("value", Int32, "Resulting sum")
//         Required("value")
//     })
// })
//
// Endpoint("add", func() {
//     // Define result type using user type
//     Result(Sum) // this works too: Result("Sum")
// })
//
// Endpoint("add", func() {
//     // Specify required attributes on user type
//     Result(Sum, func() {
//         Required("value")
//     })
// })
//
func Result(val interface{}, dsls ...func()) {
	if len(dsls) > 1 {
		eval.ReportError("too many arguments")
		return
	}
	e, ok := eval.Current().(*design.EndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	e.Result = endpointTypeDSL("Result", val, dsls...)
}
