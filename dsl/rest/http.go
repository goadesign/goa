package rest

import (
	"strings"

	"reflect"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
	"goa.design/goa.v2/dsl"
	"goa.design/goa.v2/eval"
)

// HTTP defines HTTP transport specific properties either on a API, service or
// on a single endpoint. The function maps the request and response types to
// HTTP properties such as parameters (via path wildcards or query strings),
// request and response headers and bodies as well as response status code. HTTP
// also defines HTTP specific properties such as the endpoint URLs and HTTP
// methods.
//
// As a special case HTTP may be used to define the response generated for
// invalid requests and internal errors (errors returned by the service
// endpoints that don't match any of the error responses defined in the design).
// This is the only use of HTTP allowed in the API expression. The attributes of
// the built in invalid request error are "id", "status", "code", "detail" and
// "meta", see ErrorMedia.
//
// The functions that appear in HTTP such as Header, Param or Body may take
// advantage of the request or response types (depending on whether they appear
// when describing the HTTP request or response). The properties of the header,
// parameter or body attributes inherit the properties of the attributes with the
// same names that appear in the request or response types. The functions may
// also define new attributes or override the existing request or response type
// attributes.
//
// HTTP may appear in API, a Service or an Endpoint expression.
//
// HTTP accepts a single argument which is the defining DSL function.
//
// Example:
//
//    var _ = API("calc", func() {
//        HTTP(func() {
//            Response(InvalidRequest, func() {
//                Header("Error-Code:code") // Use the "code" attribute of the
//                                          // invalid error struct to set the
//                                          // value of the Error-Code header.
//            })
//        })
//    }
//
// Example:
//
//    var _ = Service("calculator", func() {
//        Error(ErrAuthFailure)
//
//        HTTP(func() {
//            Path("/calc")      // Prefix to all request paths
//            Error(ErrAuthFailure, StatusUnauthorized) // Define
//                               // ErrAuthFailure HTTP response status code.
//            Parent("account")  // Parent service, used to prefix request
//                               // paths.
//            CanonicalEndpoint("add") // Endpoint whose path is used to prefix
//                                     // the paths of child service.
//        })
//
//        Endpoint("add", func() {
//            Description("Add two operands")
//            Payload(Operands)
//            Error(ErrBadRequest, ErrorMedia)
//
//            HTTP(func() {
//                GET("/add/{left}/{right}") // Define HTTP route. The "left"
//                                           // and "right" parameter properties
//                                           // are inherited from the
//                                           // corresponding Operands attributes.
//                Param("req:requestID")     // Use "requestID" attribute to
//                                           // define "req" query string
//                Header("requestID:X-RequestID")  // Use "requestID" attribute
//                                                 // of Operands to define shape
//                                                 // of X-RequestID header
//                Response(StatusNoContent)        // Use status 204 on success
//                Error(ErrBadRequest, BadRequest) // Use status code 400 for
//                                                 // ErrBadRequest responses
//            })
//
//        })
//    })
//
func HTTP(fn func()) {
	switch actual := eval.Current().(type) {
	case *rest.RootExpr:
		eval.Execute(fn, rest.Root)
	case *design.ServiceExpr:
		res := rest.Root.ResourceFor(actual)
		res.DSLFunc = fn
	case *design.EndpointExpr:
		res := rest.Root.ResourceFor(actual.Service)
		act := res.ActionFor(actual.Name, actual)
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
// Consumes may appear in the HTTP expression of API.
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
	switch def := eval.Current().(type) {
	case *rest.RootExpr:
		def.Consumes = append(rest.Root.Consumes, args...)
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
// Produces may appear in the HTTP expression of API.
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
	switch def := eval.Current().(type) {
	case *rest.RootExpr:
		def.Produces = append(rest.Root.Produces, args...)
	default:
		eval.IncompatibleDSL()
	}
}

// Path defines the API or service base path, i.e. the common path prefix to all
// the API or service endpoints. The path may define wildcards (see GET for a
// description of the wildcard syntax). The corresponding parameters must be
// described using Params.
func Path(val string) {
	switch def := eval.Current().(type) {
	case *design.APIExpr:
		rest.Root.Path = val
	case *rest.ResourceExpr:
		def.Path = val
		if !strings.HasPrefix(val, "//") {
			awcs := rest.ExtractWildcards(rest.Root.Path)
			wcs := rest.ExtractWildcards(val)
			for _, awc := range awcs {
				for _, wc := range wcs {
					if awc == wc {
						eval.ReportError(`duplicate wildcard "%s" in API and service base paths`, wc)
					}
				}
			}
		}
	default:
		eval.IncompatibleDSL()
	}
}

// Docs provides external documentation pointers for endpoints.
func Docs(fn func()) {
	docs := new(design.DocsExpr)
	if !eval.Execute(fn, docs) {
		return
	}

	switch expr := eval.Current().(type) {
	case *rest.FileServerExpr:
		expr.Docs = docs
	default:
		dsl.Docs(fn)
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
// GET may appear in an endpoint HTTP function.
// GET accepts one argument which is the request path.
//
// Example:
//
//     var _ = Service("Manager", func() {
//         Endpoint("GetAccount", func() {
//             Payload(GetAccount)
//             Result(Account)
//             HTTP(func() {
//                 GET("/{accountID}/details")
//                 GET("/{*accountPath}")
//             })
//         })
//     })
func GET(path string) *rest.RouteExpr {
	return route("GET", path)
}

// HEAD creates a route using the HEAD HTTP method. See GET.
func HEAD(path string) *rest.RouteExpr {
	return route("HEAD", path)
}

// POST creates a route using the POST HTTP method. See GET.
func POST(path string) *rest.RouteExpr {
	return route("POST", path)
}

// PUT creates a route using the PUT HTTP method. See GET.
func PUT(path string) *rest.RouteExpr {
	return route("PUT", path)
}

// DELETE creates a route using the DELETE HTTP method. See GET.
func DELETE(path string) *rest.RouteExpr {
	return route("DELETE", path)
}

// OPTIONS creates a route using the OPTIONS HTTP method. See GET.
func OPTIONS(path string) *rest.RouteExpr {
	return route("OPTIONS", path)
}

// TRACE creates a route using the TRACE HTTP method. See GET.
func TRACE(path string) *rest.RouteExpr {
	return route("TRACE", path)
}

// CONNECT creates a route using the CONNECT HTTP method. See GET.
func CONNECT(path string) *rest.RouteExpr {
	return route("CONNECT", path)
}

// PATCH creates a route using the PATCH HTTP method. See GET.
func PATCH(path string) *rest.RouteExpr {
	return route("PATCH", path)
}

func route(method, path string) *rest.RouteExpr {
	r := &rest.RouteExpr{Method: method, Path: path}
	a, ok := eval.Current().(*rest.ActionExpr)
	if !ok {
		eval.IncompatibleDSL()
		return r
	}
	r.Action = a
	a.Routes = append(a.Routes, r)
	return r
}

// Headers groups a set of Header expressions. It makes it possible to list
// required headers using the Required function.
//
// Headers may appear in an API or Service HTTP expression to define request
// headers common to all the API or service endpoints. Headers may also appear
// in an endpoint, response or error HTTP expression to define the endpoint
// request and response headers.
//
// Headers accepts one argument: Either a function listing the headers or a user
// type which must be an object and whose attributes define the headers.
//
// Example:
//
//     var _ = API("cellar", func() {
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
func Headers(args interface{}) {
	h, ok := eval.Current().(rest.HeaderHolder)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if fn, ok := args.(func()); ok {
		eval.Execute(fn, h)
		return
	}
	t, ok := args.(design.UserType)
	if !ok {
		eval.InvalidArgError("function or type", args)
		return
	}
	o := design.AsObject(t)
	if o == nil {
		eval.ReportError("type must be an object but got %s", reflect.TypeOf(args).Name())
	}
	h.Headers().Merge(&design.AttributeExpr{Type: o})
}

// Header describes a single HTTP header. The properties (description, type,
// validation etc.) of a header are inherited from the request or response type
// attribute with the same name by default.
//
// Header may appear in the API HTTP expression (to define request headers
// common to all the API endpoints), a specific endpoint HTTP expression (to
// define request headers), a Result expression (to define the response
// headers) or an Error expression (to define the error response headers). Header
// may also appear in a Headers expression.
//
// Header accepts the same arguments as the Attribute function. The header name
// may define a mapping between the attribute name and the HTTP header name when
// they differ. The mapping syntax is "name of attribute:name of header".
//
// Example:
//
//    var _ = Service("account", func() {
//        Endpoint("create", func() {
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
	h, ok := eval.Current().(rest.HeaderHolder)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if name == "" {
		eval.ReportError("header name cannot be empty")
	}
	eval.Execute(func() { dsl.Attribute(name, args...) }, h.Headers())
}

// Params groups a set of Param expressions. It makes it possible to list
// required parameters using the Required function.
//
// Params may appear in an API or Service HTTP expression to define the API or
// service base path and query string parameters. Params may also appear in an
// endpoint HTTP expression to define the endpoint path and query string
// parameters.
//
// Params accepts one argument: Either a function listing the parameters or a
// user type which must be an object and whose attributes define the parameters.
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
	h, ok := eval.Current().(rest.ParamHolder)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if fn, ok := args.(func()); ok {
		eval.Execute(fn, h)
		return
	}
	t, ok := args.(design.UserType)
	if !ok {
		eval.InvalidArgError("function or type", args)
		return
	}
	o := design.AsObject(t)
	if o == nil {
		eval.ReportError("type must be an object but got %s", reflect.TypeOf(args).Name())
	}
	h.Params().Merge(&design.AttributeExpr{Type: o})
}

// Param describes a single HTTP request path or query string parameter.
//
// Param may appear in the API HTTP expression (to define request parameters
// common to all the API endpoints), a service HTTP expression to define common
// parameters to all the service endpoints or a specific endpoint HTTP
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
//        Attribute("version", String, "Endpoint version", func() {
//            Enum("1.0", "2.0")
//        })
//    })
//
//    var _ = Service("account", func() {
//        HTTP(func() {
//            Path("/{parentID}")
//            Param("parentID", UInt64, "ID of parent account")
//        })
//        Endpoint("show", func() {  // default response type.
//            Payload(ShowPayload)
//            Result(AccountMedia)
//            HTTP(func() {
//                Routing(GET("/{id}"))  // HTTP request uses ShowPayload "id"
//                                       // attribute to define "id" parameter.
//                Params(func() {        // Params makes it possible to group
//                                       // Param expressions.
//                    Param("version:v") // "version" of ShowPayload to define
//                                       // path and query string parameters.
//                                       // Query string "v" maps to attribute
//                                       // "version" of ShowPayload.
//                    Param("csrf", String) // HTTP only parameter not defined in
//                                          // ShowPayload
//                    Required("crsf")   // Params makes it possible to list the
//                                       // required parameters.
//                })
//            })
//        })
//    })
//
func Param(name string, args ...interface{}) {
	h, ok := eval.Current().(rest.ParamHolder)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if name == "" {
		eval.ReportError("parameter name cannot be empty")
	}
	eval.Execute(func() { dsl.Attribute(name, args...) }, h.Params())
}

// Body describes a HTTP request or response body.
//
// Body may appear in a Endpoint HTTP expression to define the request body or
// in an Error or Result HTTP expression to define the response body. If Body is
// absent then the body is built using the endpoint request or response type
// attributes not used to describe parameters (request only) or headers.
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
//     Endpoint("create", func() {
//         Payload(CreatePayload)
//     })
//
// is equivalent to:
//
//     Endpoint("create", func() {
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
		ref    *design.AttributeExpr
		setter func(*design.AttributeExpr)
		kind   string
	)

	// Figure out reference type and setter function
	switch e := eval.Current().(type) {
	case *rest.ActionExpr:
		ref = e.EndpointExpr.Payload
		setter = func(att *design.AttributeExpr) {
			e.Body = att
		}
		kind = "Request"
	case *rest.HTTPErrorExpr:
		ref = e.ErrorExpr.AttributeExpr
		setter = func(att *design.AttributeExpr) {
			if e.Response == nil {
				e.Response = &rest.HTTPResponseExpr{}
			}
			e.Response.Body = att
		}
		kind = "Error"
		if e.Name != "" {
			kind += " " + e.Name
		}
	case *rest.HTTPResponseExpr:
		ref = e.Parent.(*rest.ActionExpr).EndpointExpr.Result
		setter = func(att *design.AttributeExpr) {
			e.Body = att
		}
		kind = "Response"
	default:
		eval.IncompatibleDSL()
		return
	}

	// Now initialize target attribute and DSL if any
	var (
		attr *design.AttributeExpr
		fn   func()
	)
	switch a := args[0].(type) {
	case string:
		obj := design.AsObject(ref.Type)
		if obj == nil {
			eval.ReportError("%s type must be an object with an attribute with name %#v, got %T", kind, a, ref.Type)
			return
		}
		var ok bool
		attr, ok = obj[a]
		if !ok {
			eval.ReportError("%s type does not have an attribute named %#v", kind, a)
			return
		}
	case design.UserType:
		attr = a.Attribute()
		if len(args) > 1 {
			var ok bool
			fn, ok = args[1].(func())
			if !ok {
				eval.ReportError("second argument must be a function")
			}
		}
	case func():
		fn = a
		attr = ref
	default:
		eval.InvalidArgError("attribute name, user type or DSL", a)
		return
	}

	// Set body attribute
	if fn != nil {
		eval.Execute(fn, attr)
	}
	setter(attr)
}

// Parent sets the name of the parent service. The parent service canonical
// endpoint path is used as prefix for all the service endpoint paths.
func Parent(name string) {
	r, ok := eval.Current().(*rest.ResourceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	r.ParentName = name
}

// CanonicalEndpoint sets the name of the service canonical endpoint. The
// canonical endpoint path is used to prefix the paths to any child service
// endpoint. The default value is "show".
func CanonicalEndpoint(name string) {
	r, ok := eval.Current().(*rest.ResourceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	r.CanonicalActionName = name
}
