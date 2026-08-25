// This file verifies that JSON-RPC responses contain only JSON-RPC message
// data. HTTP response headers and cookies cannot belong to one message in a
// batch or server-sent-event stream.
package expr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestJSONRPCSuccessResponseRejectsHTTPMetadata checks the method result
// mapping, which is the only level where a success response can be declared.
func TestJSONRPCSuccessResponseRejectsHTTPMetadata(t *testing.T) {
	for _, test := range []struct {
		name    string
		mapping func()
		wantErr string
	}{
		{
			name:    "header",
			mapping: func() { dsl.Header("value:X-Value") },
			wantErr: "JSON-RPC success response cannot map result attributes to HTTP headers",
		},
		{
			name:    "cookie",
			mapping: func() { dsl.Cookie("value:result") },
			wantErr: "JSON-RPC success response cannot map result attributes to HTTP cookies",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, jsonRPCSuccessResponseMetadataDSL(test.mapping))
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

// TestJSONRPCErrorResponseRejectsHTTPMetadata checks mappings declared at each
// level where an error response can be shared or overridden.
func TestJSONRPCErrorResponseRejectsHTTPMetadata(t *testing.T) {
	for _, scope := range []string{"API", "service", "method"} {
		for _, test := range []struct {
			name    string
			mapping func()
			wantErr string
		}{
			{
				name:    "header",
				mapping: func() { dsl.Header("detail:X-Detail") },
				wantErr: "JSON-RPC error response cannot map error attributes to HTTP headers",
			},
			{
				name:    "cookie",
				mapping: func() { dsl.Cookie("detail:error") },
				wantErr: "JSON-RPC error response cannot map error attributes to HTTP cookies",
			},
		} {
			t.Run(scope+" "+test.name, func(t *testing.T) {
				err := expr.RunInvalidDSL(t, jsonRPCErrorResponseMetadataDSL(scope, test.mapping))
				require.ErrorContains(t, err, test.wantErr)
			})
		}
	}
}

// TestResponseMetadataPreservesSupportedMappings checks that the rejection is
// limited to JSON-RPC response headers and cookies. Ordinary HTTP responses,
// JSON-RPC request headers and cookies, and server-sent-event fields remain
// valid.
func TestResponseMetadataPreservesSupportedMappings(t *testing.T) {
	expr.RunDSL(t, func() {
		dsl.Service("records", func() {
			dsl.Method("fetch", func() {
				dsl.Result(func() {
					dsl.Attribute("value", dsl.String)
					dsl.Attribute("etag", dsl.String)
					dsl.Attribute("session", dsl.String)
				})
				dsl.Error("failed", func() {
					dsl.Attribute("detail", dsl.String)
				})
				dsl.HTTP(func() {
					dsl.GET("/records")
					dsl.Response(func() {
						dsl.Header("etag:X-ETag")
						dsl.Cookie("session:SID")
					})
					dsl.Response("failed", dsl.StatusBadRequest, func() {
						dsl.Header("detail:X-Detail")
						dsl.Cookie("detail:error")
					})
				})
			})
		})
		dsl.Service("events", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/rpc")
			})
			dsl.Method("watch", func() {
				dsl.Payload(func() {
					dsl.Attribute("token", dsl.String)
					dsl.Attribute("session", dsl.String)
					dsl.Attribute("resume", dsl.String)
				})
				dsl.StreamingResult(func() {
					dsl.Attribute("value", dsl.String)
					dsl.Attribute("event_id", dsl.String)
					dsl.Attribute("event_type", dsl.String)
					dsl.Attribute("retry", dsl.Int)
				})
				dsl.JSONRPC(func() {
					dsl.Header("token:X-Token")
					dsl.Cookie("session:SID")
					dsl.ServerSentEvents(func() {
						dsl.SSERequestID("resume")
						dsl.SSEEventData("value")
						dsl.SSEEventID("event_id")
						dsl.SSEEventType("event_type")
						dsl.SSEEventRetry("retry")
					})
				})
			})
		})
	})
}

// jsonRPCSuccessResponseMetadataDSL defines one result response with the given
// transport mapping.
func jsonRPCSuccessResponseMetadataDSL(mapping func()) func() {
	return func() {
		dsl.Service("records", func() {
			dsl.Method("fetch", func() {
				dsl.Result(func() {
					dsl.Attribute("value", dsl.String)
				})
				dsl.JSONRPC(func() {
					dsl.Response(mapping)
				})
			})
		})
	}
}

// jsonRPCErrorResponseMetadataDSL defines one error mapping at the requested
// inheritance level and selects that error from one JSON-RPC method.
func jsonRPCErrorResponseMetadataDSL(scope string, mapping func()) func() {
	return func() {
		errorType := func() {
			dsl.Attribute("detail", dsl.String)
		}
		dsl.API("records", func() {
			if scope == "API" {
				dsl.Error("failed", errorType)
				dsl.JSONRPC(func() {
					dsl.Response("failed", 7001, mapping)
				})
			}
		})
		dsl.Service("records", func() {
			if scope == "service" {
				dsl.Error("failed", errorType)
				dsl.JSONRPC(func() {
					dsl.Response("failed", 7001, mapping)
				})
			}
			dsl.Method("fetch", func() {
				if scope == "API" || scope == "service" {
					dsl.Error("failed")
				} else {
					dsl.Error("failed", errorType)
				}
				dsl.JSONRPC(func() {
					if scope == "method" {
						dsl.Response("failed", 7001, mapping)
					}
				})
			})
		})
	}
}
