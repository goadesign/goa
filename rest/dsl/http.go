package dsl

import (
	"strings"

	goadesign "goa.design/goa.v2/design"
	goadsl "goa.design/goa.v2/dsl"
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
//                Header("requestID:X-RequestID")  // Use "requestID" attribute
//                                                 // of Operands to define shape
//                                                 // of X-RequestID header
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
		res := design.Root.ResourceFor(actual)
		eval.Execute(dsl, res)
	case *goadesign.EndpointExpr:
		res := design.Root.ResourceFor(actual.Service)
		act := res.Action(actual.Name)
		eval.Execute(dsl, act)
	default:
		eval.IncompatibleDSL()
	}
}

// Scheme sets the API URL schemes.
func Scheme(vals ...design.APIScheme) {
	switch def := eval.Current().(type) {
	case *goadesign.APIExpr:
		design.Root.Schemes = append(design.Root.Schemes, vals...)
	case *design.ResourceExpr:
		def.Schemes = append(def.Schemes, vals...)
	case *design.ActionExpr:
		def.Schemes = append(def.Schemes, vals...)
	default:
		eval.IncompatibleDSL()
	}
}

// Path defines the API base path, i.e. the common path prefix to all the API
// actions. The path may define wildcards (see GET for a description of the
// wildcard syntax). The corresponding parameters must be described using Params.
func Path(val string) {
	switch def := eval.Current().(type) {
	case *goadesign.APIExpr:
		design.Root.Path = val
	case *design.ResourceExpr:
		def.Path = val
		if !strings.HasPrefix(val, "//") {
			awcs := design.ExtractWildcards(design.Root.Path)
			wcs := design.ExtractWildcards(val)
			for _, awc := range awcs {
				for _, wc := range wcs {
					if awc == wc {
						eval.ReportError(`duplicate wildcard "%s" in API and resource base paths`, wc)
					}
				}
			}
		}
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

// Headers defines a mapping of attribute names to HTTP header names. Depending
// on where Headers is used it defines how to map the incoming request HTTP
// headers to the request type attributes or how to map the response type
// attributes to the outgoing HTTP response headers.
//
// Headers defines the mapping of request headers to request type attributes when
// used in the service DSL (in which case the mapping is to the service default
// type) or in a specific endpoint HTTP DSL. Headers define the mapping of the
// response type attributes to HTTP response headers when used in the Response
// DSL either at the service level (in which case the mapping is from the
// default response type) or at the specific endpoint level.
//
// Example:
//
//    Headers(func() {
//        Header("Authorization") // Map attribute Authorization or header with
//                                // same name.
//        Header("account:X-Account") // Map attribute account to header
//                                    // X-Account
//        Required("Authorization") // Required headers
//    })
//
func Headers(dsl func()) {
	switch def := eval.Current().(type) {
	case *design.ActionExpr:
		headers := design.NewAttributeMap(def, def.Resource.DefaultType)
		if eval.Execute(dsl, headers) {
			def.Headers = headers
		}

	case *design.ResourceExpr:
		headers := design.NewAttributeMap(def, def.DefaultType)
		if eval.Execute(dsl, headers) {
			def.Headers = headers
		}

	case *design.HTTPResponseExpr:
		var h *design.AttributeMapExpr
		switch actual := def.Parent.(type) {
		case *design.ResourceExpr:
			h = design.NewAttributeMap(def, actual.DefaultType)
		case *design.ActionExpr:
			h = design.NewAttributeMap(def, actual.Resource.DefaultType)
		default:
			eval.ReportError("invalid use of Response or ResponseTemplate")
		}
		if eval.Execute(dsl, h) {
			def.Headers = h
		}

	default:
		eval.IncompatibleDSL()
	}
}

// Header describes a single HTTP header.
//
// Header may appear inside Headers.
//
// Header accepts a single argument which defines both the name of the header and
// the name of the corresponding request or response type attribute. The argument
// is of the form "name of attribute:name of header". If both names are identical
// then the form "name" (with no :) my be used instead.
//
// Example:
//
//     var _ = Endpoint("
func Header(name string) {
	headers, ok := eval.Current().(*design.AttributeMapExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if err := headers.Alias(name); err != nil {
		eval.ReportError(err.Error())
	}
}

// Params describe the endpoint parameters, either path parameters identified
// via wildcards or query string parameters. Path parameters are described with
// the Param function while query string parameters are described with the Query
// function. Both the Param and Query functions appear inside the Params DSL.
//
// Each Param and Query expression must map one of the endpoint request type
// attributes. The syntax for defining parameters is:
//
//      Param("ATTRIBUTE[:PARAMETER]")
//
// And the one for query strings is:
//
//      Query("ATTRIBUTE[:PARAMETER]")
//
// where ATTRIBUTE is the name of request type attribute and PARAMETER the name
// of the HTTP path or query string parameter. The parameter name is optional
// and only required when the name differ from the name of the attribute.
//
// Params may appear inside the endpoint HTTP DSL to define the action
// parameters, serivce HTTP DSL to define common parameters to all the service
// endpoints or API HTTP DSL to define common parameters to all the API actions.
//
// Example:
//
//    var ShowRequest = Type("ShowRequest", func() {
//        Attribute("version", String, "Endpoint version", func() {
//            Enum("1.0", "2.0")
//        })
//    })
//
//    var _ = Service("Bottle", func() {
//        DefaultType(BottleMedia)   // Service "Bottle" uses "BottleMedia" as
//        Endpoint("show", func() {  // default response type.
//            Request(ShowRequest)
//            Routing(GET("/:name")) // Action show uses parameter "name"
//            Params(func() {   // Parameter inherits type, description and
//                Param("name") // validation from BottleMedia "name" attribute
//            })
//        })
//    })
//
func Params(dsl func()) {
	var params *design.AttributeMapExpr
	switch def := eval.Current().(type) {
	case *design.ActionExpr:
		params = design.NewAttributeMap(def, def.EndpointExpr.Request)
	case *design.ResourceExpr:
		params = design.NewAttributeMap(def, def.DefaultType)
	case *goadesign.APIExpr:
		params = design.NewAttributeMap(def, nil)
	default:
		eval.IncompatibleDSL()
	}
	if !eval.Execute(dsl, params) {
		return
	}
	switch def := eval.Current().(type) {
	case *design.ActionExpr:
		def.Params = params
	case *design.ResourceExpr:
		def.Params = params
	case *goadesign.APIExpr:
		design.Root.Params = params
	}
}

// Param describes a single HTTP request path or query string parameter.
//
// Param may appear inside Params.
//
// Param accepts a single argument which defines both the name of the parameter
// and the name of the corresponding request or response type attribute. The
// argument is of the form "name_of_header:name_of_attribute". If both names are
// identical then the form "common_name" my be used instead.
func Param(name string) {
	params, ok := eval.Current().(*design.AttributeMapExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if err := params.Alias(name); err != nil {
		eval.ReportError(err.Error())
	}
}
