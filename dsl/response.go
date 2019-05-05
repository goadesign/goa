package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Response describes a HTTP or a gRPC response. Response describes both success
// and error responses. When describing an error response the first argument is
// the name of the error.
//
// While a service method may only define a single result type, Response may
// appear multiple times to define multiple success HTTP responses. In this case
// the Tag expression makes it possible to identify a result type attribute and
// a corresponding string value used to select the proper success response (each
// success response is associated with a different tag value). gRPC responses
// may only define one success response.
//
// Response may appear in an API or service expression to define error responses
// common to all the API or service methods. Response may also appear in a
// method expression to define both success and error responses specific to the
// method.
//
// Response accepts one to three arguments. Success response accepts a status
// code as first argument. If the first argument is a status code then a
// function may be given as the second argument. This function may provide a
// description and describes how to map the result type attributes to transport
// specific constructs (e.g. HTTP headers and body, gRPC metadata and message).
//
// The valid invocations for successful response are thus:
//
// * Response(status)
//
// * Response(func)
//
// * Response(status, func)
//
// Error responses additionally accept the name of the error as first argument.
//
// * Response(error_name, status)
//
// * Response(error_name, func)
//
// * Response(error_name, status, func)
//
// By default (i.e. if Response only defines a status code) then:
//
//    - success HTTP responses use code 200 (OK) and error HTTP responses use code 400 (BadRequest)
//    - success gRPC responses use code 0 (OK) and error gRPC response use code 2 (Unknown)
//    - The result type attributes are all mapped to the HTTP response body or gRPC response message.
//
// Example:
//
//    Method("create", func() {
//        Payload(CreatePayload)
//        Result(CreateResult)
//        Error("an_error")
//
//        HTTP(func() {
//            Response(StatusAccepted, func() { // HTTP status code set using argument
//                Description("Response used for async creations")
//                Tag("outcome", "accepted") // Tag identifies a result type attribute and corresponding
//                                           // value for this response to be selected.
//                Header("taskHref")         // map "taskHref" attribute to header, all others to body
//            })
//
//            Response(StatusCreated, func () {
//                Tag("outcome", "created")  // CreateResult type to describe body
//            })
//
//            Response(func() {
//                Description("Response used when item already exists")
//                Code(StatusNoContent) // HTTP status code set using Code
//                Body(Empty)           // Override method result type
//            })
//
//            Response("an_error", StatusConflict) // Override default of 400
//        })
//
//        GRPC(func() {
//            Response(CodeOK, func() {
//                Metadata("taskHref") // map "taskHref" attribute to metadata, all others to message
//            })
//
//            Response("an_error", CodeInternal, func() {
//                Description("Error returned for internal errors")
//            })
//        })
//    })
//
func Response(val interface{}, args ...interface{}) {
	name, ok := val.(string)
	switch t := eval.Current().(type) {
	case *expr.HTTPExpr:
		if !ok {
			eval.InvalidArgError("name of error", val)
			return
		}
		if e := httpError(name, t, args...); e != nil {
			t.Errors = append(t.Errors, e)
		}
	case *expr.GRPCExpr:
		if !ok {
			eval.InvalidArgError("name of error", val)
			return
		}
		if e := grpcError(name, t, args...); e != nil {
			t.Errors = append(t.Errors, e)
		}
	case *expr.HTTPServiceExpr:
		if !ok {
			eval.InvalidArgError("name of error", val)
			return
		}
		if e := httpError(name, t, args...); e != nil {
			t.HTTPErrors = append(t.HTTPErrors, e)
		}
	case *expr.HTTPEndpointExpr:
		if ok {
			if e := httpError(name, t, args...); e != nil {
				t.HTTPErrors = append(t.HTTPErrors, e)
			}
			return
		}
		code, fn := parseResponseArgs(val, args...)
		if code == 0 {
			code = expr.StatusOK
		}
		resp := &expr.HTTPResponseExpr{
			StatusCode: code,
			Parent:     t,
		}
		if fn != nil {
			eval.Execute(fn, resp)
		}
		t.Responses = append(t.Responses, resp)
	case *expr.GRPCServiceExpr:
		if !ok {
			eval.InvalidArgError("name of error", val)
			return
		}
		if e := grpcError(name, t, args...); e != nil {
			t.GRPCErrors = append(t.GRPCErrors, e)
		}
	case *expr.GRPCEndpointExpr:
		if ok {
			// error response
			if e := grpcError(name, t, args...); e != nil {
				t.GRPCErrors = append(t.GRPCErrors, e)
			}
			return
		}
		code, fn := parseResponseArgs(val, args...)
		resp := &expr.GRPCResponseExpr{
			StatusCode: code,
			Parent:     t,
		}
		if fn != nil {
			eval.Execute(fn, resp)
		}
		t.Response = resp
	default:
		eval.IncompatibleDSL()
	}
}

// Code sets the Response status code.
//
// Code must appear in a Response expression.
//
// Code accepts one argument: the HTTP or gRPC status code.
func Code(code int) {
	switch t := eval.Current().(type) {
	case *expr.HTTPResponseExpr:
		t.StatusCode = code
	case *expr.GRPCResponseExpr:
		t.StatusCode = code
	default:
		eval.IncompatibleDSL()
	}
}

func grpcError(n string, p eval.Expression, args ...interface{}) *expr.GRPCErrorExpr {
	if len(args) == 0 {
		eval.ReportError("not enough arguments, use Response(name, status), Response(name, status, func()) or Response(name, func())")
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
		code = CodeUnknown
	}
	resp := &expr.GRPCResponseExpr{
		StatusCode: code,
		Parent:     p,
	}
	if fn != nil {
		eval.Execute(fn, resp)
	}
	return &expr.GRPCErrorExpr{
		Name:     n,
		Response: resp,
	}
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

func httpError(n string, p eval.Expression, args ...interface{}) *expr.HTTPErrorExpr {
	if len(args) == 0 {
		eval.ReportError("not enough arguments, use Response(name, status), Response(name, status, func()) or Response(name, func())")
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
		code = expr.StatusBadRequest
	}
	resp := &expr.HTTPResponseExpr{
		StatusCode: code,
		Parent:     p,
	}
	if fn != nil {
		eval.Execute(fn, resp)
	}
	return &expr.HTTPErrorExpr{
		Name:     n,
		Response: resp,
	}
}
