package apidsl

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
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
	if r, ok := resourceExpr(); ok {
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
		if !dslengine.Execute(dsl, action) {
			return
		}
		r.Actions[name] = action
	}
}

// Files defines a endpoint that serves static assets. The logic for what to do when the
// filename points to a file vs. a directory is the same as the standard http package ServeFile
// function. The path may end with a wildcard that matches the rest of the URL (e.g. *filepath). If
// it does the matching path is appended to filename to form the full file path, so:
//
//     Files("/index.html", "/www/data/index.html")
//
// returns the content of the file "/www/data/index.html" when requests are sent to "/index.html"
// and:
//
//    Files("/assets/*filepath", "/www/data/assets")
//
// returns the content of the file "/www/data/assets/x/y/z" when requests are sent to
// "/assets/x/y/z".
//
// Files may appear in Resource.
//
// Files accepts 2 arguments and an optional DSL. The first argument is the request path which may
// use a wildcard starting with *. The second argument is the path on disk to the files being
// served. The file path may be absolute or relative to the current path of the process.  The DSL
// allows setting a description and documentation.
//
// Example:
//
//    var _ = Resource("bottle", func() {
//        Files("/index.html", "/www/data/index.html", func() {
//            Description("Serve home page")
//            Docs(func() {
//                Description("Download docs")
//                URL("http//cellarapi.com/docs/actions/download")
//            })
//        })
//    })
//
func Files(path, filename string, dsls ...func()) {
	if r, ok := resourceExpr(); ok {
		server := &design.FileServerExpr{
			Parent:      r,
			RequestPath: path,
			FilePath:    filename,
		}
		if len(dsls) > 0 {
			if !dslengine.Execute(dsls[0], server) {
				return
			}
		}
		r.FileServers = append(r.FileServers, server)
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
	if a, ok := actionExpr(); ok {
		for _, r := range routes {
			r.Parent = a
			a.Routes = append(a.Routes, r)
		}
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

// Headers implements the DSL for describing HTTP headers. The DSL syntax is identical to the one
// of Attribute. Here is an example defining a couple of headers with validations:
//
//    Headers(func() {
//        Header("Authorization")
//        Header("X-Account", Integer, func() {
//            Minimum(1)
//        })
//        Required("Authorization")
//    })
//
// Headers can be used inside Action to define the action request headers, Response to define the
// response headers or Resource to define common request headers to all the resource actions.
func Headers(params ...interface{}) {
	if len(params) == 0 {
		dslengine.ReportError("missing parameter")
		return
	}
	dsl, ok := params[0].(func())
	if ok {
		switch def := dslengine.CurrentExpr().(type) {
		case *design.ActionExpr:
			headers := newAttribute(def.Parent.MediaType)
			if dslengine.Execute(dsl, headers) {
				def.Headers = def.Headers.Merge(headers)
			}

		case *design.ResourceExpr:
			headers := newAttribute(def.MediaType)
			if dslengine.Execute(dsl, headers) {
				def.Headers = def.Headers.Merge(headers)
			}

		case *design.ResponseExpr:
			var h *design.FieldExpr
			switch actual := def.Parent.(type) {
			case *design.ResourceExpr:
				h = newAttribute(actual.MediaType)
			case *design.ActionExpr:
				h = newAttribute(actual.Parent.MediaType)
			case nil: // API ResponseTemplate
				h = &design.FieldExpr{}
			default:
				dslengine.ReportError("invalid use of Response or ResponseTemplate")
			}
			if dslengine.Execute(dsl, h) {
				def.Headers = def.Headers.Merge(h)
			}

		default:
			dslengine.IncompatibleDSL()
		}
	} else if cors, ok := corsExpr(); ok {
		vals := make([]string, len(params))
		for i, p := range params {
			if v, ok := p.(string); ok {
				vals[i] = v
			} else {
				dslengine.ReportError("invalid parameter at position %d: must be a string", i)
				return
			}
		}
		cors.Headers = vals
	} else {
		dslengine.IncompatibleDSL()
	}
}

// Params describe the action parameters, either path parameters identified via wildcards or query
// string parameters if there is no corresponding path parameter. Each parameter is described via
// the Param function which uses the same DSL as the Attribute DSL. Here is an example:
//
//    Params(func() {
//        Param("id", Integer)        // A path parameter defined using e.g. GET("/:id")
//        Param("sort", String, func() {    // A query string parameter
//            Enum("asc", "desc")
//        })
//    })
//
// Params can be used inside Action to define the action parameters, Resource to define common
// parameters to all the resource actions or API to define common parameters to all the API actions.
//
// If Params is used inside Resource or Action then the resource base media type attributes provide
// default values for all the properties of params with identical names. For example:
//
//     var BottleMedia = MediaType("application/vnd.bottle", func() {
//         Attributes(func() {
//             Attribute("name", String, "The name of the bottle", func() {
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
//         DefaultMedia(BottleMedia) // Resource "Bottle" uses "BottleMedia" as default
//         Action("show", func() {   // media type.
//             Routing(GET("/:name"))
//             Params(func() {
//                 Param("name") // inherits type, description and validation from
//                               // BottleMedia "name" attribute
//             })
//         })
//     })
//
func Params(dsl func()) {
	var params *design.FieldExpr
	switch def := dslengine.CurrentExpr().(type) {
	case *design.ActionExpr:
		params = newAttribute(def.Parent.MediaType)
	case *design.ResourceExpr:
		params = newAttribute(def.MediaType)
	case *design.APIExpr:
		params = new(design.FieldExpr)
	default:
		dslengine.IncompatibleDSL()
	}
	params.Type = make(design.Object)
	if !dslengine.Execute(dsl, params) {
		return
	}
	switch def := dslengine.CurrentExpr().(type) {
	case *design.ActionExpr:
		def.Params = def.Params.Merge(params) // Useful for traits
	case *design.ResourceExpr:
		def.Params = def.Params.Merge(params) // Useful for traits
	case *design.APIExpr:
		def.Params = def.Params.Merge(params) // Useful for traits
	}
}
