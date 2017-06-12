package dsl

import "goa.design/goa.v2/design"
import "goa.design/goa.v2/eval"

// Description sets the expression description.
//
// Description may appear in API, Docs, Type or Attribute.
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
	case *design.MediaTypeExpr:
		expr.Description = d
	case *design.AttributeExpr:
		expr.Description = d
	case *design.DocsExpr:
		expr.Description = d
	case *design.MethodExpr:
		expr.Description = d
	case *design.ExampleExpr:
		expr.Description = d
	default:
		eval.IncompatibleDSL()
	}
}
