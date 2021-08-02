package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Meta defines a set of key/value pairs that can be assigned to an object. Each
// value consists of a slice of strings so that multiple invocation of the Meta
// function on the same target using the same key builds up the slice.
//
// Meta may appear in attributes, result types, endpoints, responses, services
// and API definitions.
//
// While keys can have any value the following names have special meanings:
//
// - "type:generate:force" forces the code generation for the type it is defined
// on. By default goa only generates types that are used explicitly by the
// service methods. The value is a slice of strings that lists the names of the
// services for which to generate the struct. The struct is generated for all
// services if left empty.
//
//    package design
//
//    var _ = Service("service1", func() { ... })
//    var _ = Service("service2", func() { ... })
//
//    var Unused = Type("Unused", func() {
//        Attribute("name", String)
//        Meta("type:generate:force", "service1", "service2")
//    })
//
// - "struct:error:name" identifies the attribute of a result type used to
// select the returned error when multiple errors are defined on the same
// method. The value of the field corresponding to the attribute with the
// struct:error:name metadata is matched against the names of the method errors
// as defined in the design. This makes it possible to define distinct transport
// mappings for the various errors (for example to return different HTTP status
// codes). There must be one and exactly one attribute with the
// struct:error:name metadata defined on result types used to define error
// results.
//
//    var CustomErrorType = ResultType("application/vnd.goa.error", func() {
//        Attributes(func() {
//            Attribute("message", String, "Error returned.", func() {
//                Meta("struct:error:name")
//            })
//            Attribute("occurred_at", String, "Time error occurred.", func() {
//                Format(FormatDateTime)
//            })
//        })
//    })
//
//    var _ = Service("MyService", func() {
//        Error("internal_error", CustomErrorType)
//        Error("bad_request", CustomErrorType)
//    })
//
// - "struct:field:name" overrides the Go struct field name generated by default
// by goa. Applicable to attributes only.
//
//    var MyType = Type("MyType", func() {
//        Attribute("ssn", String, "User SSN", func() {
//            Meta("struct:field:name", "SSN")
//        })
//    })
//
// - "struct:field:type" overrides the Go struct field type specified in the design, with one caveat;
// if the type would have been a pointer (such as its not Required) the new type will also be a pointer.
// Applicable to attributes only. The import path of the type should be passed in as the second parameter, if needed.
// If the default imported package name conflicts with another, you can override that as well with the third parameter.
//
//    var MyType = Type("BigOleMessage", func() {
//        Attribute("type", String, "Type of big payload")
//        Attribute("bigPayload", String, "Don't parse it if you don't have to",func() {
//            Meta("struct:field:type","json.RawMessage","encoding/json")
//         })
//         Attribute("id", String, func() {
//             Meta("struct:field:type","bison.ObjectId", "github.com/globalsign/mgo/bson", "bison")
//         })
//    })
//
//
// - "struct:tag:xxx" sets a generated Go struct field tag and overrides tags
// that goa would otherwise set. If the metadata value is a slice then the
// strings are joined with the space character as separator. Applicable to
// attributes only.
//
//    var MyType = Type("MyType", func() {
//        Attribute("ssn", String, "User SSN", func() {
//            Meta("struct:tag:json", "SSN,omitempty")
//            Meta("struct:tag:xml", "SSN,omitempty")
//        })
//    })
//
// - "swagger:generate" specifies whether Swagger specification should be
// generated. Defaults to true. Applicable to services, methods and file
// servers.
//
//    var _ = Service("MyService", func() {
//        Meta("swagger:generate", "false")
//    })
//
// - "swagger:summary" sets the Swagger operation summary field. Applicable to
// methods.
//
//    var _ = Service("MyService", func() {
//        Method("MyMethod", func() {
//               Meta("swagger:summary", "Summary of MyMethod")
//        })
//    })
//
// - "swagger:example" specifies whether to generate random example. Defaults to
// true. Applicable to API (applies to all attributes) or individual attributes.
//
//    var _ = API("MyAPI", func() {
//        Meta("swagger:example", "false")
//    })
//
// - "swagger:tag:xxx" sets the Swagger object field tag xxx. Applicable to
// services and methods.
//
//    var _ = Service("MyService", func() {
//        Method("MyMethod", func() {
//            Meta("swagger:tag:Backend")
//            Meta("swagger:tag:Backend:desc", "Description of Backend")
//            Meta("swagger:tag:Backend:url", "http://example.com")
//            Meta("swagger:tag:Backend:url:desc", "See more docs here")
//            Meta("swagger:tag:Backend:extension:x-data", `{"foo":"bar"}`)
//        })
//    })
//
// - "swagger:extension:xxx" sets the Swagger extensions xxx. The value can be
// any valid JSON. Applicable to API (Swagger info and tag objects), Service
// (Swagger paths object), Method (Swagger path-item object), Route (Swagger
// operation object), Param (Swagger parameter object), Response (Swagger
// response object) and Security (Swagger security-scheme object). See
// https://github.com/OAI/OpenAPI-Specification/blob/master/guidelines/EXTENSIONS.md.
//
//    var _ = API("MyAPI", func() {
//        Meta("swagger:extension:x-api", `{"foo":"bar"}`)
//    })
//
func Meta(name string, value ...string) {
	appendMeta := func(meta expr.MetaExpr, name string, value ...string) expr.MetaExpr {
		if meta == nil {
			meta = make(map[string][]string)
		}
		meta[name] = append(meta[name], value...)
		return meta
	}

	switch e := eval.Current().(type) {
	case *expr.APIExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.AttributeExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.ResultTypeExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.MethodExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.ServiceExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.HTTPServiceExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.HTTPEndpointExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.RouteExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.HTTPFileServerExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.HTTPResponseExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case expr.CompositeExpr:
		att := e.Attribute()
		att.Meta = appendMeta(att.Meta, name, value...)
	default:
		eval.IncompatibleDSL()
	}
}
