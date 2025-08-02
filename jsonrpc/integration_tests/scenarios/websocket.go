package scenarios

import (
	"fmt"

	"goa.design/goa/v3/dsl"
)

// createWebSocketDSL creates a DSL function for WebSocket streaming scenarios
// with the specified streaming pattern and data type. The function generates
// services with appropriate streaming methods based on the pattern:
//   - Server streaming: client sends one request, server streams responses
//   - Client streaming: client streams requests, server sends one response
//   - Bidirectional: both client and server stream messages
//
// Each pattern tests different aspects of WebSocket frame handling, connection
// management, and message sequencing in the JSON-RPC transport.
//
// For JSON-RPC streaming, all payloads and results must be objects with
// request ID fields to enable proper message correlation.
func createWebSocketDSL(streamingType StreamingType, dataType DataType) func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("WebSocket Test API")
		})

		// Only define user types when they're actually used
		if dataType == DataTypeUserType || dataType == DataTypeComplex {
			defineTypesForDataType(dataType)
		}

		dsl.Service("streaming", func() {
			dsl.JSONRPC(func() {
				dsl.GET("/jsonrpc/ws") // Use GET for WebSocket endpoint
			})

			switch streamingType {
			case StreamingServer:
				dsl.Method("server_stream", func() {
					// Server streaming: streaming results with request ID
					dsl.StreamingResult(createWebSocketStreamingType(dataType))

					dsl.JSONRPC(func() {
						// Method-level JSONRPC config without GET
					})
				})

			case StreamingClient:
				dsl.Method("client_stream", func() {
					// Client streaming: streaming payload with request ID
					dsl.StreamingPayload(createWebSocketStreamingType(dataType))
					dsl.Result(dsl.String) // Simple acknowledgment

					dsl.JSONRPC(func() {
						// Method-level JSONRPC config without GET
					})
				})

			case StreamingBidirectional:
				dsl.Method("bidirectional_stream", func() {
					// Bidirectional: both payload and result with request IDs
					dsl.StreamingPayload(createWebSocketStreamingType(dataType))
					dsl.StreamingResult(createWebSocketStreamingType(dataType))

					dsl.JSONRPC(func() {
						// Method-level JSONRPC config without GET
					})
				})
			}
		})
	}
}

// createWebSocketStreamingType creates object types with request ID metadata
// required for JSON-RPC streaming. All streaming types must be objects with
// an ID field that has "jsonrpc:id" metadata for request correlation.
func createWebSocketStreamingType(dataType DataType) func() {
	return func() {
		// All JSON-RPC streaming types must have a request ID field
		dsl.Attribute("id", dsl.String, func() {
			dsl.Meta("jsonrpc:id")
		})

		// Add data-specific fields based on the type
		switch dataType {
		case DataTypePrimitive:
			dsl.Attribute("data", dsl.String)
			dsl.Required("id", "data")

		case DataTypeArray:
			dsl.Attribute("items", dsl.ArrayOf(dsl.String))
			dsl.Required("id", "items")

		case DataTypeObject:
			dsl.Attribute("field1", dsl.String)
			dsl.Attribute("field2", dsl.Int)
			dsl.Attribute("field3", dsl.Boolean)
			dsl.Required("id", "field1")

		case DataTypeUserType:
			dsl.Attribute("user_id", dsl.String)
			dsl.Attribute("name", dsl.String)
			dsl.Attribute("email", dsl.String)
			dsl.Required("id", "user_id", "name")

		case DataTypeComplex:
			dsl.Attribute("sequence", dsl.Int)
			dsl.Attribute("data", dsl.MapOf(dsl.String, dsl.Any))
			dsl.Attribute("metadata", func() {
				dsl.Attribute("index", dsl.Int)
				dsl.Attribute("type", dsl.String)
			})
			dsl.Required("id", "sequence")

		default:
			dsl.Attribute("data", dsl.String)
			dsl.Required("id", "data")
		}
	}
}

