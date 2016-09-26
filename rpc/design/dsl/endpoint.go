package dsl

import (
	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/rpc/design"
)

// Endpoint defines a single service RPC endpoint.
//
// Endpoint may appear in a Service expression.
// Endpoint takes two arguments: the name of the endpoint and the defining DSL.
//
// Example:
//
//    Endpoint("add", func() {
//        Description("The add endpoint returns the sum of A and B")
//        Request(AddPayload)
//        Response(Add)
//        Metadata("option:google.api.http", `{post: "/v1/example/echo"}`)
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
