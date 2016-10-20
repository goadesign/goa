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
//    var _ = Service("divider", func() {
//        Description("divider service") // Optional description
//
//        DefaultMedia(DivideResponseMedia) // ??
//        Error("Unauthorized", Unauthorized) // Applies to all endpoints
//
//        HTTP(func() {
//                BasePath("/divide")
//                Parent("math")
//                CanonicalActionName("get")
//        })
//
//        Endpoint("divide", func() {     // Defines a single endpoint
//            Description("The divide endpoint returns the division of A and B")
//            Request(DivideRequest)      // Optional, GRPC generation uses built-in empty type if absent
//            Response(DivideResponse)    // Ditto
//            Error("DivisionByZero", ErrDivByZero) // ErrDivByZero is optional type that describes error body.
//               If gRPC error attribute is added to type, if return error matches design error then
//               error attribute is set otherwise error is returned to gRPC server.
//
//            HTTP(func() {
//                Scheme("https")
//                GET("/{ID:ParentID}/{Divisor}") // DivideRequest must have Dividend and Divisor attributes
//                POST("/{Dividend}")         // Body is DivideRequest minus Dividend attribute and headers
//                POST("/")                   // Body is DivideRequest minus headers
//                Param("{Foo:Bar}")
//                Header("Account")           // Must match one of DivideRequest attributes
//                Payload("Payload")
//                Payload(func() {
//                    Field("bar")
//                })
//                Response(func() {
//                    Status(OK)              // Default
//                    Header("Result")        // Must be an attribute of DivideResponse
//                })
//                Error("DivisionByZero", func() {
//                    Status(BadRequest)      // Default
//                    Header("Message")       // Must be an attribute of ErrDivByZero
//                })
//            })
//
//            GRPC(func() {
//                // STREAMING?
//                Proto("divider.divide") // rpc definition in proto file
//                Error("DivisionByZero", func() { // Defines which field contains error if not "Error"
//                    Field("DivByZero")
//                })
//            })
//        })
//    })
//
func Service(name string, dsl func()) *design.ServiceExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}
	if s := design.Root.Service(name); s != nil {
		eval.ReportError("service %#v is defined twice", name)
		return nil
	}
	s := &design.ServiceExpr{Name: name, DSLFunc: dsl}
	design.Root.Services = append(design.Root.Services, s)
	return s
}
