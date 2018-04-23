package dsl

import "goa.design/goa/design"
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
	case *design.APIExpr:
		expr.Description = d
	case *design.ServerExpr:
		expr.Description = d
	case *design.ServiceExpr:
		expr.Description = d
	case *design.ResultTypeExpr:
		expr.Description = d
	case *design.AttributeExpr:
		expr.Description = d
	case *design.DocsExpr:
		expr.Description = d
	case *design.MethodExpr:
		expr.Description = d
	case *design.ExampleExpr:
		expr.Description = d
	case *design.SchemeExpr:
		expr.Description = d
	default:
		eval.IncompatibleDSL()
	}
}
