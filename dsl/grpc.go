package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

const (
	// CodeOK represents the gRPC response code "OK".
	CodeOK = 0
	// CodeCanceled represents the gRPC response code "Canceled".
	CodeCanceled = 1
	// CodeUnknown represents the gRPC response code "Unknown".
	CodeUnknown = 2
	// CodeInvalidArgument represents the gRPC response code "InvalidArgument".
	CodeInvalidArgument = 3
	// CodeDeadlineExceeded represents the gRPC response code "DeadlineExceeded".
	CodeDeadlineExceeded = 4
	// CodeNotFound represents the gRPC response code "NotFound".
	CodeNotFound = 5
	// CodeAlreadyExists represents the gRPC response code "AlreadyExists".
	CodeAlreadyExists = 6
	// CodePermissionDenied represents the gRPC response code "PermissionDenied".
	CodePermissionDenied = 7
	// CodeResourceExhausted represents the gRPC response code "ResourceExhausted".
	CodeResourceExhausted = 8
	// CodeFailedPrecondition represents the gRPC response code "FailedPrecondition".
	CodeFailedPrecondition = 9
	// CodeAborted represents the gRPC response code "Aborted".
	CodeAborted = 10
	// CodeOutOfRange represents the gRPC response code "OutOfRange".
	CodeOutOfRange = 11
	// CodeUnimplemented represents the gRPC response code "Unimplemented".
	CodeUnimplemented = 12
	// CodeInternal represents the gRPC response code "Internal".
	CodeInternal = 13
	// CodeUnavailable represents the gRPC response code "Unavailable".
	CodeUnavailable = 14
	// CodeDataLoss represents the gRPC response code "DataLoss".
	CodeDataLoss = 15
	// CodeUnauthenticated represents the gRPC response code "Unauthenticated".
	CodeUnauthenticated = 16
)

// GRPC defines gRPC transport specific properties on an API, a service, or a
// single method. The function maps the request and response types to gRPC
// properties such as request and response messages.
//
// As a special case GRPC may be used to define the response generated for
// invalid requests and internal errors (errors returned by the service methods
// that don't match any of the error responses defined in the design). This is
// the only use of GRPC allowed in the API expression.
//
// The functions that appear in GRPC such as Message or Response may take
// advantage of the request or response types (depending on whether they appear
// when describing the gRPC request or response). The properties of the message
// attributes inherit the properties of the attributes with the same names that
// appear in the request or response types. The functions may also define new
// attributes or override the existing request or response type attributes.
//
// GRPC must appear in a Service or a Method expression.
//
// GRPC accepts a single argument which is the defining DSL function.
//
// Example:
//
//     var CreatePayload = Type("CreatePayload", func() {
//         Field(1, "name", String, "Name of account")
//         TokenField(2, "token", String, "JWT token for authentication")
//     })
//
//     var CreateResult = ResultType("application/vnd.create", func() {
//         Attributes(func() {
//             Field(1, "name", String, "Name of the created resource")
//             Field(2, "href", String, "Href of the created resource")
//         })
//     })
//
//     Method("create", func() {
//         Payload(CreatePayload)
//         Result(CreateResult)
//         Error("unauthenticated")
//
//         GRPC(func() {              // gRPC endpoint to define gRPC service
//             Message(func() {       // gRPC request message
//                 Attribute("token")
//             })
//             Response(CodeOK)     // gRPC success response
//             Response("unauthenticated", CodeUnauthenticated) // gRPC error
//         })
//     })
//
func GRPC(fn func()) {
	switch actual := eval.Current().(type) {
	case *expr.ServiceExpr:
		res := expr.Root.API.GRPC.ServiceFor(actual)
		res.DSLFunc = fn
	case *expr.MethodExpr:
		res := expr.Root.API.GRPC.ServiceFor(actual.Service)
		act := res.EndpointFor(actual.Name, actual)
		act.DSLFunc = fn
	default:
		eval.IncompatibleDSL()
	}
}

