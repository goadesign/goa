package dsl

import (
	"github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/rpc/design"
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
