package dsl

import (
	apidesign "github.com/goadesign/goa/design"
	apidsl "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/rest/design"
)

// Action describes a single endpoint  including the URL path, HTTP method, request parameters (via
// path wildcards or query strings) and payload (data structure describing the request HTTP body).
// Action also describe the possible responses including their HTTP status, headers and body via
// media types.
//
// An action belongs to a resource and "inherits" default values from the resource definition
// including the URL path prefix, default response media type and default payload attribute
// properties (inherited from the attribute with identical name in the resource default media type).
//
// Action may appear in Resource.
//
// Action accepts two arguments: the name of the action and its defining DSL.
//
// Example:
//
//    Action("Update", func() {
//        Description("Update account")
//        Docs(func() {
//            Description("Update docs")
//            URL("http//cellarapi.com/docs/actions/update")
//        })
//        Scheme("http")                       // "http", "https", "ws" or "wss"
//        Routing(
//            PUT("/:id"),                     // path relative to resource base path
//            PUT("//orgs/:org/accounts/:id"), // absolute path
//        )
//        Params(func() {                      // action parameters
//            Param("org", String)             // may correspond to path wildcards
//            Param("id", Integer)
//            Param("sort", func() {           // or URL query string values
//                Enum("asc", "desc")
//            })
//        })
//        Headers(func() {                     // relevant action headers
//            Header("Authorization", String)
//            Header("X-Account", Integer)
//            Required("Authorization", "X-Account")
//        })
//        Payload(UpdatePayload)                // HTTP request body type
//        // OptionalPayload(UpdatePayload)     // request body which may be omitted
//        Response(NoContent)                   // HTTP response, see Response
//        Response(NotFound)
//    })
//
func Action(name string, dsl func()) {
	if r, ok := eval.Current().(*design.ResourceExpr); ok {
		if r.Actions == nil {
			r.Actions = make(map[string]*design.ActionExpr)
		}
		action, ok := r.Actions[name]
		if !ok {
			action = &design.ActionExpr{
				Parent: r,
				Name:   name,
			}
		}
		if !eval.Execute(dsl, action) {
			return
		}
		r.Actions[name] = action
	} else {
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
	case *design.ActionExpr:
		expr.Docs = docs
	case *design.FileServerExpr:
		expr.Docs = docs
	default:
		apidsl.Docs(dsl)
	}
}

// Routing lists the action route. Each route is defined with a function named after the route HTTP
// method.
//
// The route function takes the path as argument. Route paths may use wildcards to identify action
// parameters by using the characters ':' or '*' to prefix the parameter name. The syntax `:param`
// matches a path segment (the characters in between slashes) while the syntax `*name` is a
// catch-all that matches the path until the end. See the httptreemux
// (https://godoc.org/github.com/dimfeld/httptreemux) package documentation for additional details.
//
// Example:
//
//     var _ = Resource("bottle", func() {
//         BasePath("/bottles")
//         DefaultMedia(BottleMedia)
//         Action("show", func() {
//             Routing(GET("/:id"))    // Endpoint path is "/bottles/:id"
//             Params(func()
//                 Param("id", Integer, "id of bottle", func() {
//                     Minimum(1)      // Define "id" parameter as strictly
//                 })                  // positive integer
//             })
//             Response(OK)
//         })
//         Action("update", func() {
//             Routing(
//                 PUT("/:id")         // Define action with multiple
//                 PATCH("/:id")       // routes.
//             )
//             Params(func()
//                 Param("id", Integer, "id of bottle", func() {
//                     Minimum(1)      // Define "id" parameter as strictly
//                 })                  // positive integer
//             })
//             Response(NoContent)
//         })
//     })
//
func Routing(routes ...*design.RouteExpr) {
	if a, ok := eval.Current().(*design.ActionExpr); ok {
		for _, r := range routes {
			r.Parent = a
			a.Routes = append(a.Routes, r)
		}
	} else {
		eval.IncompatibleDSL()
	}
}

// GET creates a route using the GET HTTP method. See Routing.
func GET(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "GET", Path: path}
}

// HEAD creates a route using the HEAD HTTP method. See Routing.
func HEAD(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "HEAD", Path: path}
}

// POST creates a route using the POST HTTP method. See Routing.
func POST(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "POST", Path: path}
}

// PUT creates a route using the PUT HTTP method. See Routing.
func PUT(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "PUT", Path: path}
}

// DELETE creates a route using the DELETE HTTP method. See Routing.
func DELETE(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "DELETE", Path: path}
}

// OPTIONS creates a route using the OPTIONS HTTP method. See Routing.
func OPTIONS(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "OPTIONS", Path: path}
}

// TRACE creates a route using the TRACE HTTP method. See Routing.
func TRACE(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "TRACE", Path: path}
}

// CONNECT creates a route using the CONNECT HTTP method. See Routing.
func CONNECT(path string) *design.RouteExpr {
	return &design.RouteExpr{Verb: "CONNECT", Path: path}
}

// PATCH creates a route using the PATCH HTTP method. See Routing.
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
