package testdata

import (
	. "goa.design/goa/v3/dsl"
)

// StreamingAliasedArrayDSL defines a service with a streaming endpoint that uses
// an array of aliased types with a custom package path.
var StreamingAliasedArrayDSL = func() {
	// Define a type in a custom package
	var CustomInt = Type("CustomInt", Int, func() {
		Meta("struct:pkg:path", "github.com/example/custompkg")
	})

	var PayloadType = Type("PayloadType", func() {
		Attribute("values", ArrayOf(CustomInt))
	})

	Service("StreamingAliasedArray", func() {
		Method("Stream", func() {
			// Use an array of the custom type as streaming payload
			StreamingPayload(PayloadType)
			HTTP(func() {
				GET("/stream")
			})
		})
	})
}
