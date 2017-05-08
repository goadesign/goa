package rest

import (
	"goa.design/goa.v2/design/rest"
	"goa.design/goa.v2/eval"
)

// Response describes a single HTTP response. Response describes both success and
// error responses. When describing an error response the first argument is the
// name of the error.
//
// While a service endpoint may only define a single result type Response may be
// called multiple times to define multiple success HTTP responses. In this case
// the Tag expression makes it possible to specify the name of a field in the
// endpoint result type and a value that the field must have for the
// corresponding response to be sent. The tag field must be of type String.
//
// Response allows specifying the response status code as an argument or via the
// Code expression, headers via the Header and ContentType expressions and body
// via the Body expression.
//
// By default success HTTP responses use status code 200 and error HTTP responses
// use status code 400. Also by default the responses use the endpoint result
// type (success responses) or error type (error responses) to define the
// response body shape.
//
// Additionally if the response type is a media type then the "Content-Type"
// response header is set with the corresponding content type (either the value
// set with ContentType in the media type DSL or the media type identifier).
//
// In other words given the following type:
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
//    Endpoint("create", func() {
//        Payload(CreatePayload)
//        Result(CreateResult)
//        Error("an_error")
//        HTTP(func() {
//            Response(StatusCreated) // Uses HTTP status code 201 Created and
//                                    // CreateResult type to describe body
//
//            Response(func() {
//                Description("Response used when item already exists")
//                Code(StatusNoContent) // HTTP status code set using Code
//                Body(Empty)           // Override endpoint result type
//            })
//
//            Response(StatusAccepted, func() {
//                Description("Response used for async creations")
//                Body(func() {
//                    Attribute("taskHref", String, "API href to async task")
//                })
//            })
//
//            Response("an_error", StatusConflict) // Override default of 400
//        })
//    })
//
func Response(val interface{}, args ...interface{}) {
	name, ok := val.(string)
	switch t := eval.Current().(type) {
	case *rest.ResourceExpr:
		if !ok {
			eval.InvalidArgError("name of error", val)
			return
		}
		if e := httpError(name, t, args...); e != nil {
			t.HTTPErrors = append(t.HTTPErrors, e)
		}
	case *rest.RootExpr:
		if !ok {
			eval.InvalidArgError("name of error", val)
			return
		}
		if e := httpError(name, t, args...); e != nil {
			t.HTTPErrors = append(t.HTTPErrors, e)
		}
	case *rest.ActionExpr:
		if ok {
			if e := httpError(name, t, args...); e != nil {
				t.HTTPErrors = append(t.HTTPErrors, e)
			}
			return
		}
		code, fn := parseResponseArgs(val, args...)
		if code == 0 {
			code = rest.StatusOK
		}
		resp := &rest.HTTPResponseExpr{
			StatusCode: code,
			Parent:     t,
		}
		if fn != nil {
			eval.Execute(fn, resp)
		}
		t.Responses = append(t.Responses, resp)
	default:
		eval.IncompatibleDSL()
	}
}

// Tag identifies a endpoint result type field and a value. The algorithm that
// encodes the result into the HTTP response iterates through the responses and
// uses the first response that has a matching tag (that is for which the result
// field with the tag name matches the tag value). There must be one and only
// one response with no Tag expression, this response is used when no other tag
// matches.
//
// Tag may appear in Response.
// Tag accepts two arguments: the name of the field and the (string) value.
//
// Example:
//
//    Endpoint("create", func() {
//        Result(CreateResult)
//        HTTP(func() {
//            Response(StatusCreated, func() {
//                Tag("outcome", "created") // Assumes CreateResult has attribute
//                                          // "outcome" which may be "created"
//                                          // or "accepted"
//            })
//
//            Response(StatusAccepted, func() {
//                Tag("outcome", "accepted")
//            })
//
//            Response(StatusOK)            // Default response if "outcome" is
//                                          // neither "created" nor "accepted"
//        })
//    })
//
func Tag(name, value string) {
	res, ok := eval.Current().(*rest.HTTPResponseExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	res.Tag = [2]string{name, value}
}

// Code sets the Response status code.
func Code(code int) {
	res, ok := eval.Current().(*rest.HTTPResponseExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	res.StatusCode = code
}

func parseResponseArgs(val interface{}, args ...interface{}) (code int, fn func()) {
	switch t := val.(type) {
	case int:
		code = t
		if len(args) > 1 {
			eval.ReportError("too many arguments given to Response (%d)", len(args)+1)
			return
		}
		if len(args) == 1 {
			if d, ok := args[0].(func()); ok {
				fn = d
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
		fn = t
	default:
		eval.InvalidArgError("int (HTTP status code) or function", val)
		return
	}
	return
}

func httpError(n string, p eval.Expression, args ...interface{}) *rest.HTTPErrorExpr {
	if len(args) == 0 {
		eval.ReportError("not enough arguments, use Error(name, status), Error(name, status, func()) or Error(name, func())")
		return nil
	}
	var (
		code int
		fn   func()
		val  interface{}
	)
	val = args[0]
	args = args[1:]
	code, fn = parseResponseArgs(val, args...)
	if code == 0 {
		code = rest.StatusBadRequest
	}
	resp := &rest.HTTPResponseExpr{
		StatusCode: code,
		Parent:     p,
	}
	if fn != nil {
		eval.Execute(fn, resp)
	}
	return &rest.HTTPErrorExpr{
		Name:     n,
		Response: resp,
	}
}
