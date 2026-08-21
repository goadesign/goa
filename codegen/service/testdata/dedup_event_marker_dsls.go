// This file defines streaming service designs used to verify that shared result
// types emit one event marker method in generated service code.
package testdata

import (
	. "goa.design/goa/v3/dsl"
)

// StreamingDuplicateResultTypesDSL defines two streaming methods that share the same
// result type to ensure event marker methods are not duplicated in generated service code.
var StreamingDuplicateResultTypesDSL = func() {
	API("dedup-streaming", func() { JSONRPC(func() {}) })
	var SharedEvent = Type("SharedEvent", func() {
		Attribute("message", String)
		Required("message")
	})
	Service("DupStreamService", func() {
		JSONRPC(func() { POST("/stream") })
		Method("A", func() {
			StreamingResult(SharedEvent)
			JSONRPC(func() { ServerSentEvents() })
		})
		Method("B", func() {
			StreamingResult(SharedEvent)
			JSONRPC(func() { ServerSentEvents() })
		})
	})
}
