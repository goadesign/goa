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
// a single HTTP endpoint and must use the same transport (HTTP, WebSocket or SSE).
//
// JSONRPC can be used at three levels:
//
//   - At the API level: JSONRPC maps service errors to standard JSON-RPC error
//     codes.
//   - At the service level: JSONRPC sets the HTTP endpoint path for all
//     JSON-RPC methods in the service and allows you to define common errors
//     and their error code mappings.
//   - At the method level: JSONRPC configures how the request and response "id"
//     fields are mapped to payload/result attributes, specifies whether the
//     method is a notification (no "id" field), and allows you to define
//     method-specific error code mappings.
//
// Request Handling:
//
// The generated code decodes the JSON-RPC "params" field into the method
// payload and the "id" field to the payload attribute specified by the ID
// function.  For non-streaming methods, if the result's ID attribute is not
// already set, the generated code automatically copies the request ID from the
// payload to the result's ID attribute.
//
// The generated code fully supports batch JSON-RPC requests: when the HTTP
// request body contains an array of JSON-RPC request objects, it will unmarshal
// the array, process each request independently (including error handling and
// notifications), and marshal the responses into a single array of JSON-RPC
// response objects in the HTTP response body.
//
// Streaming:
//
// Methods using StreamingResult() support either Server-Sent Events or WebSockets.
// With SSE, each result element is sent as a JSON-RPC response in a separate event.
// With WebSockets, methods can use StreamingPayload() for bidirectional streaming,
// where each payload/result element is sent as a complete JSON-RPC message.
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
//	    Error("timeout", ErrTimeout, "Request timed out") // ErrTimeout must have a limit attribute
//
//	    JSONRPC(func() {
//	        POST("/rpc")                                    // All methods use this endpoint
//	        Response("timeout", func() {                    // Custom error response
//	            Code(5001)                                  // Application error code
//	        })
//	    })
//
//	    Method("notify", func() { // Notification method (no ID mapping)
//	        Payload(func() {
//	            Attribute("message", String, "Notification message")
//	            Required("message")
//	        })
//	        JSONRPC(func() {
//	            Notification() // This method is a notification and does not expect a response
//	        })
//	    })
//
//	    Method("divide", func() { // Request/response method
//	        Payload(func() {
//	            ID("req_id")                           // Map request ID to payload field
//	            Attribute("dividend", Int, "Dividend")
//	            Attribute("divisor", Int, "Divisor")
//	            Required("dividend", "divisor")
//	        })
//	        Result(func() {
//	            ID("req_id")                           // Map request ID to result field
//	            Attribute("result", Float64)
//	        })
//	        Error("div_zero", ErrorResult, "Division by zero")
//	        JSONRPC(func() {
//	            Response("div_zero", RPCInvalidParams) // Map div_zero error to JSON-RPC code
//	        })
//	    })
//
//	    Method("updates", func() {                    // SSE streaming method
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
	case *expr.MethodExpr:
		svc := expr.Root.API.JSONRPC.ServiceFor(actual.Service, &expr.Root.API.JSONRPC.HTTPExpr)
		e := svc.EndpointFor(actual)
		if e.Meta == nil {
			e.Meta = expr.MetaExpr{}
		}
		e.Meta["jsonrpc"] = nil
		e.DSLFunc = dsl
		r := &expr.RouteExpr{Method: "POST", Path: "/", Endpoint: e}
		e.Routes = []*expr.RouteExpr{r}
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
