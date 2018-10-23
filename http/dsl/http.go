package dsl

import (
	"reflect"

	"goa.design/goa/design"
	"goa.design/goa/dsl"
	"goa.design/goa/eval"
	httpdesign "goa.design/goa/http/design"
)

// HTTP defines HTTP transport specific properties on a API, a service or a
// single method. The function maps the request and response types to HTTP
// properties such as parameters (via path wildcards or query strings), request
// or response headers, request or response bodies as well as response status
// code. HTTP also defines HTTP specific properties such as the method endpoint
// URLs and HTTP methods.
//
// The functions that appear in HTTP such as Header, Param or Body may take
// advantage of the request or response types (depending on whether they appear
// when describing the HTTP request or response). The properties of the header,
// parameter or body attributes inherit the properties of the attributes with
// the same names that appear in the request or response types. The functions
// may also define new attributes or override the existing request or response
// type attributes.
//
// HTTP must appear in API, Service or Method.
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
//            CanonicalMethod("add") // Method whose path is used to prefix
//                                   // the paths of child service.
//        })
//
//        Method("add", func() {
//            Description("Add two operands")
//            Payload(Operands)
//            Error(ErrBadRequest, ErrorResult)
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
	case *design.APIExpr:
		eval.Execute(fn, httpdesign.Root)
	case *design.ServiceExpr:
		res := httpdesign.Root.ServiceFor(actual)
		res.DSLFunc = fn
	case *design.MethodExpr:
		res := httpdesign.Root.ServiceFor(actual.Service)
		act := res.EndpointFor(actual.Name, actual)
		act.DSLFunc = fn
	default:
		eval.IncompatibleDSL()
	}
}

// Consumes adds a MIME type to the list of MIME types the API supports when
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
//    var _ = API("cellar", func() {
//        // ...
//        HTTP(func() {
//            Consumes("application/json", "application/xml")
//            // ...
//        })
//    })
//
func Consumes(args ...string) {
	switch def := eval.Current().(type) {
	case *httpdesign.RootExpr:
		def.Consumes = append(def.Consumes, args...)
	default:
		eval.IncompatibleDSL()
	}
}

// Produces adds a MIME type to the list of MIME types the API supports when
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
//    var _ = API("cellar", func() {
//        // ...
//        HTTP(func() {
//            Produces("application/json", "application/xml")
//            // ...
//        })
//    })
//
func Produces(args ...string) {
	switch def := eval.Current().(type) {
	case *httpdesign.RootExpr:
		def.Produces = append(def.Produces, args...)
	default:
		eval.IncompatibleDSL()
	}
}

// Path defines a service base path, i.e. a common path prefix to all the
// service methods. The path may define wildcards (see GET for a description of
// the wildcard syntax). The corresponding parameters must be described using
// Params. Multiple base paths may be defined for services.
func Path(val string) {
	switch def := eval.Current().(type) {
	case *httpdesign.ServiceExpr:
		def.Paths = append(def.Paths, val)
	default:
		eval.IncompatibleDSL()
	}
}

