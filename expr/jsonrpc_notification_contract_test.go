// This file checks the design rules for one-way JSON-RPC calls.
package expr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

const jsonRPCInternalError = -32603

// TestJSONRPCNotificationContract checks that notifications are explicit and
// cannot declare response or request-ID behavior.
func TestJSONRPCNotificationContract(t *testing.T) {
	t.Run("notification with payload", func(t *testing.T) {
		root := expr.RunDSL(t, jsonRPCNotificationDSL(func() {
			dsl.Payload(func() {
				dsl.Attribute("message", dsl.String)
				dsl.Required("message")
			})
			dsl.JSONRPC(func() {
				dsl.Notification()
			})
		}))
		endpoint := root.API.JSONRPC.Services[0].HTTPEndpoints[0]
		require.True(t, endpoint.IsJSONRPCNotification())
	})

	t.Run("ordinary method without result", func(t *testing.T) {
		root := expr.RunDSL(t, jsonRPCNotificationDSL(func() {
			dsl.JSONRPC(func() {})
		}))
		endpoint := root.API.JSONRPC.Services[0].HTTPEndpoints[0]
		require.False(t, endpoint.IsJSONRPCNotification())
	})

	t.Run("named string request ID", func(t *testing.T) {
		expr.RunDSL(t, func() {
			requestID := dsl.Type("RequestID", dsl.String)
			dsl.Service("notifications", func() {
				dsl.JSONRPC(func() {
					dsl.POST("/rpc")
				})
				dsl.Method("notify", func() {
					dsl.Payload(func() {
						dsl.ID("request_id", requestID)
					})
					dsl.JSONRPC(func() {})
				})
			})
		})
	})

	t.Run("notification outside JSON-RPC", func(t *testing.T) {
		err := expr.RunInvalidDSL(t, func() {
			dsl.Service("records", func() {
				dsl.Method("record", func() {
					dsl.HTTP(func() {
						dsl.POST("/records")
						dsl.Notification()
					})
				})
			})
		})
		require.Contains(t, err.Error(), `invalid use of Notification in service "records" HTTP endpoint "record"`)
	})

	invalid := []struct {
		name    string
		method  func()
		wantErr string
	}{
		{
			name: "result",
			method: func() {
				dsl.Result(dsl.String)
				dsl.JSONRPC(func() {
					dsl.Notification()
				})
			},
			wantErr: `JSON-RPC notification "notify" cannot define a result because notifications receive no response`,
		},
		{
			name: "request ID",
			method: func() {
				dsl.Payload(func() {
					dsl.ID("request_id")
				})
				dsl.JSONRPC(func() {
					dsl.Notification()
				})
			},
			wantErr: `JSON-RPC notification "notify" cannot define an ID field because notifications omit the request ID`,
		},
		{
			name: "result ID",
			method: func() {
				dsl.Result(func() {
					dsl.ID("request_id")
				})
				dsl.JSONRPC(func() {})
			},
			wantErr: `JSON-RPC method "notify" cannot define an ID field in its result because the transport copies the request ID`,
		},
		{
			name: "error ID",
			method: func() {
				dsl.Error("failed", func() {
					dsl.ID("request_id")
				})
				dsl.JSONRPC(func() {
					dsl.Response("failed", jsonRPCInternalError)
				})
			},
			wantErr: `JSON-RPC error "failed" cannot define an ID field because the transport copies the request ID`,
		},
		{
			name: "non-string request ID",
			method: func() {
				dsl.Payload(func() {
					dsl.ID("request_id", dsl.Int)
				})
				dsl.JSONRPC(func() {})
			},
			wantErr: `JSON-RPC request ID field "request_id" must be a string`,
		},
		{
			name: "nested request ID",
			method: func() {
				dsl.Payload(func() {
					dsl.Attribute("request", func() {
						dsl.ID("id")
					})
				})
				dsl.JSONRPC(func() {})
			},
			wantErr: `JSON-RPC request ID field "id" must be a direct payload field`,
		},
		{
			name: "multiple request IDs",
			method: func() {
				dsl.Payload(func() {
					dsl.ID("first")
					dsl.ID("second")
				})
				dsl.JSONRPC(func() {})
			},
			wantErr: `JSON-RPC method "notify" cannot define more than one request ID field`,
		},
		{
			name: "stream",
			method: func() {
				dsl.StreamingResult(dsl.String)
				dsl.JSONRPC(func() {
					dsl.Notification()
					dsl.ServerSentEvents(func() {})
				})
			},
			wantErr: `JSON-RPC notification "notify" cannot stream because notifications send one message and receive no response`,
		},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, jsonRPCNotificationDSL(test.method))
			require.Contains(t, err.Error(), test.wantErr)
		})
	}
}

