// Package dsl_test demonstrates the DSL defined in the rest dsl package.
//
// The tests focuses on the rest specific DSL. See package goa/dsl/_spec for the
// core goa DSL specification tests.
package dsl_test

import (
	"net/http"

	. "goa.design/goa.v2/design"
	. "goa.design/goa.v2/rest/dsl"
)

// The API expression defines the global API properties of tbe design. There can
// only be one such declaration in a given design package.
var _ = API("rest_dsl_spec", func() {

	// Error defines an error response common to all the API endpoints.
	// It accepts the name of the error as first argument and the type that
	// describes the response as second argument. If no type is provided
	// then the built-in ErrorMedia type is used. The expression below is
	// therefore equivalent to:
	//
	//     Error("api_error")
	//
	Error("api_error", ErrorMedia)

	// HTTP defines the API HTTP specific properties.
	// HTTP may appear multiple times to enable the use of traits.
	HTTP(func() {

		// Path defines the common path prefix to all API HTTP requests.
		Path("/path/{api_path_param}")

		// Params defines the API path and query string parameters.
		// The attributes defined in Params get merged into the request
		// types of all the API endpoints. The merge algorithm adds
		// new attributes to the request types if they don't already have
		// ones with the same names - overrides their properties (type,
		// description etc.) otherwise.
		Params(func() {
			// Param defines a single path or query string parameter.
			// The syntax of Param is the same as Attribute's.
			Param("api_path_param")
			// The name argument can optionally define a mapping
			// between the attribute and the query string key name
			// using the syntax "attribute name:query string key".
			Param("attribute_name:query-key")
			Required("api_path_param")
		})

		// Params also accepts a user type as argument. The user type
		// must be an object. Params may appear multiple times in which
		// case the union of all parameters defined in each Params DSL
		// is used to define the API path and query string parameters.
		Params(CommonParams)

		// Headers defines API headers common to all the API requests.
		// The attributes defined in Headers get merged into the request
		// types of all the API endpoints. The merge algorithm adds
		// new attributes to the request types if they don't already have
		// ones with the same names - overrides their properties (type,
		// description etc.) otherwise.
		Headers(func() {
			// Header defines a single header. The syntax of Header
			// is the same as Attribute's.
			// The name argument can optionally define a mapping
			// between the attribute and the HTTP header name
			// using the syntax "attribute name:header name".
			Header("name:Header-Name")
			Required("name")
		})

		// Headers also accepts a user type as argument. The user type
		// must be an object. Headers may appear multiple times in which
		// case the union of all headers defined in each Headers DSL
		// is used to define the API headers.
		Headers(CommonHeaders)

		// Error defines the HTTP response associated with the given
		// error. By default the response uses HTTP status code 400
		// ("Bad Request") and the error type attributes define the shape
		// of the body.
		//
		// Error accepts the name of the error as first argument and an
		// optional DSL used to describe the mapping of the error type
		// attributes to the HTTP response headers and body fields. If
		// the name of the error matches one of the errors defined in
		// the API DSL then
		Error("api_error", func() {
			// Code sets the HTTP response status code.
			Code(StatusUnauthorized)

			// Header defines a single header. The syntax
			// for Header is the same as Attribute's.
			// The name argument can optionally define a
			// mapping between the attribute and the HTTP
			// header name using the syntax "attribute
			// name:header name".
			//
			// If the error type defines an attribute with
			// the same name as the name of the Header
			// attribute then the header attribute inherits
			// all its properties (type, description,
			// validations, etc.) from it.
			Header("error_code:Error-Code")

			// Body defines the response body fields.
			// By default (when Body is absent) the error type
			// attributes not listed in the Header DSLs are used
			// to define the response body fields.
			Body(func() {
				// If the error type defines an attribute with
				// the same name then the Body attribute inherits
				// all its properties (type, description,
				// validations, etc.) from it.
				Attribute("id")
				Attribute("status")
				Attribute("code")
				Attribute("detail")
				Attribute("meta")
			})
		})
	})
})

