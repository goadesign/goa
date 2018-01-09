package dsl

import (
	"goa.design/goa/design"
	"goa.design/goa/eval"
)

// Metadata is a set of key/value pairs that can be assigned to an object. Each
// value consists of a slice of strings so that multiple invocation of the
// Metadata function on the same target using the same key builds up the slice.
// Metadata may be set on attributes, result types, endpoints, responses,
// services and API definitions.
//
// While keys can have any value the following names are handled explicitly by
// goa when set on attributes.
//
// struct:error:name identifies the attribute of a result type used to select
// the returned error when multiple errors are defined on the same method.
// The value of the field corresponding to the attribute with the
// struct:error:name metadata is matched against the names of the method
// errors as defined in the design. This makes it possible to define distinct
// transport mappings for the various errors (for example to return different
// HTTP status codes). There must be one and exactly one attribute with the
// struct:error:name metadata defined on result types used to define error
// results.
//
//        var CustomErrorType = ResultType("application/vnd.goa.error", func() {
//                Attribute("message", String, "Error returned.", func() {
//                        Metadata("struct:error:name")
//                })
//                Attribute("occurred_at", DateTime, "Time error occurred.")
//        })
//
//        var _ = Service("MyService", func() {
//                Error("internal_error", CustomErrorType)
//                Error("bad_request", CustomErrorType)
//        })
//
// `struct:field:name`: overrides the Go struct field name generated by default
// by goa.  Applicable to attributes only.
//
//        Metadata("struct:field:name", "MyName")
//
// `struct:field:origin`: overrides the name of the value used to initialize an
// attribute value. For example if the attributes describes an HTTP header this
// metadata specifies the name of the header in case it's different from the name
// of the attribute. Applicable to attributes only.
//
//        Metadata("struct:field:origin", "X-API-Version")
//
// `struct:tag:xxx`: sets the struct field tag xxx on generated Go structs.
// Overrides tags that goa would otherwise set.  If the metadata value is a
// slice then the strings are joined with the space character as separator.
// Applicable to attributes only.
//
//        Metadata("struct:tag:json", "myName,omitempty")
//        Metadata("struct:tag:xml", "myName,attr")
//
// `swagger:generate`: specifies whether Swagger specification should be
// generated. Defaults to true.
// Applicable to services, endpoints and file servers.
//
//        Metadata("swagger:generate", "false")
//
// `swagger:summary`: sets the Swagger operation summary field.
// Applicable to endpoints.
//
//        Metadata("swagger:summary", "Short summary of what endpoint does")
//
// `swagger:example`: specifies whether to generate random example. Defaults to
// true.
// Applicable to API (for global setting) or individual attributes.
//
//        Metadata("swagger:example", "false")
//
// `swagger:tag:xxx`: sets the Swagger object field tag xxx.
// Applicable to services and endpoints.
//
//        Metadata("swagger:tag:Backend")
//        Metadata("swagger:tag:Backend:desc", "Description of Backend")
//        Metadata("swagger:tag:Backend:url", "http://example.com")
//        Metadata("swagger:tag:Backend:url:desc", "See more docs here")
//
// `swagger:extension:xxx`: sets the Swagger extensions xxx. It can have any
// valid JSON format value.
// Applicable to:
// api as within the info and tag object,
// service within the paths object,
// endpoint as within the path-item object,
// route as within the operation object,
// param as within the parameter object,
// response as within the response object
// and security as within the security-scheme object.
// See https://github.com/OAI/OpenAPI-Specification/blob/master/guidelines/EXTENSIONS.md.
//
//        Metadata("swagger:extension:x-api", `{"foo":"bar"}`)
//
// The special key names listed above may be used as follows:
//
//        var Account = Type("Account", func() {
//                Attribute("service", String, "Name of service", func() {
//                        // Override default name
//                        Metadata("struct:field:name", "ServiceName")
//                })
//        })
//
func Metadata(name string, value ...string) {
	appendMetadata := func(metadata design.MetadataExpr, name string, value ...string) design.MetadataExpr {
		if metadata == nil {
			metadata = make(map[string][]string)
		}
		metadata[name] = append(metadata[name], value...)
		return metadata
	}

	switch expr := eval.Current().(type) {
	case design.CompositeExpr:
		att := expr.Attribute()
		att.Metadata = appendMetadata(att.Metadata, name, value...)
	case *design.AttributeExpr:
		expr.Metadata = appendMetadata(expr.Metadata, name, value...)
	case *design.ResultTypeExpr:
		expr.Metadata = appendMetadata(expr.Metadata, name, value...)
	case *design.MethodExpr:
		expr.Metadata = appendMetadata(expr.Metadata, name, value...)
	case *design.ServiceExpr:
		expr.Metadata = appendMetadata(expr.Metadata, name, value...)
	case *design.APIExpr:
		expr.Metadata = appendMetadata(expr.Metadata, name, value...)
	default:
		eval.IncompatibleDSL()
	}
}
