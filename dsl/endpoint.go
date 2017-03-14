package dsl

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
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
//        Payload(Operands)
//        Result(Sum)
//        Error(ErrInvalidOperands)
//    })
//
func Endpoint(name string, dsl func()) {
	s, ok := eval.Current().(*design.ServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	ep := &design.EndpointExpr{Name: name, Service: s, DSLFunc: dsl}
	if eval.Execute(ep.DSL(), ep) {
		s.Endpoints = append(s.Endpoints, ep)
	}
}
