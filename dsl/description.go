package dsl

import "goa.design/goa/v3/expr"
import "goa.design/goa/v3/eval"

// Description sets the expression description.
//
// Description may appear in API, Docs, Type or Attribute.
// Description may also appear in Response and FileServer.
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
	switch e := eval.Current().(type) {
	case *expr.APIExpr:
		e.Description = d
	case *expr.ServerExpr:
		e.Description = d
	case *expr.HostExpr:
		e.Description = d
	case *expr.ServiceExpr:
		e.Description = d
	case *expr.ResultTypeExpr:
		e.Description = d
	case *expr.AttributeExpr:
		e.Description = d
	case *expr.DocsExpr:
		e.Description = d
	case *expr.MethodExpr:
		e.Description = d
	case *expr.ExampleExpr:
		e.Description = d
	case *expr.SchemeExpr:
		e.Description = d
	case *expr.HTTPResponseExpr:
		e.Description = d
	case *expr.HTTPFileServerExpr:
		e.Description = d
	case *expr.GRPCResponseExpr:
		e.Description = d
	default:
		eval.IncompatibleDSL()
	}
}