// Docs provides external documentation URLs for methods.
func Docs(fn func()) {
	docs := new(design.DocsExpr)
	if !eval.Execute(fn, docs) {
		return
	}

	switch expr := eval.Current().(type) {
	case *httpdesign.FileServerExpr:
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
func GET(path string) *httpdesign.RouteExpr {
	return route("GET", path)
}

// HEAD creates a route using the HEAD HTTP method. See GET.
func HEAD(path string) *httpdesign.RouteExpr {
	return route("HEAD", path)
}

// POST creates a route using the POST HTTP method. See GET.
func POST(path string) *httpdesign.RouteExpr {
	return route("POST", path)
}

// PUT creates a route using the PUT HTTP method. See GET.
func PUT(path string) *httpdesign.RouteExpr {
	return route("PUT", path)
}

// DELETE creates a route using the DELETE HTTP method. See GET.
func DELETE(path string) *httpdesign.RouteExpr {
	return route("DELETE", path)
}

// OPTIONS creates a route using the OPTIONS HTTP method. See GET.
func OPTIONS(path string) *httpdesign.RouteExpr {
	return route("OPTIONS", path)
}

// TRACE creates a route using the TRACE HTTP method. See GET.
func TRACE(path string) *httpdesign.RouteExpr {
	return route("TRACE", path)
}

// CONNECT creates a route using the CONNECT HTTP method. See GET.
func CONNECT(path string) *httpdesign.RouteExpr {
	return route("CONNECT", path)
}

// PATCH creates a route using the PATCH HTTP method. See GET.
func PATCH(path string) *httpdesign.RouteExpr {
	return route("PATCH", path)
}

func route(method, path string) *httpdesign.RouteExpr {
	r := &httpdesign.RouteExpr{Method: method, Path: path}
	a, ok := eval.Current().(*httpdesign.EndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return r
	}
	r.Endpoint = a
	a.Routes = append(a.Routes, r)
	return r
}

// Headers groups a set of Header expressions. It makes it possible to list
// required headers using the Required function.
//
// Headers must appear in Service HTTP expression to define request headers
// common to all the service methods. Headers may also appear in a method,
// response or error HTTP expression to define the HTTP endpoint request and
// response headers.
//
// Headers accepts one argument: Either a function listing the headers or a user
// type which must be an object and whose attributes define the headers.
//
// Example:
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
func Headers(args interface{}) {
	h := headers(eval.Current())
	if h == nil {
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
	h.Merge(design.NewMappedAttributeExpr(&design.AttributeExpr{Type: o}))
}

// Header describes a single HTTP header. The properties (description, type,
// validation etc.) of a header are inherited from the request or response type
// attribute with the same name by default.
//
// Header may appear in a service HTTP expression (to define request headers
// that apply to all the service endpoints), specific method HTTP expression (to
// define request headers), a Result expression (to define the response headers)
// or an Error expression (to define the error response headers). Header may
// also appear in a Headers expression.
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
	eval.Execute(func() { dsl.Attribute(name, args...) }, h.AttributeExpr)
	h.Remap()
}

// Params groups a set of Param expressions. It makes it possible to list
// required parameters using the Required function.
//
// Params must appear in a Service HTTP expression to define the service base
// path and query string parameters. Params may also appear in an method HTTP
// expression to define the HTTP endpoint path and query string parameters.
//
// Params accepts one argument: Either a function listing the parameters or a
// user type which must be an object and whose attributes define the parameters.
//
// Example:
//
//     var _ = Service("cellar", func() {
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
	if fn, ok := args.(func()); ok {
		eval.Execute(fn, p)
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
	p.Merge(design.NewMappedAttributeExpr(&design.AttributeExpr{Type: o}))
}

// Param describes a single HTTP request path or query string parameter.
//
// Param may appear in a service HTTP expression to define common parameters to
// all the service methods or a specific method HTTP expression. Param may also
// appear in a Params expression.
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
	p := params(eval.Current())
	if p == nil {
		eval.IncompatibleDSL()
		return
	}
	if name == "" {
		eval.ReportError("parameter name cannot be empty")
	}
	eval.Execute(func() { dsl.Attribute(name, args...) }, p.AttributeExpr)
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
	e, ok := eval.Current().(*httpdesign.EndpointExpr)
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
	e, ok := eval.Current().(*httpdesign.EndpointExpr)
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
		ref    *design.AttributeExpr
		setter func(*design.AttributeExpr)
		kind   string
	)

	// Figure out reference type and setter function
	switch e := eval.Current().(type) {
	case *httpdesign.EndpointExpr:
		ref = e.MethodExpr.Payload
		setter = func(att *design.AttributeExpr) {
			e.Body = att
		}
		kind = "Request"
	case *httpdesign.ErrorExpr:
		ref = e.ErrorExpr.AttributeExpr
		setter = func(att *design.AttributeExpr) {
			if e.Response == nil {
				e.Response = &httpdesign.HTTPResponseExpr{}
			}
			e.Response.Body = att
		}
		kind = "Error"
		if e.Name != "" {
			kind += " " + e.Name
		}
	case *httpdesign.HTTPResponseExpr:
		ref = e.Parent.(*httpdesign.EndpointExpr).MethodExpr.Result
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
		if ref.Find(a) == nil {
			eval.ReportError("%q is not found in result type", a)
			return
		}
		obj := design.AsObject(ref.Type)
		if obj == nil {
			eval.ReportError("%s type must be an object with an attribute with name %#v, got %T", kind, a, ref.Type)
			return
		}
		attr = design.DupAtt(obj.Attribute(a))
		if attr.Metadata == nil {
			attr.Metadata = design.MetadataExpr{"origin:attribute": []string{a}}
		} else {
			attr.Metadata["origin:attribute"] = []string{a}
		}
		if attr == nil {
			eval.ReportError("%s type does not have an attribute named %#v", kind, a)
			return
		}
	case design.UserType:
		attr = &design.AttributeExpr{Type: a}
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
	if attr.Metadata == nil {
		attr.Metadata = design.MetadataExpr{}
	}
	attr.Metadata["http:body"] = []string{}
	setter(attr)
}

// Parent sets the name of the parent service. The parent service canonical
// method path is used as prefix for all the service HTTP endpoint paths.
func Parent(name string) {
	r, ok := eval.Current().(*httpdesign.ServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	r.ParentName = name
}

// CanonicalMethod sets the name of the service canonical method. The canonical
// method endpoint path is used to prefix the paths to any child service
// endpoint. The default value is "show".
func CanonicalMethod(name string) {
	r, ok := eval.Current().(*httpdesign.ServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	r.CanonicalEndpointName = name
}

// headers returns the mapped attribute containing the headers for the given
// expression if it's either the root, a service or an endpoint - nil otherwise.
func headers(exp eval.Expression) *design.MappedAttributeExpr {
	switch e := exp.(type) {
	case *httpdesign.ServiceExpr:
		if e.Headers == nil {
			e.Headers = design.NewEmptyMappedAttributeExpr()
		}
		return e.Headers
	case *httpdesign.EndpointExpr:
		if e.Headers == nil {
			e.Headers = design.NewEmptyMappedAttributeExpr()
		}
		return e.Headers
	case *httpdesign.HTTPResponseExpr:
		if e.Headers == nil {
			e.Headers = design.NewEmptyMappedAttributeExpr()
		}
		return e.Headers
	default:
		return nil
	}
}

// params returns the mapped attribute containing the path and query params for
// the given expression if it's either the root, a service or an endpoint - nil
// otherwise.
func params(exp eval.Expression) *design.MappedAttributeExpr {
	switch e := exp.(type) {
	case *httpdesign.ServiceExpr:
		if e.Params == nil {
			e.Params = design.NewEmptyMappedAttributeExpr()
		}
		return e.Params
	case *httpdesign.EndpointExpr:
		if e.Params == nil {
			e.Params = design.NewEmptyMappedAttributeExpr()
		}
		return e.Params
	default:
		return nil
	}
}
