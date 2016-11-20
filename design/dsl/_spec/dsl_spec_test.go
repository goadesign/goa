// This test demonstrates all the DSL functions defined in the dsl package.
package dsl_test

import . "goa.design/goa.v2/design"
import . "goa.design/goa.v2/design/dsl"

// The API expression defines the global API properties of tbe design. There can
// only be one such declaration in a given design package.
var _ = API("dsl_spec", func() {
	// API title for docs
	Title("Optional API title")

	// API description for docs
	Description("Optional API description")

	// API version
	Version("1.0")

	// API support information.
	Contact(func() {
		Name("contact name")
		Email("contact@goa.design")
		URL("https://goa.design")
	})

	// API Licensing information
	License(func() {
		Name("License name")
		URL("https://goa.design/license")
	})

	// Docs allows linking to external documentation.
	Docs(func() {
		Description("Optional description")
		URL("https://goa.design/getting-started")
	})

	// Server describes a single API host and may appear more than once.
	// URL must include protocol and hostname, may include port and base
	// path (HTTP only).  Any component except the protocol may use
	// parameters.
	Server("https://{param}.goa.design:443/basePath", func() {
		Description("Optional description")
		// Param describes a single parameter
		Param("param", String, "Optional description", func() {
			// Default value *must* be provided
			Default("default")
			// Optional list of possible values
			Enum("default", "other")
		})
	})

	// Metadata whose effects depend on the generators.
	Metadata("metadata", "value", "other value")
})

// The Service expression defines a single service. There may be any number of
// Service declarations in one design.
var _ = Service("service", func() {
	// Service description for code comments and docs
	Description("Optional service description")

	// Docs allows linking to external documentation.
	Docs(func() {
		Description("Optional description")
		URL("https://goa.design")
	})

	// DefaultType is an optional expression that can be used to define the
	// default response type for the service endpoints. The attributes of
	// the default type also define the default properties for attributes of
	// the same name in request types.
	DefaultType(ServiceDefaultType)

	// Error defines a common error response to all the service endpoints.
	Error("name_of_error")
	// ErrorMedia is a built-in media type used by default for error
	// responses.
	Error("name_of_error_2", ErrorMedia, "Optional description of error")
	// Error response attributes can be described using a media type
	Error("name_of_error_3", AErrorMediaType)
	// Error response attributes can be described using a user type
	Error("name_of_error_4", AErrorType)
	// Error response attributes can be described inline
	Error("name_of_error_5", func() {
		Description("Optional description")
		Attribute("message", String)
		Required("message")
	})

	// Endpoint describes a single endpoint. A service may define any number
	// of endpoints.
	Endpoint("endpoint", func() {
		// Endpoint description for code comments and docs
		Description("Optional description")

		// Docs allows linking to external documentation.
		Docs(func() {
			Description("Optional description")
			URL("https://goa.design")
		})

		// Request describes the request attributes. There must be only
		// one Request expression per Endpoint expression.
		// Request attributes can be described inline.
		//
		//     Request(func() {
		//         Attribute("name", String)
		//         Required("name")
		//     })
		//
		// Request attributes can be described using a user type.
		//
		//     Request(RequestType)
		//
		// Additionally Request can add to the list of required
		// attributes.
		//
		//     Request(RequestType, func() {
		//         Required("name")
		//     })
		//
		Request(RequestType, func() {
			Required("name")
		})

		// Response describes the response attributes. There must be
		// only one Response expression per Endpoint expression.
		// Response attributes can be described inline.
		//
		//     Response(func() {
		//         Attribute("name", String)
		//         Required("name")
		//     })
		//
		// Response attributes can be described using a user or media
		// type.
		//
		//     Response(ResponseMediaType)
		//
		// Additionally Response can add to the list of required
		// attributes.
		Response(ResponseMediaType, func() {
			Required("name")
		})

		// Error in an Endpoint expression defines endpoint specific
		// error responses, the syntax is identical as when used in a
		// Service expression.
		Error("endpoint_specific_error")

		// Metadata expression. Effect depends on generators.
		// Metadata takes the name of the metadta as first argument and
		// one or more values.
		Metadata("name", "some value", "some other value")
	})

	// Endpoint using the service default type as request attributes and
	// reusing the default type attributes to define the response
	// attributes.
	Endpoint("default-type", func() {
		Response(func() {
			// Inherits type, description, default value, example
			// and validations from ServiceDefaultType "value"
			// attribute.
			Attribute("value")
		})
	})

	// Endpoint with inline request and response primitive types
	Endpoint("inline-primitive", func() {
		Request(String)
		Response(String)
	})

	// Endpoint with request and response array types
	Endpoint("inline-array", func() {
		Request(ArrayOf(String))
		Response(ArrayOf(String))
	})

	// Endpoint with request and response map types
	Endpoint("inline-map", func() {
		Request(MapOf(String, String))
		Response(MapOf(String, String))
	})

	// Endpoint with inline request and response object types
	Endpoint("inline-object", func() {
		Request(func() {
			Description("Optional description")
			Attribute("required", String)
			Attribute("optional", String)
			Required("required")
		})
		Response(func() {
			Description("Optional description")
			Attribute("required", String)
			Attribute("optional", String)
			Required("required")
		})
	})
})

// ServiceDefaultType is a simple type definition.
var ServiceDefaultType = Type("ServiceDefaultType", func() {
	Attribute("value", String)
})

// AErrorType is a simple type definition.
var AErrorType = Type("AErrorType", func() {
	Attribute("msg", String)
})

// AErrorMediaType is a simple media type definition.
var AErrorMediaType = MediaType("application/vnd.goa.design.error", func() {
	TypeName("AErrorMedia")
	Attributes(func() {
		Attribute("msg", String)
	})
	View("default", func() {
		Attribute("msg")
	})
})

// RequestType is the type that describes the request parameters.
var RequestType = Type("Request", func() {
	Description("Optional description")
	Attribute("required", String)
	Attribute("name", String)
	Required("required")
})

// ResponseMediaType is the media type that describes the response shape.
var ResponseMediaType = MediaType("application/vnd.goa.response", func() {
	Description("Optional description")
	Attributes(func() {
		Attribute("required", String)
		Attribute("name", String)
		Required("required")
	})
	View("default", func() {
		Attribute("required")
		Attribute("name")
	})
})