// Message describes a gRPC request or response message.
//
// Message must appear in a gRPC method expression to define the
// attributes that must appear in a request message or in a gRPC response
// expression to define the attributes that must appear in a response message.
// If Message is absent then the request message is built using the method
// payload expression and the response message is built using the method
// result expression.
//
// Message accepts one argument of function type which lists the attributes
// that must be present in the message. For example, the Message function can
// be defined on the gRPC method expression listing the security attributes
// to appear in the request message instead of sending them in the gRPC
// metadata by default. The attributes listed in the function inherit the
// properties (description, type, meta, validations etc.) of the request or
// response type attributes with identical names.
//
// Example:
//
//     var CreatePayload = Type("CreatePayload", func() {
//         Field(1, "name", String, "Name of account")
//         TokenField(2, "token", String, "JWT token for authentication")
//     })
//
//     var CreateResult = ResultType("application/vnd.create", func() {
//         Attributes(func() {
//             Field(1, "name", String, "Name of the created resource")
//             Field(2, "href", String, "Href of the created resource")
//         })
//     })
//
//     Method("create", func() {
//         Payload(CreatePayload)
//         Result(CreateResult)
//         GRPC(func() {
//             Message(func() {
//                 Attribute("token") // "token" sent in the request message
//                                    // along with "name"
//             })
//             Response(func() {
//                 Code(CodeOK)
//                 Message(func() {
//                     Attribute("name") // "name" sent in the response
//                                       // message along with "href"
//                     Required("name")  // "name" is set to required
//                 })
//             })
//         })
//     })
//
// If the method payload/result type is a primitive, array, or a map the
// request/response message by default contains one attribute with name
// "field", "rpc:tag" set to 1, and the type set to the type of the
// method payload/result. The function argument can also be used to set
// the message field name to something other than "field".
//
// Example:
//
//     Method("add", func() {
//         Payload(Operands)
//         Result(Int)      // method Result is a primitive
//         GRPC(func() {
//             Response(CodeOK, func()
//                 Message(func() {
//                     Attribute("sum") // Response message has one field with
//                                      // name "sum" instead of the default
//                                      // "field"
//                 })
//             })
//         })
//     })
//
func Message(fn func()) {
	var setter func(*expr.AttributeExpr)
	{
		switch e := eval.Current().(type) {
		case *expr.GRPCEndpointExpr:
			setter = func(att *expr.AttributeExpr) {
				e.Request = att
			}
		case *expr.GRPCErrorExpr:
			setter = func(att *expr.AttributeExpr) {
				if e.Response == nil {
					e.Response = &expr.GRPCResponseExpr{}
				}
				e.Response.Message = att
			}
		case *expr.GRPCResponseExpr:
			setter = func(att *expr.AttributeExpr) {
				e.Message = att
			}
		default:
			eval.IncompatibleDSL()
			return
		}
	}
	attr := &expr.AttributeExpr{}
	if eval.Execute(fn, attr) {
		setter(attr)
	}
}

// Metadata defines a gRPC request metadata.
//
// Metadata must appear in a gRPC endpoint expression to describe gRPC request
// metadata.
//
// Security attributes in the method payload are automatically added to the
// request metadata unless specified explicitly in request message using
// Message function. All other attributes in method payload are added to the
// request message unless specified explicitly using Metadata (in which case
// will be added to the metadata).
//
// Metadata takes one argument of function type which lists the attributes
// that must be set in the request metadata instead of the message.
// If Metadata is set in the gRPC endpoint expression, it inherits the
// attribute properties (description, type, meta, validations etc.) from the
// method payload.
//
// Example:
//
//     var CreatePayload = Type("CreatePayload", func() {
//         Field(1, "name", String, "Name of account")
//         TokenField(2, "token", String, "JWT token for authentication")
//     })
//
//     var CreateResult = ResultType("application/vnd.create", func() {
//         Attributes(func() {
//             Field(1, "name", String, "Name of the created resource")
//             Field(2, "href", String, "Href of the created resource")
//         })
//     })
//
//     Method("create", func() {
//         Payload(CreatePayload)
//         Result(CreateResult)
//         GRPC(func() {
//             Metadata(func() {
//                 Attribute("name") // "name" sent in the request metadata
//                                   // along with "token"
//             })
//             Response(func() {
//                 Code(CodeOK)
//             })
//         })
//     })
//
func Metadata(fn func()) {
	switch e := eval.Current().(type) {
	case *expr.GRPCEndpointExpr:
		attr := &expr.AttributeExpr{}
		if eval.Execute(fn, attr) {
			e.Metadata = expr.NewMappedAttributeExpr(attr)
		}
	default:
		eval.IncompatibleDSL()
	}
}

// Trailers defines gRPC trailers in response metadata.
//
// Trailers must appear in a gRPC response expression to describe gRPC trailers
// in response metadata.
//
// Trailers takes one argument of function type which lists the attributes
// that must be set in the trailer response metadata instead of the message.
// If Trailers is set in the gRPC response expression, it inherits the
// attribute properties (description, type, meta, validations etc.) from the
// method result.
//
// Example:
//
//     var CreatePayload = Type("CreatePayload", func() {
//         Field(1, "name", String, "Name of account")
//         TokenField(2, "token", String, "JWT token for authentication")
//     })
//
//     var CreateResult = ResultType("application/vnd.create", func() {
//         Attributes(func() {
//             Field(1, "name", String, "Name of the created resource")
//             Field(2, "href", String, "Href of the created resource")
//         })
//     })
//
//     Method("create", func() {
//         Payload(CreatePayload)
//         Result(CreateResult)
//         GRPC(func() {
//             Response(func() {
//                 Code(CodeOK)
//                 Trailers(func() {
//                     Attribute("name") // "name" sent in the trailer metadata
//                 })
//             })
//         })
//     })
//
func Trailers(fn func()) {
	switch e := eval.Current().(type) {
	case *expr.GRPCResponseExpr:
		attr := &expr.AttributeExpr{}
		if eval.Execute(fn, attr) {
			e.Trailers = expr.NewMappedAttributeExpr(attr)
		}
	default:
		eval.IncompatibleDSL()
	}
}
