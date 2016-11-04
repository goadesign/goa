// This test demonstrates all the DSL functions defined in the dsl package.
package dsl_test

import . "github.com/goadesign/goa/design"
import . "github.com/goadesign/goa/design/dsl"

// The API expression defines the global API properties of tbe design. There can
// only be one such declaration in a given design package.
var _ = API("language_spec", func() {
	Title("Language Spec")
	Description("A test that covers all the supported DSL functions")

	// API version
	Version("1.0")

	// API support information.
	Contact(func() {
		Name("goa team")
		Email("admin@goa.design")
		URL("http://goa.design")
	})

	// API Licensing information
	License(func() {
		Name("MIT")
	})

	// Docs allows linking to external documentation.
	Docs(func() {
		Description("goa guide")
		URL("http://goa.design/getting-started.html")
	})

	// Default host used in generated OpenAPI spec and client.
	Host("goa.design")

	// Metadata whose effects depend on the generators.
	Metadata("metadata", "some value", "some other value")
})

// The Service expression defines a single service. There may be any number of
// Service declarations in one design.
var _ = Service("service", func() {
	Description("A Service expression")

	// Docs allows linking to external documentation.
	Docs(func() {
		Description("External documentation link")
		URL("https://goa.design")
	})

	// DefaultType is an optional expression that can be used to define the
	// default response type for the service endpoints. The attributes of
	// the default type also define the default properties for attributes of
	// the same name in request types.
	DefaultType(AType)

	// Error defines a common error response to all the service endpoints.
	Error("goa_error", ErrorMedia)
	Error("custom_error", AErrorMediaType)
	Error("non_media_error", AErrorType)

	// Endpoint describes a single endpoint. A service may define any number
	// of endpoints.
	Endpoint("endpoint", func() {
		Description("A service endpoint")

		// Docs allows linking to external documentation.
		Docs(func() {
			Description("External documentation link")
			URL("https://goa.design")
		})

		// Endpoint request type.
		Request(RequestType, func() {
			// Required fields - on top of fields marked as required
			// in Type expression.
			Required("attribute_b")
		})

		// Endpoint response type, may be a media type.
		Response(ResponseType)

		// Endpoint specific errors
		Error("ep_goa_error", ErrorMedia)
		Error("ep_custom_error", AErrorMediaType)
		Error("ep_non_media_error", AErrorType)

		// Metadata expression. Effect depends on generators.
		Metadata("metadata", "some value", "some other value")
	})

	Endpoint("another", func() {
		Description("Endpoint with inline request and response types")
		Request(ArrayOf(String))
		Response(Bytes)
	})
})

// AType is a simple type definition.
var AType = Type("AType", func() {
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

// AnotherType is a simple type definition.
var AnotherType = Type("AnotherType", func() {
	Attribute("value", Int32)
})

// RequestType is the type that describes the request parameters.
var RequestType = Type("request", func() {
	Description("Request parameters")
	Attribute("attribute_a", String)
	Attribute("attribute_b", String)
	Required("attribute_a")
})

// ResponseType is the media type that describes the response shape.
var ResponseType = MediaType("application/vnd.goa.design.response", func() {
	Description("Response media type")
	Attributes(func() {
		Attribute("attribute_a", String)
		Attribute("attribute_b", String)
		Required("attribute_a")
	})
	View("default", func() {
		Attribute("attribute_a")
	})
})
