package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Service defines a group of remotely accessible methods that are hosted
// together. The service DSL makes it possible to define the methods, their
// input and output as well as the errors they may return independently of the
// underlying transport (HTTP or gRPC). The transport specific DSLs defined by
// the HTTP and GRPC functions define the mapping between the input, output and
// error type attributes and the transport data (e.g. HTTP headers, HTTP bodies
// or gRPC messages).
//
// The Service expression is leveraged by the code generators to define the
// business layer service interface, the endpoint layer as well as the transport
// layer including input validation, marshalling and unmarshalling. It also
// affects the generated OpenAPI specification.
//
// Service is as a top level expression.
//
// Service accepts two arguments: the name of the service - which must be unique
// in the design package - and its defining DSL.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Title("divider service") // optional
//
//        Error("Unauthorized") // error that apply to all the service methods
//        HTTP(func() {         // HTTP mapping for error responses
//            // Use HTTP status 401 for 'Unauthorized' errors.
//            Response("Unauthorized", StatusUnauthorized)
//        })
//
//        Method("divide", func() {   // Defines a service method.
//            Description("Divide divides two value.") // optional
//            Payload(DividePayload)                   // input type
//            Result(Float64)                          // output type
//            Error("DivisionByZero")                  // method specific error
//            // No HTTP mapping for "DivisionByZero" means default of status
//            // 400 and error struct serialized in HTTP response body.
//
//            HTTP(func() {      // Defines HTTP transport mapping.
//                GET("/div")    // HTTP verb and path
//                Param("a")     // query string parameter
//                Param("b")     // 'a' and 'b' are attributes of DividePayload.
//                // No 'Response' DSL means default of status 200 and result
//                // marshaled in HTTP response body.
//            })
//        })
//    })
//
func Service(name string, fn func()) *expr.ServiceExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}
	if s := expr.Root.Service(name); s != nil {
		eval.ReportError("service %#v is defined twice", name)
		return nil
	}
	s := &expr.ServiceExpr{Name: name, DSLFunc: fn}
	expr.Root.Services = append(expr.Root.Services, s)
	return s
}
