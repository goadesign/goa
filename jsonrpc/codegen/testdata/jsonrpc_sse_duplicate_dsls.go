package testdata

import (
    . "goa.design/goa/v3/dsl"
)

// JSONRPCSSEDuplicateEventDSL defines two JSON-RPC SSE streaming methods that share the
// same streaming result type to ensure generated server stream switch does not duplicate cases.
var JSONRPCSSEDuplicateEventDSL = func() {
    API("jsonrpc-sse-dedupe-test", func() { JSONRPC(func() {}) })

    var SharedSSEEvent = Type("SharedSSEEvent", func() {
        Attribute("data", String)
        Required("data")
    })

    Service("JSONRPCSSEDupeService", func() {
        JSONRPC(func() { POST("/stream") })

        Method("StreamA", func() {
            StreamingResult(SharedSSEEvent)
            JSONRPC(func() { ServerSentEvents() })
        })
        Method("StreamB", func() {
            StreamingResult(SharedSSEEvent)
            JSONRPC(func() { ServerSentEvents() })
        })
    })
}

