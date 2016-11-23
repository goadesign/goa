package dsl

import (
	goadesign "goa.design/goa.v2/design"
	goadsl "goa.design/goa.v2/design/dsl"
	"goa.design/goa.v2/eval"
	"goa.design/goa.v2/rest/design"
)

// HTTP defines HTTP transport specific properties either on a service or on a
// single endpoint. The function maps the request and response types to HTTP
// properties such as parameters (via path wildcards or query strings), request
// and response headers and bodies and response status code.  HTTP also defines
// HTTP specific properties such as the endpoint URLs and HTTP methods.
//
// HTTP may appear in a Service or an Endpoint expression.
//
// HTTP accepts a single argument which the DSL function.
//
// Example:
//
//    var _ = Service("calculator", func() {
//        DefaultType(ResultMediaType)
//        Error(ErrAuthFailure)
//
//        HTTP(func() {
//            Path("/calc")      // Prefix to all request paths
//            Error(ErrAuthFailure, http.StatusUnauthorized) // Define
//                               // ErrAuthFailure HTTP response status code.
//            Scheme("http")     // HTTP scheme
//            Parent("account")  // Parent resource, used to prefix request
//                               // paths.
//            CanonicalEndpoint("add") // Endpoint whose path is used to prefix
//                                     // the paths of child resources.
//        })
//
//        Endpoint("add", func() {
//            Description("Add two operands")
//            Request(Operands)
//            Error(ErrBadRequest, ErrorMedia)
//
//            HTTP(func() {
//                GET("/add/{left}/{right}") // Define HTTP route. The "left"
//                                         // and "right" parameter properties
//                                         // are inherited from the
//                                         // corresponding Operands attributes.
//                Params(func() {            // Define endpoint path and query
//                                           // string parameters.
//                    Param("req:requestID") // Use "requestID" attribute to
//                                           // define "req" query string
//                })
//                Scheme("https")        // Override default service scheme
//                Header("X-RequestID:requestID") // Use "requestID" attribute
//                                                // of Operands to define shape
//                                                // of X-RequestID header
//                Response(http.StatusNoContent)   // Use status 204 on success
//                Error(ErrBadRequest, BadRequest) // Use status code 400 for
//                                                 // ErrBadRequest responses
//            })
//
//        })
//    })
func HTTP(dsl func()) {
	switch actual := eval.Current().(type) {
	case *goadesign.ServiceExpr:
		res := ResourceFor(actual)
		eval.Execute(dsl, res)
	case *goadesign.EndpointExpr:
		act := ActionFor(actual)
		eval.Execute(dsl, act)
		res = design.ResourceFor(actual.Service)
		res.Actions = append(res.Actions, act)
	default:
		eval.IncompatibleDSL()
	}
}

// Docs provides external documentation pointers for actions.
func Docs(dsl func()) {
	docs := new(goadesign.DocsExpr)
	if !eval.Execute(dsl, docs) {
		return
	}

	switch expr := eval.Current().(type) {
	case *design.FileServerExpr:
		expr.Docs = docs
	default:
		goadsl.Docs(dsl)
	}
}

// GET defines a route using the GET HTTP method. The route may use wildcards to
// define path parameters. Wildcards start with '{' or with '{*' and end with
// '}'.
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
//             Request(GetAccount)
//             Response(Account)
//             HTTP(func() {
//                 GET("/{accountID}")
//                 GET("/{*accountPath}")
//             })
//         })
//     })
func GET(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "GET", Path: path}
}

// HEAD creates a route using the HEAD HTTP method. See GET.
func HEAD(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "HEAD", Path: path}
}

// POST creates a route using the POST HTTP method. See GET.
func POST(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "POST", Path: path}
}

// PUT creates a route using the PUT HTTP method. See GET.
func PUT(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "PUT", Path: path}
}

// DELETE creates a route using the DELETE HTTP method. See GET.
func DELETE(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "DELETE", Path: path}
}

// OPTIONS creates a route using the OPTIONS HTTP method. See GET.
func OPTIONS(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "OPTIONS", Path: path}
}

// TRACE creates a route using the TRACE HTTP method. See GET.
func TRACE(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "TRACE", Path: path}
}

// CONNECT creates a route using the CONNECT HTTP method. See GET.
func CONNECT(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "CONNECT", Path: path}
}

