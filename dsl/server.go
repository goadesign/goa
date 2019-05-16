package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Server describes a single process listening for client requests. The DSL
// defines the set of services that the server exposes as well as host details.
// Not defining a server in a design has the same effect as defining a single
// server that exposes all of the services defined in the design in a single
// host listening on "locahost" and using port 80 for HTTP endpoints and 8080
// for GRPC endpoints.
//
// The Server expression is leveraged by the example generator to produce the
// service and client commands. It is also consumed by the OpenAPI specification
// generator. There is one specification generated per server. The first URI of
// the first host is used to set the OpenAPI v2 specification 'host' and
// 'basePath' values.
//
// Server must appear in a API expression.
//
// Server takes two arguments: the name of the server and the defining DSL.
//
// Example:
//
//    var _ = API("calc", func() {
//        Server("calcsvr", func() {
//            Description("calcsvr hosts the Calculator Service.")
//
//            // List the services hosted by this server.
//            Services("calc")
//
//            // List the Hosts and their transport URLs.
//            Host("production", func() {
//               Description("Production host.")
//               // URIs can be parameterized using {param} notation.
//               URI("https://{version}.goa.design/calc")
//               URI("grpcs://{version}.goa.design")
//
//               // Variable describes a URI variable.
//               Variable("version", String, "API version", func() {
//                   // URI parameters must have a default value and/or an
//                   // enum validation.
//                   Default("v1")
//               })
//           })
//
//           Host("development", func() {
//               Description("Development hosts.")
//               // Transport specific URLs, supported schemes are:
//               // 'http', 'https', 'grpc' and 'grpcs' with the respective default
//               // ports: 80, 443, 8080, 8443.
//               URI("http://localhost:80/calc")
//               URI("grpc://localhost:8080")
//           })
//       })
//   })
//
func Server(name string, fn ...func()) *expr.ServerExpr {
	if len(fn) > 1 {
		eval.ReportError("too many arguments given to Server")
	}
	api, ok := eval.Current().(*expr.APIExpr)
	if !ok {
		eval.IncompatibleDSL()
	}
	server := &expr.ServerExpr{Name: name}
	if len(fn) > 0 {
		eval.Execute(fn[0], server)
	}
	api.Servers = append(api.Servers, server)
	return server
}

// Services sets the list of services implemented by a server.
//
// Services must appear in a Server expression.
//
// Services takes one or more strings as argument corresponding to service
// names.
//
// Example:
//
//    var _ = Server("calcsvr", func() {
//        Services("calc", "adder")
//        Services("other") // Multiple calls to Services are OK
//    })
//
func Services(svcs ...string) {
	s, ok := eval.Current().(*expr.ServerExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	s.Services = append(s.Services, svcs...)
}

// Host defines a server host. A single server may define multiple hosts. Each
// host lists the set of URIs that identify it.
//
// The Host expression is leveraged by the example generator to produce the
// service and client commands. It is also consumed by the OpenAPI specification
// generator to initialize the server objects.
//
// Host must appear in a Server expression.
//
// Host takes two arguments: a name and a DSL function.
//
// Example:
//
//    Server("calcsvc", func() {
//        Host("development", func() {
//            URI("http://localhost:80/calc")
//            URI("grpc://localhost:8080")
//        })
//    })
//
func Host(name string, fn func()) {
	s, ok := eval.Current().(*expr.ServerExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	host := &expr.HostExpr{
		Name:       name,
		ServerName: s.Name,
		Variables:  &expr.AttributeExpr{Type: &expr.Object{}},
	}
	eval.Execute(fn, host)
	s.Hosts = append(s.Hosts, host)
}

// URI defines a server host URI. A single host may define multiple URIs. The
// supported schemes are 'http', 'https', 'grpc' and 'grpcs' where 'grpcs'
// indicates gRPC using client-side SSL/TLS. gRPC URIs may only define the
// authority component (in particular no path). URIs may be parameterized using
// the {param} notation. Note that the variables appearing in a URI must be
// provided when the service is initialized and in particular their values
// cannot defer between requests.
//
// The URI expression is leveraged by the example generator to produce the
// service and client commands. It is also consumed by the OpenAPI specification
// generator to initialize the server objects.
//
// URI must appear in a Host expression.
//
// URI takes one argument: a string representing the URI value.
//
// Example:
//
//    var _ = Server("calcsvc", func() {
//        Host("development", func() {
//            URI("http://localhost:80/{version}/calc")
//            URI("grpc://localhost:8080")
//        })
//    })
//
func URI(uri string) {
	h, ok := eval.Current().(*expr.HostExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	h.URIs = append(h.URIs, expr.URIExpr(uri))
}

// Variable defines a server host URI variable.
//
// The URI expression is leveraged by the example generator to produce the
// service and client commands. It is also consumed by the OpenAPI specification
// generator to initialize the server objects.
//
// Variable must appear in a Host expression.
//
// The Variable DSL is the same as the Attribute DSL with the following two
// restrictions:
//
//    1. The type used to define the variable must be a primitive.
//    2. The variable must have a default value and/or a enum validation.
//
// Example:
//
//    var _ = Server("calcsvr", func() {
//        Host("production", func() {
//            URI("https://{version}.goa.design/calc")
//            URI("grpcs://{version}.goa.design")
//
//            Variable("version", String, "API version", func() {
//                Enum("v1", "v2")
//            })
//        })
//    })
//
func Variable(name string, args ...interface{}) {
	if _, ok := eval.Current().(*expr.HostExpr); !ok {
		eval.IncompatibleDSL()
		return
	}
	Attribute(name, args...)
}
