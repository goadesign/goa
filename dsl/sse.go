package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// ServerSentEvents specifies that a streaming endpoint should use the
// Server-Sent Events protocol for streaming instead of WebSockets.
//
// SSE is only suitable for server-to-client streaming. Methods using SSE
// typically use POST endpoints. When multiple SSE endpoints exist in a service,
// each should have a unique path to avoid conflicts.
//
// It can be used in four ways:
//
//  1. ServerSentEvents(): StreamingResult type is used directly as the event
//     "data" field (serialized into JSON if not a primitive type)
//  2. ServerSentEvents("attributeName"): The specified attribute is used as the
//     event "data" field (serialized into JSON if not a primitive type)
//  3. ServerSentEvents(func() { ... }): Custom mapping of attributes to event
//     fields
//  4. ServerSentEvents("attributeName", func() { ... }): Define attribute name
//     used as the "data" field and custom mapping for others.
//
// ServerSentEvents can appear in an API HTTP or JSONRPC expression (to specify
// SSE for all streaming methods in the API), in a Service HTTP or JSONRPC
// expression (to specify SSE for all streaming methods in the service), or in a
// Method HTTP or JSONRPC expression. When specified at the API or service
// level, any method with a StreamingPayload will fall back to using WebSockets
// as SSE only supports server-to-client streaming.
//
// See SSEEventData, SSEEventID, SSEEventType, SSEEventRetry for more details on
// mapping result attributes to event fields. See SSERequestID for more details on
// mapping payload attributes to the Last-Event-ID request header.
//
// Example:
//
//	var Notification = Type("Notification", func() {
//	    Attribute("message", String)
//	    Attribute("timestamp", String)
//	    Required("message", "timestamp")
//	})
//
//	var _ = Service("events", func() {
//	    HTTP(func() {
//	        ServerSentEvents() // All streaming methods in this service use SSE by default
//	    })
//
//	    // Simple method with just data field
//	    Method("stream", func() {
//	        StreamingResult(Notification)
//	        HTTP(func() {
//	            GET("/events") // Messages are sent as {"data": {"message": <message>, "timestamp": <timestamp>}}
//	        })
//	    })
//	})
//
//	var _ = Service("other", func() {
//	    // Method using WebSockets
//	    Method("stream", func() {
//	        StreamingResult(Notification)
//	        HTTP(func() {
//	            GET("/websocket")
//	        })
//	    })
//
//	    // Method using SSE with custom event mapping
//	    Method("stream", func() {
//	        Payload(func() {
//	            Attribute("id", String)
//	        })
//	        StreamingResult(Notification)
//	        HTTP(func() {
//	            POST("/events")
//	            ServerSentEvents(func() {
//	                SSERequestID("id")      // Map payload "id" to Last-Event-Id header
//	                SSEEventID("timestamp") // Use result "timestamp" for event ID
//	                SSEEventData("message") // Use result "message" for event data
//	            })
//	            // Events are sent as: id: <timestamp>\ndata: <message>\n\n
//	        })
//	    })
//	})
func ServerSentEvents(args ...any) {
	if len(args) > 2 {
		eval.TooManyArgError()
		return
	}

	var fn func()
	var dataField string
	if len(args) > 0 {
		switch actual := args[0].(type) {
		case func():
			fn = actual
		case string:
			dataField = actual
		case nil:
			// Use the entire result as data field
		default:
			eval.InvalidArgError("function or string", args[0])
			return
		}
		if len(args) == 2 {
			if fn != nil {
				eval.TooManyArgError()
				return
			}
			var ok bool
			fn, ok = args[1].(func())
			if !ok {
				eval.InvalidArgError("function", args[1])
				return
			}
		}
	}

	sse := &expr.HTTPSSEExpr{
		DataField: dataField,
	}

	switch actual := eval.Current().(type) {
	case *expr.HTTPExpr:
		actual.SSE = sse
	case *expr.HTTPServiceExpr:
		actual.SSE = sse
	case *expr.HTTPEndpointExpr:
		actual.SSE = sse
	case *expr.JSONRPCExpr:
		actual.SSE = sse
	default:
		eval.IncompatibleDSL()
	}

	if fn != nil {
		eval.Execute(fn, sse)
	}
}

