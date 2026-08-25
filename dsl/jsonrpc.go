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
// a single HTTP POST endpoint. Methods may stream results over Server-Sent Events.
//
// JSONRPC can be used at three levels:
//
//   - At the API level: JSONRPC defines default error code mappings that
//     services and methods may use.
//   - At the service level: JSONRPC sets the HTTP endpoint path for all
//     JSON-RPC methods in the service and defines common errors and their
//     error code mappings.
//   - At the method level: JSONRPC configures an optional request ID payload
//     field, whether generated clients send a notification, and method-specific
//     error code mappings.
//
// Request Handling:
//
// The generated code decodes the JSON-RPC "params" field into the method
// payload. An object, map, array, or union keeps its JSON shape. A primitive,
// named primitive, byte slice, or Any value is carried as one positional value,
// for example ["hello"]. The "id" field is decoded separately into the direct
// payload field specified by ID.
//
// Ordinary Requests and Notifications:
//
// An ordinary request always includes an ID. If the payload does not define an
// ID field, the generated client creates one. Notification explicitly marks a
// method as one-way: the generated client omits the ID and does not decode a
// JSON-RPC response. A method with no result remains an ordinary request unless
// it calls Notification.
//
// Non-Streaming Batch Requests:
//
// The generated code fully supports batch JSON-RPC requests: when the HTTP
// request body contains an array of JSON-RPC request objects, it will unmarshal
// the array, process each request independently (including error handling and
// notifications), and marshal the responses into a single array of JSON-RPC
// response objects in the HTTP response body.
//
// Server-Sent Events:
//
// A JSON-RPC method may stream results by defining StreamingResult() and calling
// ServerSentEvents() in its method-level JSONRPC expression. The client sends one
// JSON-RPC request. Each streamed value is sent as a complete JSON-RPC message in
// a separate SSE event. SSEEventData selects the value used for notification
// params. SSEEventID, SSEEventType, and SSEEventRetry map result fields to the
// corresponding outer SSE lines; Goa omits each line when it has no mapping.
// SSERequestID maps a payload field to the Last-Event-ID request header. A
// viewed streaming result cannot select one field with SSEEventData because the
// selected value would omit the view name needed by the client. JSON-RPC does
// not support StreamingPayload(), bidirectional streaming, or one method that
// defines both Result() and StreamingResult(). Use separate methods when
// clients need both a stream and a final resource.
//
// Using JSON-RPC with Other Transports:
//
// Goa allows you to expose a single service or method over multiple transports.
// For example, a method can have both standard HTTP or gRPC endpoints in addition
// to a JSON-RPC endpoint.
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
// Example - Service with request and notification handling:
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
//	    Method("record", func() {
//	        Payload(func() {
//	            Attribute("message", String)
//	            Required("message")
//	        })
//	        JSONRPC(func() {
//	            Notification()
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
//	            Attribute("event_id", String, "Event ID")
//	            Attribute("data", Data, "Event data")
//	        })
//	        JSONRPC(func() {
//	            ServerSentEvents(func() {         // Stream results as server-sent events
//	                SSERequestID("last_event_id") // Map SSE Last-Event-ID header to payload "last_event_id" attribute
//	                SSEEventID("event_id")        // Use "event_id" as the SSE event ID
//	            })
//	        })
//	    })
//	})
func JSONRPC(dsl func()) {
	switch actual := eval.Current().(type) {
	case *expr.APIExpr:
		previous := actual.JSONRPC.DSLFunc
		actual.JSONRPC.DSLFunc = func() {
			if previous != nil {
				previous()
			}
			dsl()
		}
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

// ID defines the payload attribute used as the JSON-RPC request ID. It must be
// a string. When an ordinary request has no ID attribute, the generated client
// creates an ID. An optional field with no default is a pointer: nil makes the
// client create an ID, while a non-nil value, including an empty string, is sent
// exactly. A field with a default is a value, and an absent incoming ID uses
// that default. A required field with no default rejects an absent or null ID.
// Notifications cannot define an ID attribute.
//
// The transport copies the exact request ID into every JSON-RPC response. A
// service result cannot define or replace it.
//
// ID must appear as a direct field in a Payload expression. A payload may
// define at most one ID field.
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
//	    Result(Float64)
//	    JSONRPC(func() {})
//	})
func ID(name string, args ...any) {
	args = useDSL(args, func() { Meta("jsonrpc:id", "") })
	Attribute(name, args...)
}

// Notification makes generated clients call the JSON-RPC method without a
// request ID or a JSON-RPC response. The method may have a payload but cannot
// define a result, an ID field, or a stream.
//
// Notification must appear in a method-level JSONRPC expression.
//
// Example:
//
//	Method("record", func() {
//	    Payload(func() {
//	        Attribute("message", String)
//	        Required("message")
//	    })
//	    JSONRPC(func() {
//	        Notification()
//	    })
//	})
func Notification() {
	endpoint, ok := eval.Current().(*expr.HTTPEndpointExpr)
	if !ok || !endpoint.IsJSONRPC() {
		eval.IncompatibleDSL()
		return
	}
	endpoint.JSONRPCNotification = true
}
