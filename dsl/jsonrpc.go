package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

const (
	// RPCParseError indicates invalid JSON was received by the server.
	// An error occurred on the server while parsing the JSON text.
	RPCParseError = expr.RPCParseError

	// RPCInvalidRequest indicates the JSON sent is not a valid Request object.
	RPCInvalidRequest = expr.RPCInvalidRequest

	// RPCMethodNotFound indicates the method does not exist or is not available.
	RPCMethodNotFound = expr.RPCMethodNotFound

	// RPCInvalidParams indicates invalid method parameters.
	RPCInvalidParams = expr.RPCInvalidParams

	// RPCInternalError indicates an internal JSON-RPC error occurred.
	// This is the default error code for unmapped errors.
	RPCInternalError = expr.RPCInternalError
)

// JSONRPC configures a service to use JSON-RPC 2.0 transport.
// The generated code handles JSON-RPC protocol details: request parsing, method dispatch,
// response formatting, and batch processing. All service JSON-RPC methods share
// a single HTTP endpoint and must use the same transport (HTTP, WebSocket or SSE).
//
// JSONRPC can be used at three levels:
//
//   - At the API level: JSONRPC maps service errors to standard JSON-RPC error
//     codes.
//   - At the service level: JSONRPC sets the HTTP endpoint path for all
//     JSON-RPC methods in the service and defines common errors and their
//     error code mappings.
//   - At the method level: JSONRPC configures how the request and response "id"
//     fields are mapped to payload/result attributes and allows you to define
//     method-specific error code mappings. Methods without Result() are automatically
//     treated as notifications (no response expected).
//
// Request Handling:
//
// The generated code decodes the JSON-RPC "params" field into the method
// payload and the "id" field to the payload attribute specified by the ID
// function.
//
// Non-Streaming:
//
// For non-streaming methods, if the result's ID attribute is not
// already set, the generated code automatically copies the request ID from the
// payload to the result's ID attribute.
//
// Non-Streaming Batch Requests:
//
// The generated code fully supports batch JSON-RPC requests: when the HTTP
// request body contains an array of JSON-RPC request objects, it will unmarshal
// the array, process each request independently (including error handling and
// notifications), and marshal the responses into a single array of JSON-RPC
// response objects in the HTTP response body.
//
// WebSocket:
//
// For WebSocket transport, methods that use StreamingPayload() and/or StreamingResult()
// enable bidirectional streaming: each payload or result element is sent as a separate,
// complete JSON-RPC message over the WebSocket connection. When using WebSockets, all
// methods must use StreamingPayload() for their payload (if any) and StreamingResult()
// for their result (if any), because a single WebSocket connection is shared by all
// methods of a service and client. Non-streaming methods are not supported over WebSockets.
//
// WebSocket methods can have three patterns:
//   - StreamingPayload() only: Client-to-server notifications (no response)
//   - StreamingResult() only: Server-to-client notifications (no request ID, sent without client request)
//   - Both StreamingPayload() and StreamingResult(): Bidirectional request/response streaming
//
// Server-side notifications (methods with StreamingResult() but no StreamingPayload()) are
// sent from the server to the client without an associated request ID, as they are not
// responses to client requests but rather server-initiated messages.
//
// Server-Sent Events:
//
// For Server-Sent Events (SSE), enable SSE by calling the ServerSentEvents() function
// within the JSONRPC expression. In this mode, each element of the result is sent as a
// separate JSON-RPC response within its own SSE event. The SSE id field is mapped to
// the result's ID attribute. Because all methods for a given service and client
// share the same HTTP endpoint, every method must use both StreamingResult() and
// ServerSentEvents() to ensure correct streaming behavior.
//
// Using JSON-RPC with Other Transports:
//
// Goa allows you to expose a single service or method over multiple transports.
// For example, a method can have both standard HTTP or gRPC endpoints in addition
// to a JSON-RPC endpoint.
//
// Important WebSocket Limitation:
//
// A service cannot mix JSON-RPC WebSocket endpoints with pure HTTP WebSocket endpoints.
// This is because JSON-RPC WebSocket uses a single underlying WebSocket connection
// for all methods in the service, with method dispatch happening at the protocol level
// through JSON-RPC message routing. In contrast, pure HTTP WebSocket creates individual
// connections per streaming endpoint. These two approaches are fundamentally incompatible
// and cannot coexist in the same service.
//
// Error Codes:
//
// Use the predefined constants for standard JSON-RPC errors:
//   - RPCParseError (-32700): Invalid JSON
//   - RPCInvalidRequest (-32600): Invalid Request object
//   - RPCMethodNotFound (-32601): Method not found
//   - RPCInvalidParams (-32602): Invalid method parameters
//   - RPCInternalError (-32603): Internal JSON-RPC error (default for unmapped errors)
//
// Example - Complete service with request/notification handling and streaming:
//
//	Service("calc", func() {
//	    Error("timeout", ErrTimeout, "Request timed out") // Define an error that all service methods can return
//
//	    JSONRPC(func() {
//	        Response("timeout", func() {  // Define JSON-RPC error code for timeout
//	            Code(5001)
//	        })
//	    })
//
//	    Method("divide", func() {
//	        Payload(func() {
//	            ID("req_id") // Map request ID to payload field
//	            Attribute("dividend", Int, "Dividend")
//	            Attribute("divisor", Int, "Divisor")
//	            Required("dividend", "divisor")
//	        })
//	        Result(func() {
//	            ID("req_id") // Map request ID to result field
//	            Attribute("result", Float64)
//	        })
//	        Error("div_zero", ErrorResult, "Division by zero") // Define method-specific error
//	        JSONRPC(func() {
//	            Response("div_zero", RPCInvalidParams) // Map div_zero error to JSON-RPC code
//	        })
//	        HTTP(func() {
//	            POST("/divide") // Also define a standard HTTP endpoint
//	        })
//	    })
//	})
//
// Example - WebSocket streaming service:
//
//	Service("chat", func() {
//	    JSONRPC(func() {
//	        GET("/ws") // Use GET for WebSocket endpoint
//	    })
//	    Method("send", func() {
//	        StreamingPayload(func() {
//	            Attribute("message", String, "Message to send")
//	        })
//	        JSONRPC(func() {
//	            // Client-to-server notification (no response)
//	        })
//	    })
//	    Method("notify", func() {
//	        StreamingResult(func() {
//	            Attribute("event", String, "Server notification")
//	            Attribute("data", Any, "Notification data")
//	        })
//	        JSONRPC(func() {
//	            // Server-to-client notification (no request ID, server-initiated)
//	        })
//	    })
//	    Method("echo", func() {
//	        StreamingPayload(func() {
//	            ID("req_id", String, "Request ID")
//	            Attribute("message", String, "Message to echo")
//	        })
//	        StreamingResult(func() {
//	            ID("req_id", String, "Request ID")
//	            Attribute("echo", String, "Echoed message")
//	        })
//	        JSONRPC(func() {
//	            // Bidirectional request/response streaming
//	        })
//	    })
//	})
//
// Example - SSE streaming service:
//
//	Service("updater", func() {
//	    JSONRPC(func() {
//	        POST("/sse") // Use POST for SSE endpoint
//	    })
//	    Method("listen", func() {
//	        Payload(func() {
//	            ID("id", String, "JSON-RPC request ID")
//	            Attribute("last_event_id", String, "ID of last event received by client")
//	        })
//	        StreamingResult(func() {
//	            ID("id", String, "JSON-RPC request ID")
//	            Attribute("data", Data, "Event data")
//	        })
//	        JSONRPC(func() {
//	            ServerSentEvents(func() {         // Use SSE instead of WebSocket
//	                SSERequestID("last_event_id") // Map SSE Last-Event-ID header to payload "last_event_id" attribute
//	                SSEEventID("id")              // Use "id" result attribute as SSE event ID
//	            })
//	        })
//	    })
//	})
func JSONRPC(dsl func()) {
	switch actual := eval.Current().(type) {
	case *expr.APIExpr:
		eval.Execute(dsl, actual.JSONRPC)
	case *expr.ServiceExpr:
		svc := expr.Root.API.JSONRPC.ServiceFor(actual, &expr.Root.API.JSONRPC.HTTPExpr)
		svc.DSLFunc = dsl
		// Mark service as JSON-RPC
		if actual.Meta == nil {
			actual.Meta = expr.MetaExpr{}
		}
		actual.Meta["jsonrpc:service"] = []string{}
	case *expr.MethodExpr:
		// Auto-enable JSON-RPC on service if not already enabled
		if actual.Service.Meta == nil {
			actual.Service.Meta = expr.MetaExpr{}
		}
		actual.Service.Meta["jsonrpc:service"] = []string{}

		svc := expr.Root.API.JSONRPC.ServiceFor(actual.Service, &expr.Root.API.JSONRPC.HTTPExpr)
		e := svc.EndpointFor(actual)
		if e.Meta == nil {
			e.Meta = expr.MetaExpr{}
		}
		e.Meta["jsonrpc"] = []string{}
		if actual.Meta == nil {
			actual.Meta = expr.MetaExpr{}
		}
		actual.Meta["jsonrpc"] = []string{}
		e.DSLFunc = dsl
	default:
		eval.IncompatibleDSL()
	}
}

// ID defines the payload or result attribute which is used as the JSON-RPC
// request ID. It must be of type String. It is an error to omit ID on a
// JSON-RPC endpoint payload or result unless the method is a notification (see
// Notification).
//
// Note: For non-streaming methods, the generated code will automatically copy
// the request ID from the payload to the result's ID attribute, unless the
// result's ID attribute is already set.
//
// ID must appear in a Payload or Result expression.
//
// ID accepts the same arguments as the Attribute DSL function.
//
// Example:
//
//	Method("calculate", func() {
//	    Payload(func() {
//	        ID("request_id", String, "Unique request identifier")
//	        Attribute("expression", String, "Mathematical expression")
//	        Required("request_id", "expression")
//	    })
//	    Result(func() {
//	        ID("request_id", String, "Unique request identifier")
//	        Attribute("result", Float64)
//	        Required("request_id", "result")
//	    })
//	    JSONRPC(func() {
//	        POST("/")
//	    })
//	})
func ID(name string, args ...any) {
	args = useDSL(args, func() { Meta("jsonrpc:id", "") })
	Attribute(name, args...)
}
