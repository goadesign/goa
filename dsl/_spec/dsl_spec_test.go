// This test demonstrates all the DSL functions defined in the dsl package.
package dsl_test

import (
	. "goa.design/goa/v3/dsl"
	. "goa.design/goa/v3/expr"
)

// The API expression defines the global API properties of the design. There can
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
		// The attributes defined in Server get merged into the payload
		// types of all the API methods. The merge algorithm adds
		// new attributes to the payload types if they don't already have
		// ones with the same names - overrides their properties (type,
		// description etc.) otherwise.
		Param("param", String, "Optional description", func() {
			// Default value *must* be provided
			Default("default")
			// Optional list of possible values
			Enum("default", "other")
		})
	})

	// Meta whose effects depend on the generators.
	Meta("meta", "value", "other value")
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

	// Error defines a common error to all the service methods.
	Error("name_of_error_1")
	// ErrorResult is a built-in result type used by default for errors.
	Error("name_of_error_2", ErrorResult, "Optional description of error")
	// Error attributes can be described using a result type.
	Error("name_of_error_3", AErrorResultType)
	// Error attributes can be described using a user type.
	Error("name_of_error_4", AErrorType)
	// Error attributes can be described inline.
	Error("name_of_error_5", func() {
		Description("Optional description")
		Attribute("message")
		Required("message")
	})

	// Method describes a single method. A service may define any number
	// of methods.
	Method("method", func() {
		// Method description for code comments and docs
		Description("Optional description")

		// Docs allows linking to external documentation.
		Docs(func() {
			Description("Optional description")
			URL("https://goa.design")
		})

		// Payload describes the payload attributes. There can only be
		// one Payload expression per Method expression.
		// Payload attributes can be described inline.
		//
		//     Payload(func() {
		//         Attribute("name")
		//         Required("name")
		//     })
		//
		// Payload attributes can be described using a user type.
		//
		//     Payload(PayloadType)
		//
		// Additionally Payload can add to the list of required
		// attributes.
		//
		//     Payload(PayloadType, func() {
		//         Required("name")
		//     })
		//
		Payload(PayloadType, func() {
			Required("name")
		})

		// Result describes the result attributes. There can only be
		// one Result expression per Method expression.
		// Result attributes can be described inline.
		//
		//     Result(func() {
		//         Attribute("name")
		//         Required("name")
		//     })
		//
		// Result attributes can be described using a user or result
		// type.
		//
		//     Result(ResultType)
		//
		// Additionally Result can add to the list of required
		// attributes.
		Result(ResType, func() {
			Required("name")
		})

		// Error in an Method expression defines method specific
		// errors, the syntax is identical as when used in a Service
		// expression.
		Error("method_specific_error")

		// Meta expression. Effect depends on generators.
		// Meta takes the name of the metadta as first argument and
		// one or more values.
		Meta("name", "some value", "some other value")
	})

	// Method with inline payload and result object types
	Method("inline-object", func() {
		Payload(func() {
			Description("Optional description")
			Attribute("required")
			Attribute("optional")
			Required("required")
		})
		Result(func() {
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

// AErrorResultType is a simple result type definition.
var AErrorResultType = ResultType("application/vnd.goa.design.error", func() {
	TypeName("AErrorResult")
	Attributes(func() {
		Attribute("msg")
	})
	View("default", func() {
		Attribute("msg")
	})
})

// PayloadType is the type that describes the payload attributes.
var PayloadType = Type("Payload", func() {
	Description("Optional description")
	Attribute("required")
	Attribute("name")
	Required("required")
})

// ResType is the result type that describes the result shape.
var ResType = ResultType("application/vnd.goa.result", func() {
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
