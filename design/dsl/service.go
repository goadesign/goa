package dsl

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Service defines a group of related endpoints. Refer to the transport specific
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
//                                  // endpoints. Also defines default
//                                  // properties (type, description and
//                                  // validations) for attributes with
//                                  // identical names in request types.
//
//        Error("Unauthorized", Unauthorized) // Error response that applies to
//                                            // all endpoints
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

// DefaultType sets the service default response type by name or by reference.
// The attributes of the service default type also define the default properties
// for request type attributes with identical names.
//
// DefaultType may appear in Service expressions.
// DefaultType accepts one argument: the name of a reference to the type.
// Example:
//
//    var _ = Service("divider", func() {
//        DefaultType(DivideResult)
//
//        // Endpoint which uses the default type for its response.
//        Endpoint("divide", func() {
//            Request(DivideRequest)
//        })
//    })
//
func DefaultType(val interface{}) {
	if s, ok := eval.Current().(*design.ServiceExpr); ok {
		switch actual := val.(type) {
		case *design.UserTypeExpr:
			s.DefaultTypeName = actual.Name()
		case *design.MediaTypeExpr:
			s.DefaultTypeName = actual.Name()
		case string:
			s.DefaultTypeName = actual
		default:
			eval.ReportError("default type must be a string or a reference to a type")
			return
		}
	}
}

// Error describes an endpoint error response. The description includes a unique
// name (in the scope of the endpoint), an optional type, description and DSL
// that further describes the type. If no type is specified then the goa
// ErrorMedia type is used. The DSL syntax is identical to the Attribute DSL.
// Transport specific DSL may further describe the mapping between the error
// type attributes and the serialized response.
//
// Error may appear in the Service (to define error responses that apply to all
// the service endpoints) or Endppoint expressions.
// See Attribute for details on the Error arguments.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Error("invalid_arguments") // Uses type ErrorMedia
//
//        // Endpoint which uses the default type for its response.
//        Endpoint("divide", func() {
//            Request(DivideRequest)
//            Error("div_by_zero", DivByZero, "Division by zero")
//        })
//    })
//
func Error(name string, args ...interface{}) {
	if len(args) == 0 {
		args = []interface{}{design.ErrorMedia}
	}
	dt, desc, dsl := parseAttributeArgs(nil, args...)
	att := &design.AttributeExpr{
		Description: desc,
		Type:        dt,
		DSLFunc:     dsl,
	}
	if dsl != nil {
		eval.Execute(dsl, att)
	}
	erro := &design.ErrorExpr{AttributeExpr: att, Name: name}
	switch actual := eval.Current().(type) {
	case *design.ServiceExpr:
		actual.Errors = append(actual.Errors, erro)
	case *design.EndpointExpr:
		actual.Errors = append(actual.Errors, erro)
	default:
		eval.IncompatibleDSL()
	}
}
