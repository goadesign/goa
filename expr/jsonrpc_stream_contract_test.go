// This file checks which Goa stream shapes JSON-RPC can represent.
package expr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestJSONRPCStreamContract checks the request and response stream shapes that
// JSON-RPC can carry over HTTP and server-sent events.
func TestJSONRPCStreamContract(t *testing.T) {
	valid := []struct {
		name   string
		method func()
	}{
		{
			name: "one request and one response",
			method: func() {
				dsl.Payload(dsl.String)
				dsl.Result(dsl.String)
				dsl.JSONRPC(func() {})
			},
		},
		{
			name: "server stream over server sent events",
			method: func() {
				dsl.Payload(dsl.String)
				dsl.StreamingResult(dsl.String)
				dsl.JSONRPC(func() {
					dsl.ServerSentEvents(func() {})
				})
			},
		},
	}
	for _, test := range valid {
		t.Run(test.name, func(t *testing.T) {
			expr.RunDSL(t, jsonRPCStreamDSL(test.method))
		})
	}

	invalid := []struct {
		name    string
		method  func()
		wantErr string
	}{
		{
			name: "client stream",
			method: func() {
				dsl.StreamingPayload(dsl.String)
				dsl.Result(dsl.String)
				dsl.JSONRPC(func() {})
			},
			wantErr: `JSON-RPC method "stream" cannot use client streaming because one JSON-RPC request contains one params value`,
		},
		{
			name: "bidirectional stream",
			method: func() {
				dsl.StreamingPayload(dsl.String)
				dsl.StreamingResult(dsl.String)
				dsl.JSONRPC(func() {})
			},
			wantErr: `JSON-RPC method "stream" cannot use bidirectional streaming because one JSON-RPC request contains one params value`,
		},
		{
			name: "server stream without server sent events",
			method: func() {
				dsl.Payload(dsl.String)
				dsl.StreamingResult(dsl.String)
				dsl.JSONRPC(func() {})
			},
			wantErr: `JSON-RPC method "stream" with a streaming result must use ServerSentEvents()`,
		},
		{
			name: "synchronous and streaming results",
			method: func() {
				dsl.Payload(dsl.String)
				dsl.Result(dsl.Int)
				dsl.StreamingResult(dsl.String)
				dsl.JSONRPC(func() {
					dsl.ServerSentEvents(func() {})
				})
			},
			wantErr: `JSON-RPC method "stream" cannot define both Result and StreamingResult because its client stream cannot return a separate final result`,
		},
		{
			name: "matching synchronous and streaming results",
			method: func() {
				dsl.Payload(dsl.String)
				dsl.Result(dsl.String)
				dsl.StreamingResult(dsl.String)
				dsl.JSONRPC(func() {
					dsl.ServerSentEvents(func() {})
				})
			},
			wantErr: `JSON-RPC method "stream" cannot define both Result and StreamingResult because its client stream cannot return a separate final result`,
		},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, jsonRPCStreamDSL(test.method))
			require.Contains(t, err.Error(), test.wantErr)
		})
	}
}

// TestJSONRPCSSEViewedDataContract rejects a wire shape that cannot carry the
// result view the client needs to rebuild the streamed result.
func TestJSONRPCSSEViewedDataContract(t *testing.T) {
	err := expr.RunInvalidDSL(t, func() {
		viewed := dsl.ResultType("ViewedEvent", func() {
			dsl.Attribute("message", dsl.String)
			dsl.View("default", func() {
				dsl.Attribute("message")
			})
		})
		dsl.Service("streamer", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/rpc")
			})
			dsl.Method("stream", func() {
				dsl.StreamingResult(viewed)
				dsl.JSONRPC(func() {
					dsl.ServerSentEvents(func() {
						dsl.SSEEventData("message")
					})
				})
			})
		})
	})
	require.Contains(t, err.Error(), "the selected data would omit the result view needed to decode the stream")
}

// TestJSONRPCSSERequestIDOwnsLastEventID rejects another mapping that would
// make two fields compete for the same request header.
func TestJSONRPCSSERequestIDOwnsLastEventID(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "same field", header: "resume:X-Resume", want: "cannot also be mapped with Header"},
		{name: "another field", header: "other:last-event-id", want: "is reserved for SSE request ID field"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, func() {
				dsl.Service("streamer", func() {
					dsl.JSONRPC(func() {
						dsl.POST("/rpc")
					})
					dsl.Method("stream", func() {
						dsl.Payload(func() {
							dsl.Attribute("resume", dsl.String)
							dsl.Attribute("other", dsl.String)
						})
						dsl.StreamingResult(dsl.String)
						dsl.JSONRPC(func() {
							dsl.Header(test.header)
							dsl.ServerSentEvents(func() {
								dsl.SSERequestID("resume")
							})
						})
					})
				})
			})
			require.Contains(t, err.Error(), test.want)
		})
	}
}

// jsonRPCStreamDSL exposes one method through the shared JSON-RPC HTTP route.
func jsonRPCStreamDSL(method func()) func() {
	return func() {
		dsl.Service("streamer", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/rpc")
			})
			dsl.Method("stream", func() {
				method()
			})
		})
	}
}