// TestJSONRPCRequestIDBodyContract checks that the request ID exists only in
// the JSON-RPC envelope and cannot also be mapped into params.
func TestJSONRPCRequestIDBodyContract(t *testing.T) {
	t.Run("computed params omit ID", func(t *testing.T) {
		root := expr.RunDSL(t, jsonRPCNotificationDSL(func() {
			dsl.Payload(func() {
				dsl.ID("request_id")
				dsl.Attribute("message", dsl.String)
				dsl.Required("request_id", "message")
			})
			dsl.JSONRPC(func() {})
		}))
		body := root.API.JSONRPC.Services[0].HTTPEndpoints[0].Body
		require.Nil(t, body.Find("request_id"))
		require.NotNil(t, body.Find("message"))
		require.False(t, body.IsRequired("request_id"))
		require.True(t, body.IsRequired("message"))
	})

	t.Run("ID-only payload has no params body", func(t *testing.T) {
		root := expr.RunDSL(t, jsonRPCNotificationDSL(func() {
			dsl.Payload(func() {
				dsl.ID("request_id")
				dsl.Required("request_id")
			})
			dsl.JSONRPC(func() {})
		}))
		body := root.API.JSONRPC.Services[0].HTTPEndpoints[0].Body
		require.Equal(t, expr.Empty, body.Type)
	})

	t.Run("explicit params reject ID", func(t *testing.T) {
		err := expr.RunInvalidDSL(t, jsonRPCNotificationDSL(func() {
			dsl.Payload(func() {
				dsl.ID("request_id")
				dsl.Attribute("message", dsl.String)
			})
			dsl.JSONRPC(func() {
				dsl.Body(func() {
					dsl.Attribute("request_id")
					dsl.Attribute("message")
				})
			})
		}))
		require.Contains(t, err.Error(), `JSON-RPC request ID field "request_id" cannot also appear in params`)
	})
}

// TestJSONRPCRequestIDTransportContract checks that the envelope ID is not
// also read from another part of the HTTP request.
func TestJSONRPCRequestIDTransportContract(t *testing.T) {
	tests := []struct {
		name    string
		mapping func()
		wantErr string
	}{
		{name: "query parameter", mapping: func() { dsl.Param("request_id") }, wantErr: `cannot also be mapped as an HTTP parameter`},
		{name: "header", mapping: func() { dsl.Header("request_id") }, wantErr: `cannot also be mapped as an HTTP header`},
		{name: "cookie", mapping: func() { dsl.Cookie("request_id") }, wantErr: `cannot also be mapped as an HTTP cookie`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := expr.RunInvalidDSL(t, jsonRPCNotificationDSL(func() {
				dsl.Payload(func() {
					dsl.ID("request_id")
				})
				dsl.JSONRPC(func() {
					test.mapping()
				})
			}))
			require.Contains(t, err.Error(), test.wantErr)
		})
	}

	t.Run("path parameter", func(t *testing.T) {
		err := expr.RunInvalidDSL(t, func() {
			dsl.Service("records", func() {
				dsl.JSONRPC(func() {
					dsl.POST("/rpc/{request_id}")
				})
				dsl.Method("lookup", func() {
					dsl.Payload(func() {
						dsl.ID("request_id")
					})
					dsl.JSONRPC(func() {
						dsl.Param("request_id")
					})
				})
			})
		})
		require.Contains(t, err.Error(), `cannot also be mapped as an HTTP parameter`)
	})

	t.Run("SSE reconnect header", func(t *testing.T) {
		err := expr.RunInvalidDSL(t, jsonRPCNotificationDSL(func() {
			dsl.Payload(func() {
				dsl.ID("request_id")
			})
			dsl.StreamingResult(dsl.String)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSERequestID("request_id")
				})
			})
		}))
		require.Contains(t, err.Error(), `cannot also be mapped to the Last-Event-ID header`)
	})

	t.Run("ordinary HTTP mapping remains valid", func(t *testing.T) {
		expr.RunDSL(t, func() {
			dsl.Service("records", func() {
				dsl.JSONRPC(func() {
					dsl.POST("/rpc")
				})
				dsl.Method("lookup", func() {
					dsl.Payload(func() {
						dsl.ID("request_id")
						dsl.Required("request_id")
					})
					dsl.HTTP(func() {
						dsl.GET("/records/{request_id}")
						dsl.Param("request_id")
					})
					dsl.JSONRPC(func() {})
				})
			})
		})
	})
}

// TestJSONRPCMethodNameContract checks that application methods do not use the
// namespace reserved by JSON-RPC itself.
func TestJSONRPCMethodNameContract(t *testing.T) {
	err := expr.RunInvalidDSL(t, func() {
		dsl.Service("records", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/rpc")
			})
			dsl.Method("rpc.lookup", func() {
				dsl.JSONRPC(func() {})
			})
		})
	})
	require.Contains(t, err.Error(), `JSON-RPC method "rpc.lookup" cannot begin with "rpc."`)
}

// jsonRPCNotificationDSL exposes one method through the shared JSON-RPC route.
func jsonRPCNotificationDSL(method func()) func() {
	return func() {
		dsl.Service("notifications", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/rpc")
			})
			dsl.Method("notify", method)
		})
	}
}
