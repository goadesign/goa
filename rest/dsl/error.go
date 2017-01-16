package dsl

import (
	goadsl "goa.design/goa.v2/dsl"
	"goa.design/goa.v2/eval"
	"goa.design/goa.v2/rest/design"
)

// Error has two usages:
//
// 1. Outside of an HTTP Expression
//
// Outside of an HTTP expression Error describes an endpoint error response. The
// description includes a unique name (in the scope of the endpoint), an optional
// type and description as well as a function that further describes the type.
// If no type is specified then the goa ErrorMedia type is used. The syntax is
// identical to the Attribute expression.
//
// Error may appear in the Service (to define error responses that apply to all
// the service endpoints) or Endpoint expressions.
//
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
// 2. Inside an HTTP Expression
//
// Inside of an HTTP expression Error describes an error HTTP response. The
// Error expression syntax is identical to the HTTP Response expression syntax
// with the addition of the error name as first argument. The expression allows
// for specifying the response status code (as an argument of the Error function
// or via the Code function), headers (via the Header and ContentType functions)
// and body (via the Body function).
//
// By default error HTTP responses use status code 400 and the error type
// attributes to define the response body shape. Also if the error
// type is a media type then the media type identifier is used to set the value
// of the "Content-Type" HTTP header in the response.
//
// Error may appear in a service or an endpoint HTTP expression.
//
// Error takes two or three arguments. The first argument is the name of the
// error. The second argument is either a HTTP status code or a function. The
// last argument is a function and is only allowed when the second argument is a
// HTTP status code.
//
// Example:
//
//     var UnauthorizedErrorMedia = MediaType("application/vnd.goa.unauthorized", func() {
//         // ...
//     })
//
//     var NotFoundErrorMedia = MediaType("application/vnd.goa.not_found", func() {
//         // ...
//     })
//
//     var _ = Service("account", func() {
//         Error("unauthorized", UnauthorizedErrorMedia) // Applies to all the
//         HTTP(func() {                                 // service endpoints
//             Error("unauthorized", StatusUnauthorized) // Uses HTTP status code
//         })                                            // 401 and error type to
//                                                       // describe body shape
//         Endpoint("show", func() {
//             Response(AccountMedia)
//             Error("not_found", NotFoundErrorMedia)
//             HTTP(func() {
//                 Error("not_found", StatusNotFound, func() {
//                     Header("id:Account-ID") // Uses error type "id" attribute
//                 })                          // to set Account-ID response
//             })                              // header and other attributes
//         })                                  // to describe body shape
//     })
//
func Error(name string, args ...interface{}) {
	switch t := eval.Current().(type) {
	case *design.ActionExpr:
		if e := httpError(name, t, args); e != nil {
			t.HTTPErrors = append(t.HTTPErrors, e)
		}
	case *design.ResourceExpr:
		if e := httpError(name, t, args); e != nil {
			t.HTTPErrors = append(t.HTTPErrors, e)
		}
	case *design.RootExpr:
		if e := httpError(name, t, args); e != nil {
			t.HTTPErrors = append(t.HTTPErrors, e)
		}
	default:
		goadsl.Error(name, args...)
	}
}

func httpError(n string, p eval.Expression, args ...interface{}) *design.HTTPErrorExpr {
	if len(args) == 0 {
		eval.ReportError("not enough arguments, use Error(name, status), Error(name, status, func()) or Error(name, func())")
		return nil
	}
	var (
		code int
		dsl  func()
		val  interface{}
	)
	val = args[0]
	args = args[1:]
	code, dsl = parseResponseArgs(val, args...)
	if code == 0 {
		code = design.StatusBadRequest
	}
	resp := &design.HTTPResponseExpr{
		StatusCode: code,
		Parent:     p,
	}
	if dsl != nil {
		eval.Execute(dsl, resp)
	}
	return &design.HTTPErrorExpr{
		Name:     n,
		Response: resp,
	}
}
