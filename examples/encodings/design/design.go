package design

import (
	. "goa.design/goa/dsl"
)

var _ = API("encodings", func() {
	Title("Encodings Service")
	Description("Encoding example service demonstrating the use of different content types")

	Server("encodings", func() {
		Services("text")
	})
})

// Service describes a service
var _ = Service("text", func() {
	Description("The text service performs operations on strings")

	// Endpoints that return a string Result are compatibile with text/html.
	Method("concatstrings", func() {
		// The payload is the two strings to be concatenated
		Payload(func() {

			Attribute("a", String, "Left operand")
			Attribute("b", String, "Right operand")
			Required("a", "b")
		})

		Result(String)

		HTTP(func() {
			// The payload fields are encoded as path parameters.
			GET("/concatstrings/{a}/{b}")
			Response(StatusOK, func() {
				// Respond with text/html
				ContentType("text/html")
			})
		})
	})

	// Endpoints that return a string Result are compatibile with text/html.
	Method("concatbytes", func() {
		// The payload is the two strings to be concatenated
		Payload(func() {

			Attribute("a", String, "Left operand")
			Attribute("b", String, "Right operand")
			Required("a", "b")
		})

		Result(Bytes)

		HTTP(func() {
			// The payload fields are encoded as path parameters.
			GET("/concatbytes/{a}/{b}")
			Response(StatusOK, func() {
				// Respond with text/html
				ContentType("text/html")
			})
		})
	})

	// Objects can't be encoded as text/html, but the response can be set to return a text field from an object.

	Method("concatstringfield", func() {
		// The payload is the two strings to be concatenated
		Payload(func() {

			Attribute("a", String, "Left operand")
			Attribute("b", String, "Right operand")
			Required("a", "b")
		})

		Result(MyConcatenation)

		HTTP(func() {
			// The payload fields are encoded as path parameters.
			GET("/concatstringfield/{a}/{b}")
			Response(StatusOK, func() {
				// Respond with text/html
				ContentType("text/html")

				// Specify the response body to be a string field of MyConcatenation
				Body("stringfield")
			})
		})
	})

	Method("concatbytesfield", func() {
		// The payload is the two strings to be concatenated
		Payload(func() {

			Attribute("a", String, "Left operand")
			Attribute("b", String, "Right operand")
			Required("a", "b")
		})

		Result(MyConcatenation)

		HTTP(func() {
			// The payload fields are encoded as path parameters.
			GET("/concatbytesfield/{a}/{b}")
			Response(StatusOK, func() {
				// Respond with text/html
				ContentType("text/html")

				// Specify the response body to be a bytes field of MyConcatenation
				Body("bytesfield")
			})
		})
	})

	// Serve the file with relative path ../../gen/http/openapi.json for
	// requests sent to /swagger.json.
	Files("/swagger.json", "../../gen/http/openapi.json")
})

var MyConcatenation = Type("MyConcatenation", func() {
	Attribute("stringfield", String)
	Attribute("bytesfield", Bytes)
})
