package dsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/eval"
)

// Endpoint defines a single service endpoint.
//
// Endpoint may appear in a Service expression.
// Endpoint takes two arguments: the name of the endpoint and the defining DSL.
//
// Example:
//
//    Endpoint("add", func() {
//        Description("The add endpoint returns the sum of A and B")
//        Docs(func() {
//            Description("Add docs")
//            URL("http//adder.goa.design/docs/actions/add")
//        })
//        Request(Operands)
//        Response(Sum)
//        Error(ErrInvalidOperands)
//    })
//
func Endpoint(name string, dsl func()) {
	s, ok := eval.Current().(*design.ServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	ep := &design.EndpointExpr{Name: name, DSLFunc: dsl}
	s.Endpoints = append(s.Endpoints, ep)
}
