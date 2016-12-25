// This test demonstrates all the DSL functions defined in the dsl package.
package dsl_test

import . "goa.design/goa.v2/design"
import . "goa.design/goa.v2/dsl"

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
	// URL must include the protocol and hostname and may include a port.
	// The hostname and port may use parameters to define possible
	// alternative values.
	Server("https://{param}.goa.design:443", func() {
		Description("Optional description")

		// Param describes a single parameter used in the server URL.
		//
		// The syntax for Param is the same as Attribute's. The Server
		// Param declarations must include a default value.
		//
		// The attributes defined in Server get merged into the request
		// types of all the API endpoints. The merge algorithm adds
		// new attributes to the request types if they don't already have
		// ones with the same names - overrides their properties (type,
		// description etc.) otherwise.
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

	// Server definitions that appear in the Service DSL override all the API
	// level definitions.
	Server("https://service.goa.design:443", func() {
		Description("Service specific server description")
	})

	// Docs allows linking to external documentation.
	Docs(func() {
		Description("Optional description")
		URL("https://goa.design")
	})

	// Error defines a common error response to all the service endpoints.
	Error("name_of_error_1")
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
		Attribute("message")
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

		// Request describes the request attributes. There can only be
		// one Request expression per Endpoint expression.
		// Request attributes can be described inline.
		//
		//     Request(func() {
		//         Attribute("name")
		//         Required("name")
		//     })
		//
		// Request attributes can be described using a user type. The
		// user type must be an object.
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

		// Response describes the response attributes. There can only be
		// one Response expression per Endpoint expression.
		// Response attributes can be described inline.
		//
		//     Response(func() {
		//         Attribute("name")
		//         Required("name")
		//     })
		//
		// Response attributes can be described using a user or media
		// type. If using a user type it must be an object.
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

	// Endpoint with inline request and response object types
	Endpoint("inline-object", func() {
		Request(func() {
			Description("Optional description")
			Attribute("required")
			Attribute("optional")
			Required("required")
		})
		Response(func() {
			Description("Optional description")
			Attribute("required")
			Attribute("optional")
			Required("required")
		})
	})
})

// ServiceDefaultType is a simple type definition.
var ServiceDefaultType = Type("ServiceDefaultType", func() {
	Attribute("value")
})

// AErrorType is a simple type definition.
var AErrorType = Type("AErrorType", func() {
	Attribute("msg")
})

// AErrorMediaType is a simple media type definition.
var AErrorMediaType = MediaType("application/vnd.goa.design.error", func() {
	TypeName("AErrorMedia")
	Attributes(func() {
		Attribute("msg")
	})
	View("default", func() {
		Attribute("msg")
	})
})

// RequestType is the type that describes the request parameters.
var RequestType = Type("Request", func() {
	Description("Optional description")
	Attribute("required")
	Attribute("name")
	Required("required")
})

// ResponseMediaType is the media type that describes the response shape.
var ResponseMediaType = MediaType("application/vnd.goa.response", func() {
	Description("Optional description")
	Attributes(func() {
		Attribute("required")
		Attribute("name")
		Required("required")
	})
	View("default", func() {
		Attribute("required")
		Attribute("name")
	})
})
