// Package dsl_test demonstrates the DSL defined in the DSL DSL package.
//
// The tests focuses on the HTTP specific DSL. See package goa/dsl/_spec for the
// core goa DSL specification tests.
package dsl_test

import (
	. "goa.design/goa/http/design"
	. "goa.design/goa/http/dsl"
)

// The API expression defines the global API properties of tbe design. There can
// only be one such declaration in a given design package.
var _ = API("http_dsl_spec", func() {

	// Error defines an error response common to all the API methods.
	// It accepts the name of the error as first argument and the type that
	// describes the response as second argument. If no type is provided
	// then the built-in ErrorResult type is used. The expression below is
	// therefore equivalent to:
	//
	//     Error("api_error")
	//
	Error("api_error", ErrorResult)

	// HTTP defines the API HTTP specific properties.
	HTTP(func() {

		// Consumes lists the mime types corresponding to the encodings
		// supported by the API in requests. goa knows how to generate
		// the decoding code for the following mime types:
		// "application/json", "application/xml" and "application/gob".
		// The decoding code for other mime types must be written and
		// provided to the generated handler constructors.
		Consumes("application/json", "application/xml")

		// Produces lists the mime types corresponding to the encodings
		// used by the API to encode responses. goa knows how to
		// generate the encoding code for the following mime types:
		// "application/json", "application/xml" and "application/gob".
		// The encoding code for other mime types must be written and
		// provided to the generated handler constructors.
		Produces("application/json", "application/xml")

		// Path defines the common path prefix to all API HTTP requests.
		Path("/path/{api_path_param}")

		// Params groups path and query string parameter expressions.
		// The attributes defined in Params get merged into the request
		// types of all the API methods. The merge algorithm adds
		// new attributes to the request types if they don't already have
		// one with the same name or overrides the existing attribute
		// properties (type, description etc.) if they do.
		Params(func() {
			// Param defines a single path or query string parameter.
			// The arguments of Param are the same as the Attribute
			// function.
			Param("api_path_param")
			// The name argument can optionally define a mapping
			// between the attribute and the query string key name
			// using the syntax "attribute name:query string key".
			Param("attribute_name:query-key")
			Required("api_path_param")
		})

		// Params also accepts a user type as argument. The user type
		// must be an object. Params may appear multiple times in which
		// case the union of all parameters defined in each Params
		// expression is used to define the API path and query string
		// parameters.
		Params(CommonParams)

		// Headers defines API headers common to all the API requests.
		// The attributes defined in Headers get merged into the request
		// types of all the API methods. The merge algorithm adds
		// new attributes to the request types if they don't already have
		// one with the same name or overrides the existing attribute
		// properties (type, description etc.) if they do.
		Headers(func() {
			// Header defines a single header. The arguments of
			// Header are the same as the Attribute function.
			// The name argument can optionally define a mapping
			// between the attribute and the HTTP header name
			// using the syntax "attribute name:header name".
			Header("name:Header-Name")
			Required("name")
		})

		// Headers also accepts a user type as argument. The user type
		// must be an object. Headers may appear multiple times in which
		// case the union of all headers defined in each Headers
		// expression is used to define the API headers.
		Headers(CommonHeaders)

		// Response defines the HTTP response associated with the given
		// error. By default the response uses HTTP status code 400
		// ("Bad Request") and the error type attributes define the shape
		// of the body.
		//
		// Response the name of the error as first argument, an
		// HTTP status code as second argument and an optional function
		// used to describe the mapping of the error type attributes to
		// the HTTP response headers and body fields. The name of the
		// error must correspond to one of the errors defined in the API
		// expression.
		Response("api_error", StatusUnauthorized, func() {

			// Headers list the error response headers.
			Headers(func() {
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
				Header("error_code:Code")
			})

			// Body defines the response body fields.
			// By default (when Body is absent) the error type
			// attributes not listed in the Headers DSL are used
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

	// Error defines a common error response to all the service methods.
	// It accepts the name of the error as first argument and the type that
	// describes the response as second argument. If no type is provided
	// then the built-in ErrorResult type is used. The expression below is
	// therefore equivalent to:
	//
	//     Error("service_error")
	//
	Error("service_error", ErrorResult)

	// HTTP specific properties
	HTTP(func() {
		// HTTP request path prefix to all the service methods
		Path("/service/{service_param}")

		// Params groups path and query string parameter expressions.
		// The attributes defined in Params must correspond to
		// attributes defined in the method paylods.
		Params(func() {
			// Param defines a single path or query string
			// parameter. The arguments of Param are the same as the
			// Attribute function.
			Param("service_param")
			// The name argument can optionally define a mapping
			// between the attribute and the query string key name
			// using the syntax "attribute name:query string key".
			Param("attribute_name:query-key")
			Required("service_path_param")
		})

		// Params also accepts a user type as argument. The user type
		// must be an object. Params may appear multiple times in which
		// case the union of all parameters defined in each Params
		// expression is used to define the service path and query
		// string parameters.
		Params(CommonParams)

		// Headers defines HTTP headers common to all the service
		// endpoints. The attributes defined in Headers must correspond
		// to attributes defined in the method payloads.
		Headers(func() {
			// Header defines a single header. The arguments of
			// Header are the same as the Attribute function.
			// The name argument can optionally define a mapping
			// between the attribute and the HTTP header name
			// using the syntax "attribute name:header name".
			Header("name:Header-Name")
			Required("name")
		})

		// Headers also accepts a user type as argument. The user type
		// must be an object. Headers may appear multiple times in which
		// case the union of all headers defined in each Headers
		// expression is used to define the common service endpoint
		// headers.
		Headers(CommonHeaders)

		// Consumes lists the mime types corresponding to the encodings
		// supported by the API in requests. goa knows how to generate
		// the decoding code for the following mime types:
		// "application/json", "application/xml" and "application/gob".
		// The decoding code for other mime types must be written and
		// provided to the generated handler constructors.
		Consumes("application/json", "application/xml")

		// Produces lists the mime types corresponding to the encodings
		// used by the API to encode responses. goa knows how to
		// generate the encoding code for the following mime types:
		// "application/json", "application/xml" and "application/gob".
		// The encoding code for other mime types must be written and
		// provided to the generated handler constructors.
		Produces("application/json", "application/xml")

		// Response defines the HTTP response associated with the given
		// error. By default the response uses HTTP status code 400
		// ("Bad Request") and the error type attributes define the
		// shape of the body.
		//
		// Response the name of the error as first argument, an HTTP
		// status code as second argument and an optional function used
		// to describe the mapping of the error type attributes to the
		// HTTP response headers and body fields. The name of the error
		// must correspond to one of the errors defined in the service
		// expression.
		Response("service_error", StatusForbidden, func() {
			// Headers list the error response headers.
			Headers(func() {
				// Header defines a single header. The name of
				// the header attribute must match the name of a
				// error type attribute. The name argument can
				// optionally define a mapping between the
				// attribute and the HTTP header name using the
				// syntax "attribute name:header name".
				Header("name:Header-Name")
			})

			// Body defines the response body fields.
			// By default (when Body is absent) the error type
			// attributes not listed in the Headers DSL are used
			// to define the response body fields.
			Body(func() {
				// The names of the attributes must match the
				// names of attributes of the method result
				// type.
				Attribute("id")
				Attribute("status")
				Attribute("code")
				Attribute("detail")
				Attribute("meta")
			})
		})

		// Parent defines the parent service. The parent service
		// canonical method path is used to prefix all the service
		// method paths.
		// The argument given to Parent can be either the parent service
		// name or a value returned by Service.
		Parent("parent_service")

		// CanonicalMethod identifies the method whose path is used
		// to prefix all the child service method paths.
		CanonicalMethod("method")

		// Files defines an method that serve static assets. The files
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

	// Method describes a single method. A service may define any number
	// of methods.
	Method("method", func() {
		// Payload describes the request attributes. There can only be
		// one Payload expression per Method expression.
		// Payload attributes can be described inline.
		//
		//     Payload(func() {
		//         Attribute("name", String)
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
		Payload(PayloadType)

		// Result describes the result attributes. The syntax is
		// identical to Payload with the exception that it makes it
		// possible to list the views used by the response when the first
		// argument is a result type. Listing no view has the same effect
		// as listing all views in this case.
		Result(ResponseResultType, "view")

		// Error in an Method expression defines method specific
		// error responses, the syntax is identical as when used in a
		// API expression.
		Error("method_error")

		// HTTP defines HTTP transport specific properties.
		HTTP(func() {

			// GET, POST, PUT etc. set the method HTTP route. The
			// complete path is computed by appending the API prefix
			// path with the service prefix path with the method
			// path.
			PUT("/method_path/{method_path_param}")

			// Params defines the path and query string parameters.
			Params(func() {
				// Param defines a single path or query string
				// parameter.
				// If the name of the parameter attribute matches
				// the name of one of the request type attribute
				// then it inherits all its properties
				// (description, type, validations etc.) from it.
				Param("method_path_param")
				Param("method_query_param")
			})

			// Headers list request headers that are relevant to the
			// method handler.
			Headers(func() {
				// Header defines a single header. The syntax
				// for Header is the same as Attribute's.
				// The name argument can optionally define a
				// mapping between the attribute and the HTTP
				// header name using the syntax "attribute
				// name:header name".
				//
				// If the request type defines an attribute with
				// the same name as the name of the Header
				// attribute then the header attribute inherits
				// all its properties (type, description,
				// validations, etc.) from it.
				Header("name:Header-Name")
			})

			// Body defines the method request body fields. This
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
				// Attribute defines a single body field. If
				// the request type defines an attribute with the
				// same name then the Body attribute inherits
				// all its properties (type, description,
				// validations, etc.) from it.
				Attribute("request_type_attribute")
			})

			// Response defines a single HTTP response. There may be
			// more than one Response expression in a single
			// Method HTTP expression to describe multiple possible
			// responses. Response accepts the HTTP status code as
			// first argument and an optional DSL as last argument.
			Response(StatusOK, func() {
				// ContentType allows setting the value of the
				// response Content-Type header explicitely. By
				// default this header is set with the response
				// result type identifier if the response type is
				// a result type.
				ContentType("application/json")

				// Headers list the response type attributes
				// mapped to the response headers. The mapping
				// uses the attribute  name as header name
				// unsill the attribute name follows the format
				// "Header-Name:attribute name".
				Headers(func() {
					// Header defines a single header. The
					// syntax for Header is the same as
					// Attribute's. The name argument can
					// optionally define a mapping between
					// the attribute and the HTTP header
					// name. If the response type defines
					// an attribute with the same name as
					// the name of the Header attribute then
					// the header attribute inherits all its
					// properties (type, description,
					// validations, etc.) from it.
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
					// Attribute defines a single body field.
					// If the response type defines an
					// attribute with the same name then the
					// Body attribute inherits all its
					// properties (type, description,
					// validations, etc.) from it.
					Attribute("response_type_attribute")
					Required("response_type_attribute")
				})
			})

			// If the method response type is Empty then the
			// response Body function must be omitted or use Empty.
			// If the method type is not Empty then the response
			// Body function may use Empty to signifiy an empty
			// response body.
			// As a convenience responses using HTTP status code 204
			// (No Content) that do not call Body default to an empty
			// body.
			Response(StatusNoContent)

			// Response defines a method specific error response.
			// The DSL is identical to API level HTTP Response
			// expressions.
			Response("service_error")
		})
	})

	// Method using the service request and response types default HTTP
	// mappings.
	Method("another_method", func() {
		Payload(PayloadType)
		Response(ResponseResultType)

		HTTP(func() {

			// No Body function means the method HTTP request body
			// is defined by the method request type PayloadType.
			PUT("/another")

			// No DSL means the response body shape and content type
			// is defined by the method response type
			// ResponseResultType.
			Response(StatusOK)
		})
	})
})

// CommonParams is an object whose attributes define HTTP parameters common to
// all the API methods.
var CommonParams = Type("CommonParams", func() {
	Attribute("query")
	Attribute("other:O")
	Required("query")
})

// CommonHeaders is an object whose attributes define HTTP headers common to all
// the API methods.
var CommonHeaders = Type("CommonHeaders", func() {
	Attribute("Header-Name")
	Attribute("attribute:Other-Name")
	Required("attribute")
})

// PayloadType is the type that describes the request parameters.
var PayloadType = Type("Payload", func() {
	Description("Optional description")
	Attribute("required", String)
	Attribute("name", String)
	Required("required")
})

// ResponseResultType is the result type that describes the response shape.
var ResponseResultType = ResultType("application/vnd.goa.response", func() {
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
	View("view", func() {
		Attribute("name")
	})
})
