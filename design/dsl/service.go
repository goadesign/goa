package dsl

import (
	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/rpc/design"
)

// Service defines a group of related endpoints.
//
// Service is as a top level expression.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Description("divider service") // Optional description
//
//        DefaultType(DivideResult) // Default response type for the service
//                                  // endpoints. Also defines default
//                                  // properties (type, description and
//                                  // validations) for attributes with
//                                  // identical names in request types.
//
//        Error("Unauthorized", Unauthorized) // Error response that applies to
//                                            // all endpoints
//
//        HTTP(func() {                  // HTTP specific expressions
//                BasePath("/divide")    // Common path prefix to all endpoints
//                Parent("math")         // Parent resource
//                CanonicalActionName("get") // Action whose first route defines
//                                           // the path prefix to all child
//                                           // resources.
//        })
//
//        GRPC(func() {            // gRPC specific expressions
//                Package("math")  // Name of protobuf package
//                Name("Divider")  // Name of protobuf service
//        })
//
//        Endpoint("divide", func() {     // Defines a single endpoint
//            Description("The divide endpoint returns the division of A and B")
//            Request(DivideRequest)      // Request type listing all request
//                                        // parameters in its attributes.
//            Response(DivideResponse)    // Response type.
//            Error("DivisionByZero", DivByZero) // Error, has a name and
//                                               // optionally a type
//                                               // (DivByZero) describes the
//                                               // error response.
//
//            HTTP(func() {               // HTTP specific expressions
//                GET("/{Dividend}/{Divisor}") // Use request type attributes
//                                        // "Dividend" and "Divisor" to define
//                                        // path parameters.
//                Response(OK)
//            })
//
//            GRPC(func() {
//                Name("Divide")
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
