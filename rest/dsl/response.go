package dsl

import (
	"goa.design/goa.v2/eval"
	"goa.design/goa.v2/rest/design"
)

// Response describes a single HTTP response. Response describes both success and
// error responses. When describing an error response the first argument is the
// name of the error.
//
// While a service endpoint may only define a single result type Response may
// define multiple success HTTP responses. The expression allows for specifying
// the response status code (as an argument of the Response function or via the
// Code function), headers (via the Header and ContentType functions) and body
// (via the Body function).
//
// By default success HTTP responses use status code 200 and error HTTP responses
// use status code 400. Also by default the responses use the endpoint result
// type (success responses) or error type (error responses) to define the
// response body shape.
//
// If the response type is a media type then the media type identifier is
// used to set the value of the "Content-Type" HTTP header in the response. Also
// in this case Response accepts an optional list of views corresponding to the
// media type views used by the response. Specifying no view has the same effect
// as specifying all views. In other words given the following type:
//
//     var AccountMedia = MediaType("application/vnd.goa.account", func() {
//         Attributes(func() {
//             Attribute("href", String, "Account API href")
//             Attribute("name", String, "Account name")
//         })
//         View("default", func() {
//             Attribute("href")
//             Attribute("name")
//         })
//         View("terse", func() {
//             Attribute("name")
//         })
//     })
//
// the following:
//
//     Endpoint("show", func() {
//         Response(AccountMedia)
//     })
//
// is equivalent to:
//
//     Endpoint("show", func() {
//         Response(AccountMedia, "default", "terse")
//         HTTP(func() {
//             Response(func() {
//                 Code(StatusOK)
//                 ContentType("application/vnd.goa.account")
//                 Body(AccountMedia)
//             })
//         })
//     })
//
// Also by default attributes of the response type that are not used to define
// headers are used to define the response body shape.
//
// The following:
//
//     Endpoint("show", func() {
//         Response(ShowResponse, "default")
//         HTTP(func() {
//             Response(func() {
//                 Header("href")
//             })
//         })
//     })
//
// is thus equivalent to:
//
//     Endpoint("show", func() {
//         Response(ShowResponse)
//         HTTP(func() {
//             Response(func() {
//                 Code(StatusOK)
//                 ContentType("application/vnd.goa.account")
//                 Header("href", String, "Account API href")
//                 Body(func() {
//                     Attribute("name", String, "Account name")
//                 })
//             })
//         })
//     })
//
// Response may appear in a API or service HTTP expression to define error
// responses common to all the API or service endpoints. Response may also appear
// in an endpoint HTTP expression to define both the success and error responses
// specific to the endpoint.
//
// Response takes one to three arguments. Success responses accept a status code
// or a function as first argument. If the first argument is a status code then
// a function may be given as second argument. The valid invocations are thus:
//
// * Response(func)
//
// * Response(status)
//
// * Response(status, func)
//
// Error responses additionally accept the name of the error as first argument.
//
// * Response(error_name, func)
//
// * Response(error_name, status)
//
// * Response(error_name, status, func)
//
// Example:
//
// Endpoint("create", func() {
//     Payload(CreatePayload)
//     Result(CreateResult)
//     Error("an_error")
//     HTTP(func() {
//         Response(StatusCreated) // Uses HTTP status code 201 Created and
//                                 // CreateResult type to describe body
//
//         Response(func() {
//             Description("Response used when item already exists")
//             Code(StatusNoContent) // HTTP status code set using Code
//             Body(Empty)           // Override endpoint result type
//         })
//
//         Response(StatusAccepted, func() {
//             Description("Response used for async creations")
//             Body(func() {
//                 Attribute("taskHref", String, "API href to async task")
//             })
//         })
//
//         Response("an_error", StatusConflict) // Override default of 400
//     })
// })
//
func Response(val interface{}, args ...interface{}) {
	var a *design.ActionExpr
	switch t := eval.Current().(type) {
	case *design.ActionExpr:
		name, ok := val.(string)
		if ok {
			if e := httpError(name, t, args); e != nil {
				t.HTTPErrors = append(t.HTTPErrors, e)
			}
			return
		}
		a = t
	case *design.ResourceExpr:
		name, ok := val.(string)
		if !ok {
			eval.InvalidArgError("name of error", val)
			return
		}
		if e := httpError(name, t, args); e != nil {
			t.HTTPErrors = append(t.HTTPErrors, e)
		}
		return
	case *design.RootExpr:
		name, ok := val.(string)
		if !ok {
			eval.InvalidArgError("name of error", val)
			return
		}
		if e := httpError(name, t, args); e != nil {
			t.HTTPErrors = append(t.HTTPErrors, e)
		}
		return
	default:
		eval.IncompatibleDSL()
		return
	}
	code, dsl := parseResponseArgs(val, args...)
	if code == 0 {
		code = design.StatusOK
	}
	resp := &design.HTTPResponseExpr{
		StatusCode: code,
		Parent:     a,
	}
	if dsl != nil {
		eval.Execute(dsl, resp)
	}
	a.Responses = append(a.Responses, resp)
}

// Code sets the Response status code.
func Code(code int) {
	res, ok := eval.Current().(*design.HTTPResponseExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	res.StatusCode = code
}

func parseResponseArgs(val interface{}, args ...interface{}) (code int, dsl func()) {
	switch t := val.(type) {
	case int:
		code = t
		if len(args) > 1 {
			eval.ReportError("too many arguments given to Response (%d)", len(args)+1)
			return
		}
		if len(args) == 1 {
			if d, ok := args[0].(func()); ok {
				dsl = d
			} else {
				eval.InvalidArgError("function", args[0])
				return
			}
		}
	case func():
		if len(args) > 0 {
			eval.InvalidArgError("int (HTTP status code)", val)
			return
		}
		dsl = t
	default:
		eval.InvalidArgError("int (HTTP status code) or function", val)
		return
	}
	return
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
