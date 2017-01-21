package dsl

import (
	"goa.design/goa.v2/eval"
	"goa.design/goa.v2/rest/design"
)

// Response describes a single HTTP response. A service endpoint may define
// multiple HTTP responses. The expression allows for specifying the response
// status code (as an argument of the Response function or via the Code
// function), headers (via the Header and ContentType functions) and body (via
// the Body function).
//
// By default HTTP responses use status code 200 and the endpoint response type
// attributes to define the response body shape. Also if the endpoint response
// type is a media type then the media type identifier is used to set the value
// of the "Content-Type" HTTP header in the response. In other words given the
// following type:
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
//         Response(AccountMedia)
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
//         Response(ShowResponse)
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
// Response may appear in a endpoint HTTP expression.
//
// Response takes one or two arguments. The first argument is either a HTTP
// status code or a function. The second argument is a function and is only
// allowed when the first argument is a HTTP status code.
//
// Example:
//
// Endpoint("create", func() {
//     Request(CreateRequest)
//     Response(CreateResponse)
//     HTTP(func() {
//         Response(StatusCreated) // Uses HTTP status code 201 Created and
//                                 // CreateResponse type to describe body
//
//         Response(func() {
//             Description("Response used when item already exists")
//             Code(StatusNoContent) // HTTP status code set using Code
//             Body(Empty)           // Override endpoint response type
//         })
//
//         Response(StatusAccepted, func() {
//             Description("Response used for async creations")
//             Body(func() {
//                 Attribute("taskHref", String, "API href to async task")
//             })
//         })
//     })
// })
//
func Response(val interface{}, args ...interface{}) {
	a, ok := eval.Current().(*design.ActionExpr)
	if !ok {
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
