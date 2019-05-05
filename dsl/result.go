package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Result defines the data type of a method output.
//
// Result must appear in a Method expression.
//
// Result takes one to three arguments. The first argument is either a type or a
// DSL function. If the first argument is a type then an optional description
// may be passed as second argument. Finally a DSL may be passed as last
// argument that further specializes the type by providing additional
// validations (e.g. list of required attributes) The DSL may also specify a
// view when the first argument is a result type corresponding to the view
// rendered by this method. If no view is specified then the generated code
// defines response methods for all views.
//
// The valid syntax for Result is thus:
//
//    Result(Type)
//
//    Result(func())
//
//    Result(Type, "description")
//
//    Result(Type, func())
//
//    Result(Type, "description", func())
//
// Examples:
//
//    // Define result using primitive type
//    Method("add", func() {
//        Result(Int32)
//    })
//
//    // Define result using primitive type and description
//    Method("add", func() {
//        Result(Int32, "Resulting sum")
//    })
//
//    // Define result using primitive type, description and validations.
//    Method("add", func() {
//        Result(Int32, "Resulting sum", func() {
//            Minimum(0)
//        })
//    })
//
//    // Define result using object defined inline
//    Method("add", func() {
//        Result(func() {
//            Description("Result defines a single field which is the sum.")
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
func Result(val interface{}, args ...interface{}) {
	if len(args) > 2 {
		eval.ReportError("too many arguments")
		return
	}
	e, ok := eval.Current().(*expr.MethodExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	e.Result = methodDSL("Result", val, args...)
}

// StreamingResult defines a method that streams instances of the given type.
//
// StreamingResult must appear in a Method expression.
//
// The arguments to a StreamingResult DSL is same as the Result DSL.
//
// Examples:
//
//    // Method result is a stream of integers
//    Method("add", func() {
//        StreamingResult(Int32)
//    })
//
//    Method("add", func() {
//        StreamingResult(Int32, "Resulting sum")
//    })
//
//    // Method result is a stream of integers with validation set on each
//    Method("add", func() {
//        StreamingResult(Int32, "Resulting sum", func() {
//            Minimum(0)
//        })
//    })
//
//    // Method result is a stream of objects defined inline
//    Method("add", func() {
//        StreamingResult(func() {
//            Description("Result defines a single field which is the sum.")
//            Attribute("value", Int32, "Resulting sum")
//            Required("value")
//        })
//    })
//
//    // Method result is a stream of user type
//    Method("add", func() {
//        StreamingResult(Sum)
//    })
//
//    // Method result is a stream of result type with a view
//    Method("add", func() {
//        StreamingResult(Sum, func() {
//            View("default")
//            Required("value")
//        })
//    })
//
func StreamingResult(val interface{}, args ...interface{}) {
	if len(args) > 2 {
		eval.ReportError("too many arguments")
		return
	}
	e, ok := eval.Current().(*expr.MethodExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	e.Result = methodDSL("Result", val, args...)
	if e.Stream == expr.ClientStreamKind {
		e.Stream = expr.BidirectionalStreamKind
	} else {
		e.Stream = expr.ServerStreamKind
	}
}