// The Service expression defines a single service. There may be any number of
// Service declarations in one design.
var _ = Service("service", func() {

	// Server describes a single service host and may appear more than once.
	// URL must include the protocol and hostname and may include a port.
	// The hostname and port may use parameters to define possible
	// alternative values.
	// API Server definitions are overridden by the Service Server
	// definitions when present.
	Server("https://service.goa.design:443", func() {
		Description("Optional description")
	})

	// Error defines a common error response to all the service endpoints.
	// The DSL is identical as when used in an API expression.
	Error("service_error")

	// HTTP specific properties, see the API HTTP DSL for the descriptions
	// of the DSL functions.
	HTTP(func() {
		// HTTP request path prefix to all the service endpoints
		// (appended to API path prefix if there is one).
		Path("/service/{service_param}")

		// Service path prefix and query string parameters if any.
		// The Service HTTP Params definition works identically to the
		// API HTTP Params definition.
		Params(func() {
			Param("service_param")
		})

		// Service specific errors. Syntax and logic is identical to the
		// API level HTTP Error expressions.
		Error("service_error", http.StatusForbidden, func() {
			Header("name:Header-Name")
			Body(func() {
				Attribute("id")
				Attribute("status")
				Attribute("code")
				Attribute("detail")
				Attribute("meta")
			})
		})

		// Parent defines the parent service. The parent service
		// canonical endpoint path is used to prefix all the service
		// endpoint paths.
		// The argument given to Parent can be either the parent service
		// name or a value returned by Service.
		Parent("parent_service")

		// CanonicalEndpoint identifies the endpoint whose path is used
		// to prefix all the child service endpoint paths.
		CanonicalEndpoint("endpoint")

		// Files defines an endpoint that serve static assets. The files
		// being served are identified by path, either a file path for
		// service the file or a directory path for service files in
		// that directory. The HTTP path for requesting the files is
		// defined by the first argument of Files. The path may end with
		// a wildcard startign with * to capture the path suffix that
		// gets appended to the directory path.
		Files("/public/*filepath", "/www/data/", func() {
			Description("Optional description")
			Docs(func() {
				Description("Additional documentation")
				URL("https://goa.design")
			})
		})
	})

	// Endpoint describes a single endpoint. A service may define any number
	// of endpoints.
	Endpoint("endpoint", func() {
		// Request describes the request attributes. There can only be
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
		Request(RequestType)

		// Response describes the response attributes. The syntax is
		// identical to Request.
		Response(ResponseMediaType)

		// Error in an Endpoint expression defines endpoint specific
		// error responses, the syntax is identical as when used in a
		// API expression.
		Error("endpoint_error")

		// HTTP defines HTTP transport specific properties.
		HTTP(func() {

			// GET, POST, PUT etc. set the endpoint HTTP route. The
			// complete path is computed by appending the API prefix
			// path with the service prefix path with the endpoint
			// path.
			PUT("/endpoint_path/{endpoint_path_param}")

			// Param defines a single path or query string
			// parameter.
			// If the name of the parameter attribute matches
			// the name of one of the request type attribute
			// then it inherits all its properties
			// (description, type, validations etc.) from it.
			Param("endpoint_path_param")
			Param("endpoint_query_param")

			// Header defines a single header. The syntax for
			// Header is the same as Attribute's. The name argument
			// can optionally define a mapping between the attribute
			// and the HTTP header name using the syntax "attribute
			// name:header name".
			//
			// If the request type defines an attribute with the
			// same name as the name of the Header attribute then
			// the header attribute inherits all its properties
			// (type, description, validations, etc.) from it.
			Header("name:Header-Name")

			// Body defines the endpoint request body fields. This
			// function is optional and if not called the body
			// fields are defined by using all the request type
			// attributes not used by Param or Header. Body also
			// accepts the name of a request type attribute instead
			// of a DSL in which case the type used to represent
			// the request body is the type of the attribute (which
			// could be a primitive type, an array, a map or an
			// object).
			//
			//    Body("request_type_attribute")
			//
			Body(func() {
				// Attribute defines a single body field. If
				// the request type defines an attribute with the
				// same name then the Body attribute inherits
				// all its properties (type, description,
				// validations, etc.) from it.
				Attribute("request_type_attribute")
			})

			// Response defines a single HTTP response. There may be
			// more than one Response expression in a single
			// Endpoint HTTP expression to describe multiple possible
			// responses.
			Response(func() {
				// Code defines the HTTP response status code.
				// The default is 200 OK for responses whose type
				// is not Empty - 204 No Content otherwise.
				Code(StatusOK)
				// ContentType allows setting the value of the
				// response Content-Type header explicitly. By
				// default this header is set with the response
				// media type identifier if the response type is
				// a media type.
				ContentType("application/json")

				// Header defines a single header. The syntax
				// for Header is the same as Attribute's. The
				// name argument can optionally define a mapping
				// between the attribute and the HTTP header
				// name. If the response type defines an
				// attribute with the same name as the name of
				// the Header attribute then the header
				// attribute inherits all its properties (type,
				// description, validations, etc.) from it.
				Header("response_type_attribute")

				// Body defines the response body fields. This
				// function is optional and if not called the
				// body fields are defined by using all the
				// response type attributes not used by headers.
				// Body also accepts the name of a response type
				// attribute instead of a DSL in which case the
				// type used to represent the response body is
				// the type of the attribute (which could be a
				// primitive type, an array, a map or an
				// object).
				//
				//     Body("response_type_attribute")
				//
				Body(func() {
					// Attribute defines a single body
					// field. If the response type defines
					// an attribute with the same name then
					// the Body attribute inherits all its
					// properties (type, description,
					// validations, etc.) from it.
					Attribute("response_type_attribute")
					Required("response_type_attribute")
				})
			})

			// Error defines a endpoint specific error response. The
			// DSL is identical to API level HTTP Error expressions.
			Error("service_error")
		})
	})

	// Endpoint using the service request and response types default HTTP
	// mappings.
	Endpoint("another_endpoint", func() {
		Request(RequestType)
		Response(ResponseMediaType)

		HTTP(func() {

			// No Body function means the endpoint HTTP request body
			// is defined by the endpoint request type RequestType.
			PUT("/another")

			// No Response DSL means the response body shape and
			// content type is defined by the endpoint response type
			// ResponseMediaType and the status code is 200 OK.
		})
	})
})

// CommonParams is an object whose attributes define HTTP parameters common to
// all the API endpoints.
var CommonParams = Type("CommonParams", func() {
	Attribute("query")
	Attribute("other:O")
	Required("query")
})

// CommonHeaders is an object whose attributes define HTTP headers common to all
// the API endpoints.
var CommonHeaders = Type("CommonHeaders", func() {
	Attribute("Header-Name")
	Attribute("attribute:Other-Name")
	Required("attribute")
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
