// Package dsl_test demonstrates the DSL defined in the rest dsl package.
//
// The tests focuses on the rest specific DSL. See package
// goadesign/design/dsl/_spec for the core goa DSL specification tests.
package dsl_test

import (
	"net/http"

	. "goa.design/goa.v2/rest/design"
	. "goa.design/goa.v2/rest/design/dsl"
)

// The API expression defines the global API properties of tbe design. There can
// only be one such declaration in a given design package.
var _ = API("rest_dsl_spec", func() {

	// Error defines an error response common to all the API endpoints.
	// It accepts the name of the error as first argument and the type that
	// describes the response as second argument. If no type is provided
	// then the built-in ErrorMedia type is used.
	Error("api_error", ErrorMedia)

	// HTTP defines the API HTTP specific properties
	HTTP(func() {

		// Scheme defines the default HTTP scheme, the argument must be
		// one of "http", "https", "ws" or "wss".
		Scheme("https")

		// Path defines the common path prefix to all API HTTP requests.
		Path("/path/{api_path_param}")

		// Params define the API path and query string parameters
		Params(func() {
			// Param defines a single path or query string
			// parameter. The Param arguments and DSL are identical
			// to Attribute.
			Param("api_path_param", String)
			Param("api_query_param", String)
		})

		// Error defines the HTTP response associated with the given
		// error. By default the response uses HTTP status code 400 and
		// the error type attributes to define the contents of the body.
		// Error accepts the name of the error as first argument, an
		// HTTP status code as second argument and an optional DSL used
		// to describe the mapping of the response type attribute to the
		// HTTP response headers and body.
		Error("api_error", http.StatusUnauthorized, func() {

			// Headers list the error response headers.
			Headers(func() {
				// Header defines a single header. If the name
				// of the header differs from the name of the
				// error type attribute then the mapping is
				// defined using the syntax
				// "Header-Name:attribute_name".
				Header("Error-Code:code")
			})

			// Body defines the error type attributes used to render
			// the response body. By default the attributes not
			// listed in the Headers DSL are used.
			Body(func() {
				// Attribute specifies the name of a request
				// type attribute. The other properties are
				// inherited from the request type.
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
	// DefaultType is an optional expression that can be used to define the
	// default response type for the service endpoints. The attributes of
	// the default type also define the default properties for attributes of
	// the same name in request types.
	DefaultType(ResponseMediaType)

	// Error defines a common error response to all the service endpoints.
	// The DSL is identical as when used in an API expression.
	Error("service_error")

	// HTTP specific properties, see the API HTTP DSL for the descriptions
	// of the DSL functions.
	HTTP(func() {
		// Override default API scheme for all the service endpoints.
		Scheme("http")

		// HTTP request path prefix to all the service endpoints
		// (appended to API path prefix if there is one).
		Path("/service/{service_param}")

		// Service path prefix parameters if any.
		Params(func() {
			Param("service_param", String)
		})

		// Service specific errors.
		Error("service_error", http.StatusForbidden, func() {
			Headers(func() {
				Header("Error-Code:code")
			})
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
	})

	// Endpoint describes a single endpoint. A service may define any number
	// of endpoints.
	Endpoint("endpoint", func() {
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
		Request(RequestType)

		Response(func() {
			// Inherits type, description, default value, example
			// and validations from ResponseMediaType "name"
			// attribute.
			Attribute("name")
		})

		// Error in an Endpoint expression defines endpoint specific
		// error responses, the syntax is identical as when used in a
		// API expression.
		Error("endpoint_error")

		// HTTP defines HTTP transport specific properties.
		HTTP(func() {

			// GET, POST, PUT etc. set the endpoint HTTP route. The
			// complete path is computed by appending the API prefix
			// path with the resource prefix path with the endpoint
			// path.
			PUT("/endpoint_path/{endpoint_path_param}")

			// Params defines the path and query string parameters.
			Params(func() {
				// Param defines a single path or query string
				// parameter. The Param arguments and DSL are
				// identical to Attribute.
				Param("endpoint_path_param")
				Param("endpoint_query_param")
			})

			// Headers list headers that are relevant to the
			// endpoint handler.
			Headers(func() {
				// Header defines a single header. The arguments
				// and DSL is identical to Attribute. The names
				// of the header must match the name of a
				// request type attribute or must be of the form
				// "Header name:Attribute name".
				Header("Header-Name:name")
			})

			// Body defines the endpoint request body fields. This
			// function is optional and if not called the body
			// fields are defined by using all the request type
			// attributes not used by params or headers. Body also
			// accepts the name of a request type attribute instead
			// of a DSL in which case the type used to represent
			// the request body is the type of the attribute (which
			// could be a primitive type, an array, a map or an
			// object).
			//
			//    Body("request_type_attribute")
			//
			Body(func() {
				Attribute("request_type_attribute")
			})

			// Response defines a single HTTP response. There may be
			// more than one Response expression in a single
			// Endpoint expression to describe multiple possible
			// responses. Response accepts the HTTP status code as
			// first argument, a type as optional second argument
			// and an optional DSL as last argument.
			Response(OK, func() {
				// ContentType allows setting the value of the
				// response Content-Type header explicitely. By
				// default this header is set with the response
				// media type identifier if the response type is
				// a media type.
				ContentType("application/json")

				// Headers list the response type attributes
				// mapped to the response headers. The mapping
				// uses the attribute  name as header name
				// unsill the attribute name follows the format
				// "Header-Name:attribute name".
				Headers(func() {
					// Header defines a single header, the
					// argument must match one of the
					// response type attribute names.
					Header("response_type_attribute")
					Required("response_type_attribute")
				})

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
					Attribute("response_type_attribute")
					Required("response_type_attribute")
				})
			})

			// If the endpoint response type is Empty then the
			// response body must be empty, the Body function is
			// optional. If the endpoint type is not Empty then the
			// response Body function may use Empty to signifiy an
			// empty response body.
			Response(NoContent, func() {
				Body(Empty)
			})

			// Error defines a endpoint specific error response. The
			// DSL is identical to API level HTTP Error expressions.
			Error("service_error")
		})
	})

	// Endpoint that does not define a response type and therefore inherits
	// from the service default type.
	Endpoint("another_endpoint", func() {
		Request(RequestType)

		HTTP(func() {

			// No Body function means the endpoint HTTP request body
			// is defined by the endpoint request type RequestType.
			PUT("/another")

			// No type argument or DSL means the response body shape
			// is defined by the endpoint response type, here the
			// service default type.
			Response(OK)
		})
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
