package testdata

import (
	. "goa.design/goa/v3/dsl"
)

// JSONRPCKitchenSinkDSL exercises the full JSON-RPC generated surface in one
// design so golden tests can pin every generator output: a plain JSON-RPC
// service (required and optional request IDs, a no-payload method, a method
// with no result, custom errors with JSON-RPC code mappings), a
// WebSocket-only streaming service, an SSE streaming service, a service
// mixing HTTP and JSON-RPC transports on the same methods, and a plain HTTP
// service sharing the design.
var JSONRPCKitchenSinkDSL = func() {
	API("kitchen-sink", func() {
		JSONRPC(func() {})
	})

	Service("Calc", func() {
		JSONRPC(func() {
			POST("/rpc")
		})
		Method("add", func() {
			Payload(func() {
				ID("id", String, "Request ID")
				Attribute("a", Int)
				Attribute("b", Int)
				Required("id", "a", "b")
			})
			Result(func() {
				ID("id", String, "Request ID")
				Attribute("sum", Int)
				Required("id", "sum")
			})
			Error("overflow")
			JSONRPC(func() {
				Response("overflow", RPCInvalidParams)
			})
		})
		Method("ping", func() {
			Result(func() {
				Attribute("pong", String)
			})
			JSONRPC(func() {})
		})
		Method("log", func() {
			Payload(func() {
				ID("id", String, "Optional request ID")
				Attribute("message", String)
				Required("message")
			})
			JSONRPC(func() {})
		})
	})

	Service("Chat", func() {
		JSONRPC(func() {
			Path("/ws")
		})
		Method("echo", func() {
			StreamingPayload(func() {
				ID("id", String, "Request ID")
				Attribute("msg", String)
			})
			StreamingResult(func() {
				ID("id", String, "Request ID")
				Attribute("echo", String)
			})
			JSONRPC(func() {})
		})
	})

	Service("Feed", func() {
		JSONRPC(func() {
			POST("/feed")
		})
		Method("watch", func() {
			Payload(func() {
				ID("request_id", String, "Request ID")
				Attribute("last_event_id", String, "Last received event ID")
				Required("request_id")
			})
			StreamingResult(func() {
				Attribute("event_id", String, "Event ID")
				Attribute("data", String, "Event data")
			})
			JSONRPC(func() {
				ServerSentEvents(func() {
					SSERequestID("last_event_id")
					SSEEventID("event_id")
				})
			})
		})
	})

	Service("Mixed", func() {
		JSONRPC(func() {
			Path("/mixed/rpc")
		})
		Method("lookup", func() {
			Payload(func() {
				ID("id", String, "Request ID")
				Attribute("key", String)
				Required("id", "key")
			})
			Result(func() {
				ID("id", String, "Request ID")
				Attribute("value", String)
				Required("id")
			})
			HTTP(func() {
				POST("/lookup")
			})
			JSONRPC(func() {})
		})
	})

	Service("Health", func() {
		Method("check", func() {
			Result(String)
			HTTP(func() {
				GET("/healthz")
			})
		})
	})
}
