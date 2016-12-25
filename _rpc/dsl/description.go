package dsl

import (
	"goa.design/goa.v2/dsl"
	"goa.design/goa.v2/eval"
	"goa.design/goa.v2/rpc/design"
)

// Description sets the expression description.
//
// Description may appear in API, Service, Endpoint, Type or Field.
//
// Example:
//
//    var _ = Service("adder", func() {
//        Description("The adder service")
//    })
//
func Description(d string) {
	switch expr := eval.Current().(type) {
	case *design.ServiceExpr:
		expr.Description = d
	case *design.EndpointExpr:
		expr.Description = d
	default:
		dsl.Description(d)
	}
}
