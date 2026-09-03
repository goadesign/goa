// This file checks that a method which returns many results cannot put result
// fields in HTTP headers or cookies. One connection has only one HTTP response,
// so it cannot carry different values for each result.
package expr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestStreamingSuccessResponseRejectsHeadersAndCookies checks HTTP methods over
// SSE and WebSocket connections and JSON-RPC methods over SSE.
func TestStreamingSuccessResponseRejectsHeadersAndCookies(t *testing.T) {
	transports := []struct {
		name string
		dsl  func(func())
	}{
		{name: "HTTP server-sent events", dsl: httpStreamingResponseMappingDSL(true)},
		{name: "HTTP WebSocket", dsl: httpStreamingResponseMappingDSL(false)},
		{name: "JSON-RPC server-sent events", dsl: jsonRPCStreamingResponseMappingDSL(true)},
	}
	mappings := []struct {
		name  string
		apply func()
		error string
	}{
		{name: "header", apply: func() { Header("metadata:X-Metadata") }, error: "streaming success response cannot map result attributes to HTTP headers"},
		{name: "cookie", apply: func() { Cookie("metadata:session") }, error: "streaming success response cannot map result attributes to HTTP cookies"},
	}

	for _, transport := range transports {
		for _, mapping := range mappings {
			t.Run(transport.name+" "+mapping.name, func(t *testing.T) {
				err := expr.RunInvalidDSL(t, func() {
					transport.dsl(mapping.apply)
				})
				require.ErrorContains(t, err, mapping.error)
			})
		}
	}
}

// httpStreamingResponseMappingDSL creates an HTTP method which returns many
// results and places one result field in its HTTP response.
func httpStreamingResponseMappingDSL(sse bool) func(func()) {
	return func(mapping func()) {
		Service("stream", func() {
			Method("watch", func() {
				StreamingResult(streamingMappedResult())
				HTTP(func() {
					GET("/watch")
					if sse {
						ServerSentEvents(func() {})
					}
					Response(func() {
						mapping()
					})
				})
			})
		})
	}
}

// jsonRPCStreamingResponseMappingDSL creates a JSON-RPC method which returns
// many results and places one result field in its HTTP response.
func jsonRPCStreamingResponseMappingDSL(sse bool) func(func()) {
	return func(mapping func()) {
		Service("stream", func() {
			JSONRPC(func() {
				if sse {
					POST("/watch")
				} else {
					GET("/watch")
				}
			})
			Method("watch", func() {
				StreamingResult(streamingMappedResult())
				JSONRPC(func() {
					if sse {
						ServerSentEvents(func() {})
					}
					Response(func() {
						mapping()
					})
				})
			})
		})
	}
}

// streamingMappedResult defines the two fields used by these tests.
func streamingMappedResult() func() {
	return func() {
		Attribute("value", String)
		Attribute("metadata", String)
	}
}
