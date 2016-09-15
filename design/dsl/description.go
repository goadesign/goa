package dsl

import "github.com/goadesign/goa/design"
import "github.com/goadesign/goa/eval"

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
	case *design.AttributeExpr:
		expr.Description = d
	case *design.DocsExpr:
		expr.Description = d
	default:
		eval.IncompatibleDSL()
	}
}
