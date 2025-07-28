package expr_test

import (
	"strings"
	"testing"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestHTTPServiceValidate(t *testing.T) {
	cases := []struct {
		Name          string
		DSL           func()
		Error         string
		ContainsError string
	}{
		{"valid jsonrpc websocket", validJSONRPCWebSocketDSL, "", ""},
		{"jsonrpc websocket with headers", jsonrpcWebSocketWithHeadersDSL, "", `JSON-RPC endpoint "method" using WebSocket cannot have header mappings`},
		{"jsonrpc websocket with cookies", jsonrpcWebSocketWithCookiesDSL, "", `JSON-RPC endpoint "method" using WebSocket cannot have cookie mappings`},
		{"jsonrpc websocket with params", jsonrpcWebSocketWithParamsDSL, "", `JSON-RPC endpoint "method" using WebSocket cannot have parameter mappings`},
		{"jsonrpc websocket with all mappings", jsonrpcWebSocketWithAllMappingsDSL, "", `JSON-RPC endpoint "method" using WebSocket cannot have header mappings`},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Error == "" && tc.ContainsError == "" {
				expr.RunDSL(t, tc.DSL)
			} else {
				err := expr.RunInvalidDSL(t, tc.DSL)
				if tc.Error != "" {
					if err.Error() != tc.Error {
						t.Errorf("got error %q, expected %q", err.Error(), tc.Error)
					}
				} else if tc.ContainsError != "" {
					if !strings.Contains(err.Error(), tc.ContainsError) {
						t.Errorf("error %q does not contain expected substring %q", err.Error(), tc.ContainsError)
					}
				}
			}
		})
	}
}

// Test DSL functions

var validJSONRPCWebSocketDSL = func() {
	Service("calc", func() {
		Method("method", func() {
			StreamingPayload(func() {
				ID("request_id", String)
				Attribute("data", String)
				Required("request_id")
			})
			StreamingResult(func() {
				ID("response_id", String)
				Attribute("value", String)
				Required("response_id")
			})
			JSONRPC(func() {})
		})
	})
}

var jsonrpcWebSocketWithHeadersDSL = func() {
	Service("calc", func() {
		Method("method", func() {
			StreamingPayload(func() {
				ID("request_id", String)
				Attribute("data", String)
				Required("request_id")
			})
			StreamingResult(func() {
				ID("response_id", String)
				Attribute("value", String)
				Required("response_id")
			})
			JSONRPC(func() {
				Headers(func() {
					Header("X-API-Version", String)
				})
			})
		})
	})
}

var jsonrpcWebSocketWithCookiesDSL = func() {
	Service("calc", func() {
		Method("method", func() {
			StreamingPayload(func() {
				ID("request_id", String)
				Attribute("data", String)
				Required("request_id")
			})
			StreamingResult(func() {
				ID("response_id", String)
				Attribute("value", String)
				Required("response_id")
			})
			JSONRPC(func() {
				Cookie("session", String)
			})
		})
	})
}

var jsonrpcWebSocketWithParamsDSL = func() {
	Service("calc", func() {
		Method("method", func() {
			StreamingPayload(func() {
				ID("request_id", String)
				Attribute("data", String)
				Required("request_id")
			})
			StreamingResult(func() {
				ID("response_id", String)
				Attribute("value", String)
				Required("response_id")
			})
			JSONRPC(func() {
				Params(func() {
					Param("id", String)
				})
			})
		})
	})
}

var jsonrpcWebSocketWithAllMappingsDSL = func() {
	Service("calc", func() {
		Method("method", func() {
			StreamingPayload(func() {
				ID("request_id", String)
				Attribute("data", String)
				Required("request_id")
			})
			StreamingResult(func() {
				ID("response_id", String)
				Attribute("value", String)
				Required("response_id")
			})
			JSONRPC(func() {
				Headers(func() {
					Header("X-API-Version", String)
				})
				Cookie("session", String)
				Params(func() {
					Param("id", String)
				})
			})
		})
	})
}