// createWebSocketRequests creates test requests for WebSocket scenarios based
// on the streaming pattern and data type. The requests include sequences of
// send/receive operations that exercise the streaming functionality.
//
// For server streaming, the client sends a trigger and receives multiple messages.
// For client streaming, the client sends multiple messages and receives an acknowledgment.
// For bidirectional streaming, both send and receive operations are interleaved
// to test concurrent message handling.
func createWebSocketRequests(streamingType StreamingType, dataType DataType) []TestRequest {
	switch streamingType {
	case StreamingServer:
		return []TestRequest{
			{
				Method: "server_stream", // JSON-RPC method name without service prefix
				// No params for server streaming
				StreamingMessages: []StreamMessage{
					{Direction: DirectionReceive, Data: createStreamData(dataType, 1)},
					{Direction: DirectionReceive, Data: createStreamData(dataType, 2)},
					{Direction: DirectionReceive, Data: createStreamData(dataType, 3)},
				},
			},
		}

	case StreamingClient:
		return []TestRequest{
			{
				Method: "client_stream", // JSON-RPC method name without service prefix
				StreamingMessages: []StreamMessage{
					{Direction: DirectionSend, Data: createStreamData(dataType, 1), Delay: 10},
					{Direction: DirectionSend, Data: createStreamData(dataType, 2), Delay: 10},
					{Direction: DirectionSend, Data: createStreamData(dataType, 3), Delay: 10},
				},
				ExpectedResult: "received 3 messages",
			},
		}

	case StreamingBidirectional:
		return []TestRequest{
			{
				Method: "bidirectional_stream", // JSON-RPC method name without service prefix
				StreamingMessages: []StreamMessage{
					{Direction: DirectionSend, Data: createStreamData(dataType, 1)},
					{Direction: DirectionReceive, Data: createStreamData(dataType, 1)},
					{Direction: DirectionSend, Data: createStreamData(dataType, 2), Delay: 10},
					{Direction: DirectionReceive, Data: createStreamData(dataType, 2)},
					{Direction: DirectionSend, Data: createStreamData(dataType, 3), Delay: 10},
					{Direction: DirectionReceive, Data: createStreamData(dataType, 3)},
				},
			},
		}

	default:
		return nil
	}
}

// createStreamData creates streaming data for the given type and index,
// generating unique messages for each position in the stream. The index
// parameter ensures each message is distinguishable, which helps verify
// message ordering and detect dropped or duplicated messages.
//
// The generated data matches the JSON-RPC streaming DSL structure with
// ID attributes for request tracking and proper object format.
func createStreamData(dataType DataType, index int) any {
	// All JSON-RPC streaming data must include an ID for request tracking
	baseID := fmt.Sprintf("req-%d", index)

	switch dataType {
	case DataTypePrimitive:
		return map[string]any{
			"id":   baseID,
			"data": fmt.Sprintf("message %d", index),
		}

	case DataTypeArray:
		return map[string]any{
			"id":    baseID,
			"items": []string{fmt.Sprintf("item%d-1", index), fmt.Sprintf("item%d-2", index)},
		}

	case DataTypeObject:
		return map[string]any{
			"id":     baseID,
			"field1": fmt.Sprintf("Message %d", index),
			"field2": index,
			"field3": index%2 == 0,
		}

	case DataTypeUserType:
		return map[string]any{
			"id":      baseID,
			"user_id": fmt.Sprintf("user%d", index),
			"name":    fmt.Sprintf("Stream User %d", index),
			"email":   fmt.Sprintf("stream%d@example.com", index),
		}

	case DataTypeComplex:
		return map[string]any{
			"id":       baseID,
			"sequence": index,
			"data": map[string]any{
				"value": fmt.Sprintf("complex-%d", index),
			},
			"metadata": map[string]any{
				"index": index,
				"type":  "stream",
			},
		}

	default:
		return map[string]any{
			"id":   baseID,
			"data": fmt.Sprintf("data-%d", index),
		}
	}
}

// defineTypesForDataType defines necessary Goa types based on the data type
// enum value. This centralizes type definitions that are shared across
// different transport scenarios, ensuring consistency in type structures.
//
// The function is called during DSL generation to register user-defined
// and complex types before they're referenced in method definitions. This
// avoids duplication and ensures all scenarios use the same type definitions.
func defineTypesForDataType(dataType DataType) {
	switch dataType {
	case DataTypeUserType:
		dsl.Type("UserType", func() {
			dsl.Attribute("id", dsl.String)
			dsl.Attribute("name", dsl.String)
			dsl.Attribute("email", dsl.String)
			dsl.Required("id", "name")
		})

	case DataTypeComplex:
		dsl.Type("ComplexType", func() {
			dsl.Attribute("sequence", dsl.Int)
			dsl.Attribute("data", dsl.MapOf(dsl.String, dsl.Any))
			dsl.Attribute("metadata", func() {
				dsl.Attribute("index", dsl.Int)
				dsl.Attribute("type", dsl.String)
			})
			dsl.Required("sequence")
		})
	}
}
