package design

import (
	. "goa.design/goa/dsl"
)

// API describes the global properties of the API server.
var _ = API("calc", func() {
	Title("Calculator Service")
	Description("HTTP service for adding numbers, a goa teaser")

	// Server describes a single process listening for client requests. The DSL
	// defines the set of services that the server hosts as well as hosts details.
	Server("calc", func() {
		Description("calc hosts the Calculator Service.")

		// List the services hosted by this server.
		Services("calc")

		// List the Hosts and their transport URLs.
		Host("development", func() {
			Description("Development hosts.")
			// Transport specific URLs, supported schemes are:
			// 'http', 'https', 'grpc' and 'grpcs' with the respective default
			// ports: 80, 443, 8080, 8443.
			URI("http://localhost:8000/calc")
			URI("grpc://localhost:8080")
		})

		Host("production", func() {
			Description("Production hosts.")
			// URIs can be parameterized using {param} notation.
			URI("https://{version}.goa.design/calc")
			URI("grpcs://{version}.goa.design")

			// Variable describes a URI variable.
			Variable("version", String, "API version", func() {
				// URL parameters must have a default value and/or an
				// enum validation.
				Default("v1")
			})
		})
	})
})

// Service describes a service
var _ = Service("calc", func() {
	Description("The calc service performs operations on numbers")

	// Method describes a service method (endpoint)
	Method("add", func() {
		// Payload describes the method payload.
		// Here the payload is an object that consists of two fields.
		Payload(func() {
			// Attribute describes an object field
			Field(1, "a", Int, "Left operand")
			Field(2, "b", Int, "Right operand")
			Required("a", "b")
		})

		// Result describes the method result.
		// Here the result is a simple integer value.
		Result(Int)

		// HTTP describes the HTTP transport mapping.
		HTTP(func() {
			// Requests to the service consist of HTTP GET requests.
			// The payload fields are encoded as path parameters.
			GET("/add/{a}/{b}")
			// Responses use a "200 OK" HTTP status.
			// The result is encoded in the response body.
			Response(StatusOK)
		})

		// GRPC describes the gRPC transport mapping.
		GRPC(func() {
			// Responses use a "OK" gRPC code.
			// The result is encoded in the response message.
			Response(CodeOK)
		})
	})

	// Serve the file with relative path ../../gen/http/openapi.json for
	// requests sent to /swagger.json.
	Files("/swagger.json", "../../gen/http/openapi.json")
})
