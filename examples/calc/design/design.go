package design

import . "goa.design/goa/http/design"
import . "goa.design/goa/http/dsl"

// API describes the global properties of the API server.
var _ = API("calc", func() {
	Title("Calculator Service")
	Description("HTTP service for adding numbers, a goa teaser")
})

// Service describes a service
var _ = Service("calc", func() {
	Description("The calc service performs operations on numbers")
	// Method describes a service method (endpoint)
	Method("add", func() {
		// Payload describes the method payload
		// Here the payload is an object that consists of two fields
		Payload(func() {
			// Attribute describes an object field
			Attribute("a", Int, "Left operand")
			Attribute("b", Int, "Right operand")
			Required("a", "b")
		})
		// Result describes the method result
		// Here the result is a simple integer value
		Result(Int)
		// HTTP describes the HTTP transport mapping
		HTTP(func() {
			// Requests to the service consist of HTTP GET requests
			// The payload fields are encoded as path parameters
			GET("/add/{a}/{b}")
			// Responses use a "200 OK" HTTP status
			// The result is encoded in the response body
			Response(StatusOK)
		})
	})
	Method("added", func() {
		/*Payload(func() {
			Attribute("foo", MapOf(String, ArrayOf(Int)), "Foo Param")
		})*/
		Payload(MapOf(String, ArrayOf(Int)))
		Result(Int)
		HTTP(func() {
			GET("/add")
			//MapParams("foo")
			MapParams()
			Response(StatusOK)
		})
	})
})

var _ = Service("openapi", func() {
	// Serve the file with relative path ../../http/openapi.json for requests
	// sent to /swagger.json.
	Files("/swagger.json", "../../gen/http/openapi.json")
})
