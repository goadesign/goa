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

// JSONRPC configures a service or method to use JSON-RPC 2.0 transport.
// The generated code handles JSON-RPC protocol details: request parsing, method dispatch,
// response formatting, and batch processing. All service JSON-RPC methods share
// a single HTTP endpoint.
//
// At API level, JSONRPC maps global errors to JSON-RPC error codes.
// At service level, it configures the HTTP endpoint and common settings.
// At method level, it configures request ID mapping and method-specific settings.
//
// Request Handling:
// The generated code unmarshals the JSON-RPC "params" field into the method payload.
// Use ID("field") to map a payload attribute to the request "id" field, enabling
// the method to distinguish between requests (with ID) and notifications (without ID).
// Without ID mapping, all requests are treated as notifications.
//
// Streaming:
// Methods using StreamingResult() support either Server-Sent Events or WebSockets.
// With SSE, each result element is sent as a JSON-RPC response in a separate event.
// With WebSockets, methods can use StreamingPayload() for bidirectional streaming,
// where each payload/result element is sent as a complete JSON-RPC message.
//
// Error Codes:
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
//	    Error("timeout", ErrTimeout, "Request timed out") // ErrTimeout must have a limit attribute
//
//	    JSONRPC(func() {
//	        POST("/rpc")                                    // All methods use this endpoint
//	        Response("timeout", func() {                    // Custom error response
//	            Code(5001)                                  // Application error code
//	        })
//	    })
//
//	    Method("add", func() { // Notification method (no ID mapping)
//	        Payload(func() {
//	            Attribute("a", Int, "First operand")
//	            Attribute("b", Int, "Second operand")
//	            Required("a", "b")
//	        })
//	        Result(Int)
//	        JSONRPC(func() {}) // Generate JSON-RPC transport code for this method
//	    })
//
//	    Method("divide", func() { // Request/response method
//	        Payload(func() {
//	            Attribute("req_id", String, "Request ID") // Will contain JSON-RPC request ID
//	            Attribute("dividend", Int, "Dividend")
//	            Attribute("divisor", Int, "Divisor")
//	            Required("dividend", "divisor")
//	        })
//	        Result(Float64)
//	        Error("div_zero", ErrorResult, "Division by zero")
//
//	        JSONRPC(func() {
//	            ID("req_id")                           // Map request ID to payload field
//	            Response("div_zero", RPCInvalidParams) // Map div_zero error to JSON-RPC code
//	        })
//	    })
//
//	    Method("updates", func() {                    // SSE streaming method
//	        Payload(func() {
//	            Attribute("req_id", String, "Request ID")
//	            Attribute("last_event_id", String, "ID of last event received by client")
//	        })
//	        StreamingResult(func() {
//	            Attribute("event_id", String, "Event ID")
//	            Attribute("data", Data, "Event data")
//	        })
//
//	        JSONRPC(func() {
//	            ID("req_id")                       // Map JSON-RPC request ID to "req_id" payload attribute
//	            ServerSentEvents(func() {          // Use SSE instead of WebSocket
//	                SSERequestID("last_event_id")  // Map SSE Last-Event-ID header to payload "last_event_id" attribute
//	                SSEEventID("event_id")         // Use "event_id" result attribute as SSE event ID
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
	case *expr.MethodExpr:
		svc := expr.Root.API.JSONRPC.ServiceFor(actual.Service, &expr.Root.API.JSONRPC.HTTPExpr)
		e := svc.EndpointFor(actual)
		e.DSLFunc = dsl
		r := &expr.RouteExpr{Method: "POST", Path: "/", Endpoint: e}
		e.Routes = []*expr.RouteExpr{r}
	default:
		eval.IncompatibleDSL()
	}
}

// ID maps a payload attribute to the JSON-RPC request ID field.
//
// By default, Goa looks for an attribute named "id" in the payload to use as
// the JSON-RPC request ID. ID allows overriding this default to use a
// different attribute name.
//
// The specified attribute must exist in the method payload and should be of
// type String. If the attribute doesn't exist or ID is not specified,
// the generated code will automatically generate request IDs on the client side.
//
// The JSON-RPC response ID is automatically set to match the request ID
// according to the JSON-RPC specification.
//
// ID must appear in a JSONRPC expression within a Method.
//
// ID accepts one argument: the name of the payload attribute.
//
// Example:
//
//	Method("calculate", func() {
//	    Payload(func() {
//	        Attribute("request_id", String, "Unique request identifier")
//	        Attribute("expression", String, "Mathematical expression")
//	        Required("request_id", "expression")
//	    })
//	    Result(func() {
//	        Attribute("result", Float64)
//	        Required("result")
//	    })
//	    JSONRPC(func() {
//	        POST("/")
//	        ID("request_id") // Use "request_id" instead of default "id"
//	    })
//	})
func ID(name string) {
	endpoint, ok := eval.Current().(*expr.HTTPEndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	endpoint.IDAttribute = name
}

// Notification indicates that the method is a notification and does not
// expect a response.
//
// Notification must appear in a JSONRPC expression within a Method.
//
// Example:
//
//	Method("notify", func() {
//	    Payload(func() {
//	        Attribute("message", String, "Notification message")
//	        Required("message")
//	    })
//	    JSONRPC(func() {
//	        Notification() // This method is a notification and does not expect a response
//	    })
//	})
func Notification() {
	endpoint, ok := eval.Current().(*expr.HTTPEndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	endpoint.IsNotification = true
}
