package dsl

import "goa.design/goa/expr"
import "goa.design/goa/eval"

// Description sets the expression description.
//
// Description must appear in API, Docs, Type or Attribute.
//
// Description accepts one arguments: the description string.
//
// Example:
//
//    API("adder", func() {
//        Description("Adder API")
//    })
//
func Description(d string) {
	switch expr := eval.Current().(type) {
	case *expr.APIExpr:
		expr.Description = d
	case *expr.ServerExpr:
		expr.Description = d
	case *expr.HostExpr:
		expr.Description = d
	case *expr.ServiceExpr:
		expr.Description = d
	case *expr.ResultTypeExpr:
		expr.Description = d
	case *expr.AttributeExpr:
		expr.Description = d
	case *expr.DocsExpr:
		expr.Description = d
	case *expr.MethodExpr:
		expr.Description = d
	case *expr.ExampleExpr:
		expr.Description = d
	case *expr.SchemeExpr:
		expr.Description = d
	default:
		eval.IncompatibleDSL()
	}
}
