package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var JSONRPCSSEStringDSL = func() {
	API("jsonrpc-sse-test", func() {
		JSONRPC(func() {})
	})
	Service("JSONRPCSSEStringService", func() {
		Method("Stream", func() {
			Payload(func() {
				ID("id", String, "Request ID")
			})
			StreamingResult(String)
			JSONRPC(func() {
				GET("/stream")
				ServerSentEvents()
			})
		})
	})
}

var JSONRPCSSEObjectDSL = func() {
	API("jsonrpc-sse-test", func() {
		JSONRPC(func() {})
	})
	Service("JSONRPCSSEObjectService", func() {
		Method("Stream", func() {
			Payload(func() {
				ID("id", String, "Request ID")
				Attribute("last_event_id", String, "Last event ID")
			})
			StreamingResult(func() {
				ID("id", String, "Event ID") 
				Attribute("data", String, "Event data")
			})
			JSONRPC(func() {
				POST("/stream")
				ServerSentEvents(func() {
					SSERequestID("last_event_id")
					SSEEventID("id")
				})
			})
		})
	})
}