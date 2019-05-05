package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Headers describes HTTP request/response or gRPC response headers.
// When used in a HTTP expression, it groups a set of Header expressions and
// makes it possible to list required headers using the Required function.
// When used in a GRPC response expression, it defines the headers to be sent
// in the response metadata.
//
// To define HTTP headers, Headers must appear in an Service HTTP expression
// to define request headers common to all the service methods. Headers may
// also appear in a method, response or error HTTP expression to define the
// HTTP endpoint request and response headers.
//
// To define gRPC response header metadata, Headers must appear in a GRPC
// response expression.
//
// Headers accepts one argument which is a function listing the headers.
//
// Example:
//
//     // HTTP headers
//
//     var _ = Service("cellar", func() {
//         HTTP(func() {
//             Headers(func() {
//                 Header("version:Api-Version", String, "API version", func() {
//                     Enum("1.0", "2.0")
//                 })
//                 Required("version")
//             })
//         })
//     })
//
//     // gRPC response header metadata
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
//                 Headers(func() {
//                     Attribute("name") // "name" sent in the header metadata
//                 })
//             })
//         })
//     })
//
func Headers(args interface{}) {
	fn, ok := args.(func())
	if !ok {
		eval.InvalidArgError("function", args)
		return
	}
	switch e := eval.Current().(type) {
	case *expr.GRPCResponseExpr:
		attr := &expr.AttributeExpr{}
		if eval.Execute(fn, attr) {
			e.Headers = expr.NewMappedAttributeExpr(attr)
		}
	default:
		h := headers(eval.Current())
		if h == nil {
			eval.IncompatibleDSL()
			return
		}
		eval.Execute(fn, h)
	}
}
