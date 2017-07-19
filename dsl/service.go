package dsl

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Service defines a group of related methods. Refer to the transport specific
// DSLs to learn how to provide transport specific information.
//
// Service is as a top level expression.
// Service accepts two arguments: the name of the service (which must be unique
// in the design package) and its defining DSL.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Description("divider service") // Optional description
//
//        DefaultType(DivideResult) // Default response type for the service
//                                  // methods. Also defines default
//                                  // properties (type, description and
//                                  // validations) for attributes with
//                                  // identical names in request types.
//
//        Error("Unauthorized", Unauthorized) // Error response that applies to
//                                            // all methods
//
//        Method("divide", func() {     // Defines a single method
//            Description("The divide method returns the division of A and B")
//            Request(DivideRequest)    // Request type listing all request
//                                      // parameters in its attributes.
//            Response(DivideResponse)  // Response type.
//            Error("DivisionByZero", DivByZero) // Error, has a name and
//                                               // optionally a type
//                                               // (DivByZero) describes the
//                                               // error response.
//        })
//    })
//
func Service(name string, fn func()) *design.ServiceExpr {
	if _, ok := eval.Current().(eval.TopExpr); !ok {
		eval.IncompatibleDSL()
		return nil
	}
	if s := design.Root.Service(name); s != nil {
		eval.ReportError("service %#v is defined twice", name)
		return nil
	}
	s := &design.ServiceExpr{Name: name, DSLFunc: fn}
	design.Root.Services = append(design.Root.Services, s)
	return s
}