// PATCH creates a route using the PATCH HTTP method. See GET.
func PATCH(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "PATCH", Path: path}
}

// Headers define relevant action HTTP headers. The DSL syntax is identical to
// the one of Attributes.
//
// Headers can be used inside a service HTTP DSL to define common request
// headers to all the service endpoints, a endpoint DSL to define endpoint
// request specific headers or Response to define response headers.
//
// Example:
//
//    Headers(func() {
//        Header("Authorization")
//        Header("X-Account", Integer, func() {
//            Minimum(1)
//        })
//        Required("Authorization")
//    })
//
func Headers(dsl func()) {
	switch def := eval.Current().(type) {
	case *design.ActionExpr:
		headers := newAttribute(def.Parent.MediaType)
		if eval.Execute(dsl, headers) {
			def.Headers = def.Headers.Merge(headers)
		}

	case *design.ResourceExpr:
		headers := newAttribute(def.MediaType)
		if eval.Execute(dsl, headers) {
			def.Headers = def.Headers.Merge(headers)
		}

	case *design.ResponseExpr:
		var h *goadesign.AttributeExpr
		switch actual := def.Parent.(type) {
		case *design.ResourceExpr:
			h = newAttribute(actual.MediaType)
		case *design.ActionExpr:
			h = newAttribute(actual.Parent.MediaType)
		case nil: // API ResponseTemplate
			h = &goadesign.AttributeExpr{}
		default:
			eval.ReportError("invalid use of Response or ResponseTemplate")
		}
		if eval.Execute(dsl, h) {
			def.Headers = def.Headers.Merge(h)
		}

	default:
		eval.IncompatibleDSL()
	}
}

// Params describe the endpoint parameters, either path parameters identified
// via wildcards or query string parameters if there is no corresponding path
// parameter. Each parameter is described with the Param function which appears
// in the Params DSL.
//
// Params may appear inside the endpoint HTTP DSL to define the action
// parameters, serivce HTTP DSL to define common parameters to all the service
// endpoints or API HTTP DSL to define common parameters to all the API actions.
//
// Example:
//
//    var _ = API("cellar", func() { // Define API "cellar"
//        Path("/api/{version}")     // Base path uses parameter defined by :version
//        Params(func() {            // Define parameters
//            Param("version", String, func() { // Define version parameter
//                Enum("v1", "v2")              // Syntax is identical to Attribute's
//            })
//        })
//    })
//
// If Params is used inside Resource or Action then the resource base media type attributes provide
// default values for all the properties of params with identical names. For example:
//
//     var BottleMedia = MediaType("application/vnd.bottle", func() {
//         Attributes(func() {
//             Attribute("name", String, "Name of bottle", func() {
//                 MinLength(2) // BottleMedia has one attribute "name" which is a
//                              // string that must be at least 2 characters long.
//             })
//         })
//         View("default", func() {
//             Attribute("name")
//         })
//     })
//
//     var _ = Resource("Bottle", func() {
//         DefaultMedia(BottleMedia)  // Resource "Bottle" uses "BottleMedia" as
//         Action("show", func() {    // default media type.
//             Routing(GET("/:name")) // Action show uses parameter "name"
//             Params(func() {   // Parameter "name" inherits type, description
//                 Param("name") // and validation from BottleMedia "name" attribute
//             })
//         })
//     })
//
func Params(dsl func()) {
	var params *goadesign.AttributeExpr
	switch def := eval.Current().(type) {
	case *design.ActionExpr:
		params = newAttribute(def.Parent.MediaType)
	case *design.ResourceExpr:
		params = newAttribute(def.MediaType)
	case *goadesign.APIExpr:
		params = new(goadesign.AttributeExpr)
	default:
		eval.IncompatibleDSL()
	}
	params.Type = make(goadesign.Object)
	if !eval.Execute(dsl, params) {
		return
	}
	switch def := eval.Current().(type) {
	case *design.ActionExpr:
		def.Params = def.Params.Merge(params) // Useful for traits
	case *design.ResourceExpr:
		def.Params = def.Params.Merge(params) // Useful for traits
	case *goadesign.APIExpr:
		design.Root.Params = design.Root.Params.Merge(params) // Useful for traits
	}
}