// SSERequestID defines the attribute of the Payload type that provides the
// Last-Event-ID request header value. The attribute must exist in the Payload
// type and must be of type String.
//
// SSERequestID must appear in a `ServerSentEvents` expression.
//
// SSERequestID accepts a single argument: the name of the attribute of the
// Payload type that provides the Last-Event-ID request header value.
//
// Example:
//
//	Method("stream", func() {
//	    Payload(func() {
//	        Attribute("id", String)
//	    })
//	    StreamingResult(Notification)
//	    HTTP(func() {
//	        GET("/events")
//	        ServerSentEvents(func() {   // Use SSE for this method
//	            SSERequestID("id")      // Use payload "id" field to set "Last-Event-Id" request header
//	            SSEEventID("timestamp") // Use result "timestamp" attribute for "id" event field
//	            SSEEventData("message") // Use result "message" attribute for "data" event field
//	        })
//	    })
//	})
func SSERequestID(name string) {
	if name == "" {
		eval.ReportError("request ID field name cannot be empty")
		return
	}
	sse, ok := eval.Current().(*expr.HTTPSSEExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	sse.RequestIDField = name
}

// SSEEventData defines the attribute of the StreamingResult type that provides the
// data field for a Server-Sent Event. The attribute must exist in the
// StreamingResult type.
//
// SSEEventData must appear in a `ServerSentEvents` expression.
//
// SSEEventData accepts a single argument: the name of the attribute of the
// StreamingResult type that provides the data field for a Server-Sent Event.
//
// Example:
//
//	Method("stream", func() {
//	    StreamingResult(Payload)
//	    HTTP(func() {
//	        GET("/events")
//	        ServerSentEvents(func() {
//	            SSEEventData("message") // Use payload "message" attribute for SSE data field, other attributes are ignored
//	        })
//	    })
//	})
func SSEEventData(name string) {
	if name == "" {
		eval.ReportError("data field name cannot be empty")
		return
	}
	sse, ok := eval.Current().(*expr.HTTPSSEExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	sse.DataField = name
}

// SSEEventID defines the attribute of the StreamingResult type that provides the
// id field for a Server-Sent Event. The attribute must exist in the
// StreamingResult type and must be of type String.
//
// SSEEventID must appear in a `ServerSentEvents` expression.
//
// SSEEventID accepts a single argument: the name of the attribute of the
// StreamingResult type that provides the id field for a Server-Sent Event.
//
// Example:
//
//	Method("stream", func() {
//	    StreamingResult(Payload)
//	    HTTP(func() {
//	        GET("/events")
//	        ServerSentEvents(func() {
//	            SSEEventID("timestamp") // Use "timestamp" attribute for SSE id field
//	        })
//	    })
//	})
func SSEEventID(name string) {
	if name == "" {
		eval.ReportError("id field name cannot be empty")
		return
	}
	sse, ok := eval.Current().(*expr.HTTPSSEExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	sse.IDField = name
}

// SSEEventType defines the attribute of the StreamingResult type that provides the
// event field (event type) for a Server-Sent Event. The attribute must exist in the
// StreamingResult type and must be of type String.
//
// SSEEventType must appear in a `ServerSentEvents` expression.
//
// SSEEventType accepts a single argument: the name of the attribute of the
// StreamingResult type that provides the event field for a Server-Sent Event.
//
// Example:
//
//	Method("stream", func() {
//	    StreamingResult(Payload)
//	    HTTP(func() {
//	        GET("/events")
//	        ServerSentEvents(func() {
//	            SSEEventType("type") // Use payload "type" attribute for SSE event field
//	        })
//	    })
//	})
func SSEEventType(name string) {
	if name == "" {
		eval.ReportError("event field name cannot be empty")
		return
	}
	sse, ok := eval.Current().(*expr.HTTPSSEExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	sse.EventField = name
}

// SSEEventRetry defines the attribute of the StreamingResult type that provides
// the retry field for a Server-Sent Event. The attribute must exist in the
// StreamingResult type and must be of type Int or UInt.
//
// SSEEventRetry must appear in a `ServerSentEvents` expression.
//
// SSEEventRetry accepts a single argument: the name of the attribute of the
// StreamingResult type that provides the retry field for a Server-Sent Event.
//
// Example:
//
//	Method("stream", func() {
//	    StreamingResult(Notification)
//	    HTTP(func() {
//	        GET("/events")
//	        ServerSentEvents(func() {
//	            SSEEventRetry("retry") // Use "retry" attribute for SSE retry field
//	        })
//	    })
//	})
func SSEEventRetry(name string) {
	if name == "" {
		eval.ReportError("retry field name cannot be empty")
		return
	}
	sse, ok := eval.Current().(*expr.HTTPSSEExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	sse.RetryField = name
}
