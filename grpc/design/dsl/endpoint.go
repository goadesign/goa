package dsl

import (
	"github.com/goadesign/grpc/design"
	"github.com/goadesign/grpc/eval"
)

// Endpoint defines a single service RPC endpoint.
//
// Endpoint may appear in a EndpointGroup expression.
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
	ep := &design.EndpointExpr{Name: name, DSLFunc: dsl}
	switch current := eval.Current().(type) {
	case *ServiceExpr:
		current.Endpoints = append(current.Endpoints, ep)
	case *EndpointGroupExpr:
		current.Endpoints = append(current.Endpoints, ep)
	default:
		eval.IncompatibleDSL()
	}
}

// EndpointGroup defines a group of endpoints that share common properties and are implemented
// together.
//
// EndpointGroup is as a top level expression.
//
// Example:
//
//    var _ = EndpointGroup("operands", func() {
//        Service("adder")                          // Identifies the service(s) which expose
//                                                  // the endpoints in this group.
//        Description("operands related endpoints") // Optional description
//
//        Endpoint("add", func() {                  // Defines a single endpoint
//            Description("The add endpoint returns the sum of A and B")
//            Request(AddPayload)
//            Response(Add)
//        })
//    })
//
func EndpointGroup(name string, dsl ...func()) *EndpointGroupExpr {
	epg := &design.EndpointGroupExpr{Name: name, DSLFunc: dsl}
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}
	Root.EndpointGroups = append(Root.EndpointGroups, epg)
	return epg
}
