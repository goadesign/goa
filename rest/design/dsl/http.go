package dsl

import (
	apidesign "goa.design/goa.v2/design"
	apidsl "goa.design/goa.v2/design/dsl"
	"goa.design/goa.v2/eval"
	"goa.design/goa.v2/rest/design"
)

// HTTP defines HTTP transport specific properties either on a service or on a single endpoint. The
// function maps the request and response types to HTTP properties such as parameters (via path
// wildcards or query strings), request and response headers and bodies and response status code.
// HTTP also defines HTTP specific properties such as the endpoint URLs and HTTP methods.
//
// HTTP may appear in a Service or an Endpoint expression.
//
// HTTP accepts a single argument which the DSL function.
//
// Example:
//
//    var _ = Service("Manager", func() {
//        DefaultType(Account)
//        Error(ErrAuthFailure)
//
//        HTTP(func() {
//            BasePath("/accounts")               // Prefix to all service HTTP request paths
//            Error(ErrAuthFailure, Unauthorized) // Use HTTP status code 401 for ErrAuthFailure
//                                                // error responses.
//                                                // ErrUnauthorized error responses
//            Scheme("http")                      // "http", "https", "ws" or "wss"
//        })
//
//        Endpoint("update", func() {
//            Description("Change account name")
//            Request(UpdateAccount)
//            Response(Empty)
//            Error(ErrNotFound)
//            Error(ErrBadRequest, ErrorResponse)
//
//            HTTP(func() {
//                PUT("/{accountID}")    // "accountID" attribute of UpdateAccount
//                Query("req:requestID") // Use "requestID" attribute to define "req" query string
//                Scheme("https")        // Override default service scheme
//                Body(func() {
//                    Attribute("name")  // "name" attribute of UpdateAccount
//                    Required("name")
//                })
//                Header("X-RequestID:requestID") // Use "requestID" attribute of UpdateAccount to
//                                                // define shape of "X-RequestID" header
//                Response(NoContent)             // Use HTTP status code 204 on success
//                Error(ErrNotFound, NotFound)    // Use HTTP status code 404 for ErrNotFound
//                Error(ErrBadRequest, BadRequest, ErrorMedia) // Use status code 400 for
//                                                // ErrBadRequest, also use ErrorMedia media type
//                                                // to describe response body.
//            })
//
//        })
//    })
func HTTP(dsl func()) {
	switch actual := eval.Current().(type) {
	case *apidesign.ServiceExpr:
		res := NewResourceExpr(actual, dsl)
		design.Root.Resources = append(design.Root.Resources, res)
	case *apidesign.EndpointExpr:
		act := NewActionExpr(actual, dsl)
		res = design.ResourceFor(actual.Service)
		res.Actions = append(res.Actions, act)
	default:
		eval.IncompatibleDSL()
	}
}

// Docs provides external documentation pointers for actions.
func Docs(dsl func()) {
	docs := new(apidesign.DocsExpr)
	if !eval.Execute(dsl, docs) {
		return
	}

	switch expr := eval.Current().(type) {
	case *design.FileServerExpr:
		expr.Docs = docs
	default:
		apidsl.Docs(dsl)
	}
}

// GET defines a route using the GET HTTP method. The route may use wildcards to define path
// parameters. Wildcards start with '{' or with '{*' and end with '}'. A wildcard that starts with
// '{' matches a section of the path (the value in between two slashes). A wildcard that starts with
// '{*' matches the rest of the path. Such wildcards must terminate the path.
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

// Headers define relevant action HTTP headers. The DSL syntax is identical to the one of Attribute.
// Here is an example defining a couple of headers with validations:
//
// Headers can be used inside Action to define the action request headers, Response to define the
// response headers or Resource to define common request headers to all the resource actions.
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
func Headers(params ...interface{}) {
	if len(params) == 0 {
		eval.ReportError("missing parameter")
		return
	}
	dsl, ok := params[0].(func())
	if ok {
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
			var h *apidesign.AttributeExpr
			switch actual := def.Parent.(type) {
			case *design.ResourceExpr:
				h = newAttribute(actual.MediaType)
			case *design.ActionExpr:
				h = newAttribute(actual.Parent.MediaType)
			case nil: // API ResponseTemplate
				h = &apidesign.AttributeExpr{}
			default:
				eval.ReportError("invalid use of Response or ResponseTemplate")
			}
			if eval.Execute(dsl, h) {
				def.Headers = def.Headers.Merge(h)
			}

		default:
			eval.IncompatibleDSL()
		}
	} else {
		eval.IncompatibleDSL()
	}
}

// Params describe the action parameters, either path parameters identified via wildcards or query
// string parameters if there is no corresponding path parameter. Each parameter is described
// with the Param function which appears in the Params DSL.
//
// Params may appear inside Action to define the action parameters, Resource to define common
// parameters to all the resource actions or API to define common parameters to all the API actions.
//
// Example:
//
//    var _ = API("cellar", func() { // Define API "cellar"
//        BasePath("/api/:version")  // Base path uses parameter defined by :version
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
	var params *apidesign.AttributeExpr
	switch def := eval.Current().(type) {
	case *design.ActionExpr:
		params = newAttribute(def.Parent.MediaType)
	case *design.ResourceExpr:
		params = newAttribute(def.MediaType)
	case *apidesign.APIExpr:
		params = new(apidesign.AttributeExpr)
	default:
		eval.IncompatibleDSL()
	}
	params.Type = make(apidesign.Object)
	if !eval.Execute(dsl, params) {
		return
	}
	switch def := eval.Current().(type) {
	case *design.ActionExpr:
		def.Params = def.Params.Merge(params) // Useful for traits
	case *design.ResourceExpr:
		def.Params = def.Params.Merge(params) // Useful for traits
	case *apidesign.APIExpr:
		design.Root.Params = design.Root.Params.Merge(params) // Useful for traits
	}
}
