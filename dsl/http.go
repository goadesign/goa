package dsl

import (
	"fmt"
	"strings"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

const (
	StatusContinue           = expr.StatusContinue
	StatusSwitchingProtocols = expr.StatusSwitchingProtocols
	StatusProcessing         = expr.StatusProcessing

	StatusOK                   = expr.StatusOK
	StatusCreated              = expr.StatusCreated
	StatusAccepted             = expr.StatusAccepted
	StatusNonAuthoritativeInfo = expr.StatusNonAuthoritativeInfo
	StatusNoContent            = expr.StatusNoContent
	StatusResetContent         = expr.StatusResetContent
	StatusPartialContent       = expr.StatusPartialContent
	StatusMultiStatus          = expr.StatusMultiStatus
	StatusAlreadyReported      = expr.StatusAlreadyReported
	StatusIMUsed               = expr.StatusIMUsed

	StatusMultipleChoices  = expr.StatusMultipleChoices
	StatusMovedPermanently = expr.StatusMovedPermanently
	StatusFound            = expr.StatusFound
	StatusSeeOther         = expr.StatusSeeOther
	StatusNotModified      = expr.StatusNotModified
	StatusUseProxy         = expr.StatusUseProxy

	StatusTemporaryRedirect = expr.StatusTemporaryRedirect
	StatusPermanentRedirect = expr.StatusPermanentRedirect

	StatusBadRequest                   = expr.StatusBadRequest
	StatusUnauthorized                 = expr.StatusUnauthorized
	StatusPaymentRequired              = expr.StatusPaymentRequired
	StatusForbidden                    = expr.StatusForbidden
	StatusNotFound                     = expr.StatusNotFound
	StatusMethodNotAllowed             = expr.StatusMethodNotAllowed
	StatusNotAcceptable                = expr.StatusNotAcceptable
	StatusProxyAuthRequired            = expr.StatusProxyAuthRequired
	StatusRequestTimeout               = expr.StatusRequestTimeout
	StatusConflict                     = expr.StatusConflict
	StatusGone                         = expr.StatusGone
	StatusLengthRequired               = expr.StatusLengthRequired
	StatusPreconditionFailed           = expr.StatusPreconditionFailed
	StatusRequestEntityTooLarge        = expr.StatusRequestEntityTooLarge
	StatusRequestURITooLong            = expr.StatusRequestURITooLong
	StatusUnsupportedResultType        = expr.StatusUnsupportedResultType
	StatusRequestedRangeNotSatisfiable = expr.StatusRequestedRangeNotSatisfiable
	StatusExpectationFailed            = expr.StatusExpectationFailed
	StatusTeapot                       = expr.StatusTeapot
	StatusUnprocessableEntity          = expr.StatusUnprocessableEntity
	StatusLocked                       = expr.StatusLocked
	StatusFailedDependency             = expr.StatusFailedDependency
	StatusUpgradeRequired              = expr.StatusUpgradeRequired
	StatusPreconditionRequired         = expr.StatusPreconditionRequired
	StatusTooManyRequests              = expr.StatusTooManyRequests
	StatusRequestHeaderFieldsTooLarge  = expr.StatusRequestHeaderFieldsTooLarge
	StatusUnavailableForLegalReasons   = expr.StatusUnavailableForLegalReasons

	StatusInternalServerError           = expr.StatusInternalServerError
	StatusNotImplemented                = expr.StatusNotImplemented
	StatusBadGateway                    = expr.StatusBadGateway
	StatusServiceUnavailable            = expr.StatusServiceUnavailable
	StatusGatewayTimeout                = expr.StatusGatewayTimeout
	StatusHTTPVersionNotSupported       = expr.StatusHTTPVersionNotSupported
	StatusVariantAlsoNegotiates         = expr.StatusVariantAlsoNegotiates
	StatusInsufficientStorage           = expr.StatusInsufficientStorage
	StatusLoopDetected                  = expr.StatusLoopDetected
	StatusNotExtended                   = expr.StatusNotExtended
	StatusNetworkAuthenticationRequired = expr.StatusNetworkAuthenticationRequired
)

// HTTP defines the HTTP transport specific properties of an API, a service or a
// single method. The function maps the method payload and result types to HTTP
// properties such as parameters (via path wildcards or query strings), request
// or response headers, request or response bodies as well as response status
// code. HTTP also defines HTTP specific properties such as the method endpoint
// URLs and HTTP methods.
//
// The functions that appear in HTTP such as Header, Param or Body may take
// advantage of the method payload or result types (depending on whether they
// appear when describing the HTTP request or response). The properties of the
// header, parameter or body attributes inherit the properties of the attributes
// with the same names that appear in the method payload or result types.
//
// HTTP must appear in API, a Service or an Method expression.
//
// HTTP accepts an optional argument which is the defining DSL function.
//
// Example:
//
//    var _ = API("calc", func() {
//        HTTP(func() {
//            Path("/api") // Prefix to HTTP path of all requests.
//        })
//    })
//
// Example:
//
//    var _ = Service("calculator", func() {
//        Error("unauthorized")
//
//        HTTP(func() {
//            Path("/calc")      // Prefix to all request paths
//            Error("unauthorized", StatusUnauthorized) // Define "unauthorized"
//                               // error HTTP response status code.
//            Parent("account")  // Parent service, used to prefix request
//                               // paths.
//            CanonicalMethod("show") // Method whose path is used to prefix
//                                    // the paths of child service.
//        })
//
//        Method("div", func() {
//            Description("Divide two operands.")
//            Payload(Operands)
//            Error("div_by_zero")
//
//            HTTP(func() {
//                GET("/div/{left}/{right}") // Define HTTP route. The "left"
//                                           // and "right" parameter properties
//                                           // are inherited from the
//                                           // corresponding Operands attributes.
//                Param("integer:int")       // Load "integer" attribute of
//                                           // Operands from "int" query string.
//                Header("requestID:X-RequestId")  // Load "requestID" attribute
//                                                 // of Operands from
//                                                 // X-RequestId header
//                Response(StatusOK)               // Use status 200 on success
//                Error("div_by_zero", BadRequest) // Use status code 400 for
//                                                 // "div_by_zero" responses
//            })
//        })
//    })
//
func HTTP(fns ...func()) {
	if len(fns) > 1 {
		eval.InvalidArgError("zero or one function", fmt.Sprintf("%d functions", len(fns)))
		return
	}
	fn := func() {}
	if len(fns) == 1 {
		fn = fns[0]
	}
	switch actual := eval.Current().(type) {
	case *expr.APIExpr:
		eval.Execute(fn, expr.Root)
	case *expr.ServiceExpr:
		res := expr.Root.API.HTTP.ServiceFor(actual)
		res.DSLFunc = fn
	case *expr.MethodExpr:
		res := expr.Root.API.HTTP.ServiceFor(actual.Service)
		act := res.EndpointFor(actual.Name, actual)
		act.DSLFunc = fn
	default:
		eval.IncompatibleDSL()
	}
}

// Consumes adds a MIME type to the list of MIME types the APIs supports when
// accepting requests. While the DSL supports any MIME type, the code generator
// only knows to generate the code for "application/json", "application/xml" and
// "application/gob". The service code must provide the decoders for other MIME
// types.
//
// Consumes must appear in the HTTP expression of API.
//
// Consumes accepts one or more strings corresponding to the MIME types.
//
// Example:
//
//    API("cellar", func() {
//        // ...
//        HTTP(func() {
//            Consumes("application/json", "application/xml")
//            // ...
//        })
//    })
//
func Consumes(args ...string) {
	switch e := eval.Current().(type) {
	case *expr.RootExpr:
		e.API.HTTP.Consumes = append(e.API.HTTP.Consumes, args...)
	default:
		eval.IncompatibleDSL()
	}
}

// Produces adds a MIME type to the list of MIME types the APIs supports when
// writing responses. While the DSL supports any MIME type, the code generator
// only knows to generate the code for "application/json", "application/xml" and
// "application/gob". The service code must provide the encoders for other MIME
// types.
//
// Produces must appear in the HTTP expression of API.
//
// Produces accepts one or more strings corresponding to the MIME types.
//
// Example:
//
//    API("cellar", func() {
//        // ...
//        HTTP(func() {
//            Produces("application/json", "application/xml")
//            // ...
//        })
//    })
//
func Produces(args ...string) {
	switch e := eval.Current().(type) {
	case *expr.RootExpr:
		e.API.HTTP.Produces = append(e.API.HTTP.Produces, args...)
	default:
		eval.IncompatibleDSL()
	}
}

// Path defines an API or service base path, i.e. a common HTTP path prefix to
// all the API or service methods. The path may define wildcards (see GET for a
// description of the wildcard syntax). The corresponding parameters must be
// described using Params. Multiple base paths may be defined for services.
//
// Path must appear in a API HTTP expression or a Service HTTP expression.
//
// Path accepts one argument: the HTTP path prefix.
func Path(val string) {
	switch def := eval.Current().(type) {
	case *expr.RootExpr:
		if expr.Root.API.HTTP.Path != "" {
			eval.ReportError(`only one base path may be specified for an API, got base paths %q and %q`, expr.Root.API.HTTP.Path, val)
		}
		expr.Root.API.HTTP.Path = val
	case *expr.HTTPServiceExpr:
		if !strings.HasPrefix(val, "//") {
			rp := expr.Root.API.HTTP.Path
			awcs := expr.ExtractHTTPWildcards(rp)
			wcs := expr.ExtractHTTPWildcards(val)
			for _, awc := range awcs {
				for _, wc := range wcs {
					if awc == wc {
						eval.ReportError(`duplicate wildcard "%s" in API and service base paths`, wc)
					}
				}
			}
		}
		def.Paths = append(def.Paths, val)
	default:
		eval.IncompatibleDSL()
	}
}

// GET defines a route using the GET HTTP method. The route may use wildcards to
// define path parameters. Wildcards start with '{' or with '{*' and end with
// '}'. They must appear after a '/'.
//
// A wildcard that starts with '{' matches a section of the path (the value in
// between two slashes).
//
// A wildcard that starts with '{*' matches the rest of the path. Such wildcards
// must terminate the path.
//
// GET must appear in a method HTTP function.
//
// GET accepts one argument which is the request path.
//
// Example:
//
//     var _ = Service("Manager", func() {
//         Method("GetAccount", func() {
//             Payload(GetAccount)
//             Result(Account)
//             HTTP(func() {
//                 GET("/{accountID}/details")
//                 GET("/{*accountPath}")
//             })
//         })
//     })
func GET(path string) *expr.RouteExpr {
	return route("GET", path)
}

// HEAD creates a route using the HEAD HTTP method. See GET.
func HEAD(path string) *expr.RouteExpr {
	return route("HEAD", path)
}

// POST creates a route using the POST HTTP method. See GET.
func POST(path string) *expr.RouteExpr {
	return route("POST", path)
}

// PUT creates a route using the PUT HTTP method. See GET.
func PUT(path string) *expr.RouteExpr {
	return route("PUT", path)
}

// DELETE creates a route using the DELETE HTTP method. See GET.
func DELETE(path string) *expr.RouteExpr {
	return route("DELETE", path)
}

// OPTIONS creates a route using the OPTIONS HTTP method. See GET.
func OPTIONS(path string) *expr.RouteExpr {
	return route("OPTIONS", path)
}

// TRACE creates a route using the TRACE HTTP method. See GET.
func TRACE(path string) *expr.RouteExpr {
	return route("TRACE", path)
}

// CONNECT creates a route using the CONNECT HTTP method. See GET.
func CONNECT(path string) *expr.RouteExpr {
	return route("CONNECT", path)
}

// PATCH creates a route using the PATCH HTTP method. See GET.
func PATCH(path string) *expr.RouteExpr {
	return route("PATCH", path)
}

func route(method, path string) *expr.RouteExpr {
	r := &expr.RouteExpr{Method: method, Path: path}
	a, ok := eval.Current().(*expr.HTTPEndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return r
	}
	r.Endpoint = a
	a.Routes = append(a.Routes, r)
	return r
}

// Header describes a single HTTP header or gRPC metadata header. The properties
// (description, type, validation etc.) of a header are inherited from the
// request or response type attribute with the same name by default.
//
// Header must appear in the API HTTP expression (to define request headers
// common to all the API endpoints), a specific method HTTP expression (to
// define request headers) or a Response expression (to define the response
// headers). Header may also appear in a method GRPC expression (to define
// headers sent in message metadata), or in a Response expression (to define
// headers sent in result metadata). Finally Header may also appear in a Headers
// expression.
//
// Header accepts the same arguments as the Attribute function. The header name
// may define a mapping between the attribute name and the HTTP header name when
// they differ. The mapping syntax is "name of attribute:name of header".
//
// Example:
//
//    var _ = Service("account", func() {
//        Method("create", func() {
//            Payload(CreatePayload)
//            Result(Account)
//            HTTP(func() {
//                Header("auth:Authorization", String, "Auth token", func() {
//                    Pattern("^Bearer [^ ]+$")
//                })
//                Response(StatusCreated, func() {
//                    Header("href") // Inherits description, type, validations
//                                   // etc. from Account href attribute
//                })
//            })
//        })
//    })
//
func Header(name string, args ...interface{}) {
	h := headers(eval.Current())
	if h == nil {
		eval.IncompatibleDSL()
		return
	}
	if name == "" {
		eval.ReportError("header name cannot be empty")
	}
	eval.Execute(func() { Attribute(name, args...) }, h.AttributeExpr)
	h.Remap()
}

// Params groups a set of Param expressions. It makes it possible to list
// required parameters using the Required function.
//
// Params must appear in an API or Service HTTP expression to define the API or
// service base path and query string parameters. Params may also appear in an
// method HTTP expression to define the HTTP endpoint path and query string
// parameters.
//
// Params accepts one argument which is a function listing the parameters.
//
// Example:
//
//     var _ = API("cellar", func() {
//         HTTP(func() {
//             Params(func() {
//                 Param("version", String, "API version", func() {
//                     Enum("1.0", "2.0")
//                 })
//                 Required("version")
//             })
//         })
//     })
//
func Params(args interface{}) {
	p := params(eval.Current())
	if p == nil {
		eval.IncompatibleDSL()
		return
	}
	fn, ok := args.(func())
	if !ok {
		eval.InvalidArgError("function", args)
		return
	}
	eval.Execute(fn, p)
}

// Param describes a single HTTP request path or query string parameter.
//
// Param must appear in the API HTTP expression (to define request parameters
// common to all the API endpoints), a service HTTP expression to define common
// parameters to all the service methods or a specific method HTTP
// expression. Param may also appear in a Params expression.
//
// Param accepts the same arguments as the Function Attribute.
//
// The name may be of the form "name of attribute:name of parameter" to define a
// mapping between the attribute and parameter names when they differ.
//
// Example:
//
//    var ShowPayload = Type("ShowPayload", func() {
//        Attribute("id", UInt64, "Account ID")
//        Attribute("version", String, "Version", func() {
//            Enum("1.0", "2.0")
//        })
//    })
//
//    var _ = Service("account", func() {
//        HTTP(func() {
//            Path("/{parentID}")
//            Param("parentID", UInt64, "ID of parent account")
//        })
//        Method("show", func() {  // default response type.
//            Payload(ShowPayload)
//            Result(AccountResult)
//            HTTP(func() {
//                GET("/{id}")           // HTTP request uses ShowPayload "id"
//                                       // attribute to define "id" parameter.
//                Params(func() {        // Params makes it possible to group
//                                       // Param expressions.
//                    Param("version:v") // "version" of ShowPayload to define
//                                       // path and query string parameters.
//                                       // Query string "v" maps to attribute
//                                       // "version" of ShowPayload.
//                })
//            })
//        })
//    })
//
func Param(name string, args ...interface{}) {
	p := params(eval.Current())
	if p == nil {
		eval.IncompatibleDSL()
		return
	}
	if name == "" {
		eval.ReportError("parameter name cannot be empty")
	}
	eval.Execute(func() { Attribute(name, args...) }, p.AttributeExpr)
	p.Remap()
}

// MapParams describes the query string parameters in a HTTP request.
//
// MapParams must appear in a Method HTTP expression to map the query string
// parameters with the Method's Payload.
//
// MapParams accepts one optional argument which specifes the Payload
// attribute to which the query string parameters must be mapped. This Payload
// attribute must be a map. If no argument is specified, the query string
// parameters are mapped with the entire Payload (the Payload must be a map).
//
// Example:
//
//     var _ = Service("account", func() {
//         Method("index", func() {
//             Payload(MapOf(String, Int))
//             HTTP(func() {
//                 GET("/")
//                 MapParams()
//             })
//         })
//    })
//
//    var _ = Service("account", func() {
//        Method("show", func() {
//            Payload(func() {
//                Attribute("p", MapOf(String, String))
//                Attribute("id", String)
//            })
//            HTTP(func() {
//                GET("/{id}")
//                MapParams("p")
//            })
//        })
//    })
//
func MapParams(args ...interface{}) {
	if len(args) > 1 {
		eval.ReportError("too many arguments")
	}
	e, ok := eval.Current().(*expr.HTTPEndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	var mapName string
	if len(args) > 0 {
		mapName, ok = args[0].(string)
		if !ok {
			eval.ReportError("argument must be a string")
		}
	}
	e.MapQueryParams = &mapName
}

// MultipartRequest indicates that HTTP requests made to the method use
// MIME multipart encoding as defined in RFC 2046.
//
// MultipartRequest must appear in a HTTP endpoint expression.
//
// goa generates a custom encoder that writes the payload for requests made to
// HTTP endpoints that use MultipartRequest. The generated encoder accept a
// user provided function that does the actual mapping of the payload to the
// multipart content. The user provided function accepts a multipart writer
// and a reference to the payload and is responsible for encoding the payload.
// goa also generates a custom decoder that reads back the multipart content
// into the payload struct. The generated decoder also accepts a user provided
// function that takes a multipart reader and a reference to the payload struct
// as parameter. The user provided decoder is responsible for decoding the
// multipart content into the payload. The example command generates a default
// implementation for the user decoder and encoder.
//
func MultipartRequest() {
	e, ok := eval.Current().(*expr.HTTPEndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	e.MultipartRequest = true
}

// Body describes a HTTP request or response body.
//
// Body must appear in a Method HTTP expression to define the request body or in
// an Error or Result HTTP expression to define the response body. If Body is
// absent then the body is built using the HTTP endpoint request or response
// type attributes not used to describe parameters (request only) or headers.
//
// Body accepts one argument which describes the shape of the body, it can be:
//
//  - The name of an attribute of the request or response type. In this case the
//    attribute type describes the shape of the body.
//
//  - A function listing the body attributes. The attributes inherit the
//    properties (description, type, validations etc.) of the request or
//    response type attributes with identical names.
//
// Assuming the type:
//
//     var CreatePayload = Type("CreatePayload", func() {
//         Attribute("name", String, "Name of account")
//     })
//
// The following:
//
//     Method("create", func() {
//         Payload(CreatePayload)
//     })
//
// is equivalent to:
//
//     Method("create", func() {
//         Payload(CreatePayload)
//         HTTP(func() {
//             Body(func() {
//                 Attribute("name")
//             })
//         })
//     })
//
func Body(args ...interface{}) {
	if len(args) == 0 {
		eval.ReportError("not enough arguments, use Body(name), Body(type), Body(func()) or Body(type, func())")
		return
	}

	var (
		ref    *expr.AttributeExpr
		setter func(*expr.AttributeExpr)
		kind   string
	)

	// Figure out reference type and setter function
	switch e := eval.Current().(type) {
	case *expr.HTTPEndpointExpr:
		ref = e.MethodExpr.Payload
		setter = func(att *expr.AttributeExpr) {
			e.Body = att
		}
		kind = "Request"
	case *expr.HTTPErrorExpr:
		ref = e.ErrorExpr.AttributeExpr
		setter = func(att *expr.AttributeExpr) {
			if e.Response == nil {
				e.Response = &expr.HTTPResponseExpr{}
			}
			e.Response.Body = att
		}
		kind = "Error"
		if e.Name != "" {
			kind += " " + e.Name
		}
	case *expr.HTTPResponseExpr:
		ref = e.Parent.(*expr.HTTPEndpointExpr).MethodExpr.Result
		setter = func(att *expr.AttributeExpr) {
			e.Body = att
		}
		kind = "Response"
	default:
		eval.IncompatibleDSL()
		return
	}

	// Now initialize target attribute and DSL if any
	var (
		attr *expr.AttributeExpr
		fn   func()
	)
	switch a := args[0].(type) {
	case string:
		if !expr.IsObject(ref.Type) {
			eval.ReportError("%s type must be an object with an attribute with name %#v, got %T", kind, a, ref.Type)
			return
		}
		attr = ref.Find(a)
		if attr == nil {
			eval.ReportError("%s type does not have an attribute named %#v", kind, a)
			return
		}
		attr = expr.DupAtt(attr)
		if attr.Meta == nil {
			attr.Meta = expr.MetaExpr{"origin:attribute": []string{a}}
		} else {
			attr.Meta["origin:attribute"] = []string{a}
		}
		if rt, ok := attr.Type.(*expr.ResultTypeExpr); ok {
			// If the attribute type is a result type add the type to the
			// GeneratedTypes so that the type's DSLFunc is executed.
			*expr.Root.GeneratedTypes = append(*expr.Root.GeneratedTypes, rt)
		}
	case expr.UserType:
		attr = &expr.AttributeExpr{Type: a}
		if len(args) > 1 {
			var ok bool
			fn, ok = args[1].(func())
			if !ok {
				eval.ReportError("second argument must be a function")
			}
		}
	case func():
		fn = a
		if ref == nil {
			eval.ReportError("Body is set but Payload is not defined")
			return
		}
		attr = ref
	default:
		eval.InvalidArgError("attribute name, user type or DSL", a)
		return
	}

	// Set body attribute
	if fn != nil {
		eval.Execute(fn, attr)
	}
	if attr.Meta == nil {
		attr.Meta = expr.MetaExpr{}
	}
	attr.Meta["http:body"] = []string{}
	setter(attr)
}

// Parent sets the name of the parent service. The parent service canonical
// method path is used as prefix for all the service HTTP endpoint paths.
//
// Parent must appear in a Service expression.
//
// Parent accepts one argument: the name of the parent service.
func Parent(name string) {
	r, ok := eval.Current().(*expr.HTTPServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	r.ParentName = name
}

// CanonicalMethod sets the name of the service canonical method. The canonical
// method endpoint HTTP path is used to prefix the paths to child service
// endpoints (a child service is a service that uses the Parent function). The
// default value is "show".
//
// CanonicalMethod must appear in the HTTP expresssion of a Service.
//
// CanonicalMethod accepts one argument: the name of the canonical service
// method.
func CanonicalMethod(name string) {
	r, ok := eval.Current().(*expr.HTTPServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	r.CanonicalEndpointName = name
}

// Tag identifies a method result type field and a value. The algorithm that
// encodes the result into the HTTP response iterates through the responses and
// uses the first response that has a matching tag (that is for which the result
// field with the tag name matches the tag value). There must be one and only
// one response with no Tag expression, this response is used when no other tag
// matches.
//
// Tag must appear in Response.
//
// Tag accepts two arguments: the name of the field and the (string) value.
//
// Example:
//
//    Method("create", func() {
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
	res, ok := eval.Current().(*expr.HTTPResponseExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	res.Tag = [2]string{name, value}
}

// ContentType sets the value of the Content-Type response header.
//
// ContentType may appear in a ResultType or a Response expression.
// ContentType accepts one argument: the mime type as defined by RFC 6838.
//
//    var _ = ResultType("application/vnd.myapp.mytype", func() {
//        ContentType("application/json")
//    })
//
//    var _ = Method("add", func() {
//	  HTTP(func() {
//            Response(OK, func() {
//                ContentType("application/json")
//            })
//        })
//    })
//
func ContentType(typ string) {
	switch actual := eval.Current().(type) {
	case *expr.ResultTypeExpr:
		actual.ContentType = typ
	case *expr.HTTPResponseExpr:
		actual.ContentType = typ
	default:
		eval.IncompatibleDSL()
	}
}

// headers returns the mapped attribute containing the headers for the given
// expression if it's either the root, a service or an endpoint - nil otherwise.
func headers(exp eval.Expression) *expr.MappedAttributeExpr {
	switch e := exp.(type) {
	case *expr.RootExpr:
		if e.API.HTTP.Headers == nil {
			e.API.HTTP.Headers = expr.NewEmptyMappedAttributeExpr()
		}
		return e.API.HTTP.Headers
	case *expr.HTTPServiceExpr:
		if e.Headers == nil {
			e.Headers = expr.NewEmptyMappedAttributeExpr()
		}
		return e.Headers
	case *expr.HTTPEndpointExpr:
		if e.Headers == nil {
			e.Headers = expr.NewEmptyMappedAttributeExpr()
		}
		return e.Headers
	case *expr.HTTPResponseExpr:
		if e.Headers == nil {
			e.Headers = expr.NewEmptyMappedAttributeExpr()
		}
		return e.Headers
	case *expr.MappedAttributeExpr:
		return e
	default:
		return nil
	}
}

// params returns the mapped attribute containing the path and query params for
// the given expression if it's either the root, a API server, a service or an
// endpoint - nil otherwise.
func params(exp eval.Expression) *expr.MappedAttributeExpr {
	switch e := exp.(type) {
	case *expr.RootExpr:
		if e.API.HTTP.Params == nil {
			e.API.HTTP.Params = expr.NewEmptyMappedAttributeExpr()
		}
		return e.API.HTTP.Params
	case *expr.HTTPServiceExpr:
		if e.Params == nil {
			e.Params = expr.NewEmptyMappedAttributeExpr()
		}
		return e.Params
	case *expr.HTTPEndpointExpr:
		if e.Params == nil {
			e.Params = expr.NewEmptyMappedAttributeExpr()
		}
		return e.Params
	case *expr.MappedAttributeExpr:
		return e
	default:
		return nil
	}
}
