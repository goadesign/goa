// This file defines the HTTP service used to verify required primitive array
// elements in generated request and response types.
package testdata

import (
	. "goa.design/goa/v3/dsl"
)

// RequiredPrimitiveArrayDSL defines primitive and named primitive arrays whose
// JSON elements must not be null.
var RequiredPrimitiveArrayDSL = func() {
	alias := Type("RequiredArrayAlias", String, func() {
		Pattern("^[a-z]+$")
	})
	Service("RequiredArrays", func() {
		Method("Store", func() {
			Payload(func() {
				Attribute("names", ArrayOfRequired(String))
				Attribute("aliases", ArrayOfRequired(alias))
				Required("names", "aliases")
			})
			Result(func() {
				Attribute("names", ArrayOfRequired(String))
				Attribute("aliases", ArrayOfRequired(alias))
				Required("names", "aliases")
			})
			HTTP(func() {
				POST("/required-arrays")
			})
		})
	})
}
