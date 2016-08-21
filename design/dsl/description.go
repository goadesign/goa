package dsl

import "github.com/goadesign/goa/design"

// Description sets the expression description.
//
// Description may appear in Service, Type or Attribute.
//
// Example:
//
//    Service("adder", func() {
//        Description("Adder service")
//    })
//
func Description(d string) {
	switch expr := eval.Current().(type) {
	case *design.ServiceExpr:
		expr.Description = d
	case *design.AttributeExpr:
		expr.Description = d
	case *design.DocsExpr:
		expr.Description = d
	default:
		eval.IncompatibleDSL()
	}
}
