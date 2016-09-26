package dsl

import (
	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/rpc/design"
)

// Service defines a group of related endpoints that are exposed from the same process.
//
// Service is as a top level expression.
//
// Example:
//
//    var _ = Service("adder", func() {
//        Description("adder service") // Optional description
//
//        Endpoint("add", func() {     // Defines a single endpoint
//            Description("The add endpoint returns the sum of A and B")
//            Request(AddPayload)
//            Response(Add)
//        })
//    })
//
func Service(name string, adsl ...func()) *design.ServiceExpr {
	s := &design.ServiceExpr{Name: name}
	if len(adsl) > 1 {
		eval.ReportError("too many arguments in call to Service")
		return nil
	}
	if len(adsl) == 1 {
		s.DSLFunc = adsl[0]
	}
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}
	design.Root.Services = append(design.Root.Services, s)
	return s
}
