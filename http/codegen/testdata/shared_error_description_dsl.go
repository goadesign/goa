// This file defines HTTP designs that reuse one named error type from several
// methods so OpenAPI tests can distinguish type text from response text.
package testdata

import . "goa.design/goa/v3/dsl"

var (
	// SharedErrorDescriptionDSL describes the shared type and declares the first
	// method before the second method.
	SharedErrorDescriptionDSL = sharedErrorDescriptionDSL(false, "Shared error value")

	// ReversedSharedErrorDescriptionDSL declares the same methods in reverse
	// order to prove that method order does not change the shared schema.
	ReversedSharedErrorDescriptionDSL = sharedErrorDescriptionDSL(true, "Shared error value")

	// UndescribedSharedErrorDSL leaves the shared type without a description so
	// a method description cannot become the shared schema description.
	UndescribedSharedErrorDSL = sharedErrorDescriptionDSL(false, "")
)

// sharedErrorDescriptionDSL returns a design with two method errors that share
// one type but explain different failures to callers.
func sharedErrorDescriptionDSL(reverse bool, typeDescription string) func() {
	return func() {
		sharedError := Type("SharedError", func() {
			if typeDescription != "" {
				Description(typeDescription)
			}
			Attribute("message", String, "Error message", func() {
				Example("shared failure")
			})
			Required("message")
		})

		Service("errors", func() {
			first := func() {
				Method("first", func() {
					Error("first_error", sharedError, "First failure")
					HTTP(func() {
						GET("/first")
						Response("first_error", StatusBadRequest)
					})
				})
			}
			second := func() {
				Method("second", func() {
					Error("second_error", sharedError, "Second failure")
					HTTP(func() {
						GET("/second")
						Response("second_error", StatusBadRequest)
					})
				})
			}
			if reverse {
				second()
				first()
				return
			}
			first()
			second()
		})
	}
}
