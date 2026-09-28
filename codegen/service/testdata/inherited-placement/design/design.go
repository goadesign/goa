// Package design exercises inherited placement through a loaded Goa design.
// Only Packet selects a package; both services also use its unlocated child.
package design

import (
	external "goa.design/goa/v3/codegen/service/testdata/external-union"
	"goa.design/goa/v3/codegen/service/testdata/inherited-placement/schema"
	. "goa.design/goa/v3/dsl"
)

var (
	_ = API("placement", func() {})

	values = schema.Define()

	// Packet selects the package for its complete reachable type graph.
	Packet = Type("Packet", func() {
		Meta("struct:pkg:path", "shared/types")
		Field(1, "name", String)
		Field(2, "label", values.Text)
		Field(3, "words", ArrayOf(String))
		Field(4, "table", MapOf(String, ArrayOf(String)))
		Field(5, "entries", ArrayOf(values.Entry))
		Required("name", "entries")
		ConvertTo(external.Envelope{})
		CreateFrom(external.Envelope{})
	})

	_ = Service("alpha", func() {
		Method("exchange", func() {
			Payload(Packet)
			Result(Packet)
			HTTP(func() { POST("/alpha/exchange") })
			GRPC(func() {})
		})
		Method("child", func() {
			Payload(values.Child)
			Result(values.Child)
			HTTP(func() { POST("/alpha/child") })
			GRPC(func() {})
		})
	})

	_ = Service("beta", func() {
		Method("exchange", func() {
			Payload(Packet)
			Result(Packet)
			HTTP(func() { POST("/beta/exchange") })
			GRPC(func() {})
		})
		Method("child", func() {
			Payload(values.Child)
			Result(values.Child)
			HTTP(func() { POST("/beta/child") })
			GRPC(func() {})
		})
	})
)
