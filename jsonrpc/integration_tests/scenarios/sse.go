package scenarios

import (
	"goa.design/goa/v3/dsl"
)

// createSSEDSL creates a DSL function for Server-Sent Events scenarios with
// the specified payload and result types. SSE provides a unidirectional stream
// from server to client over HTTP, suitable for real-time updates and notifications.
//
// The generated service includes a subscribe method that optionally accepts
// parameters and streams results to the client. The DSL configures the
// JSON-RPC endpoint with SSE transport semantics.
func createSSEDSL(payloadType, resultType DataType) func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("SSE Test API")
		})

		// Define types if needed
		defineTypesForDataType(resultType)

		dsl.Service("events", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/events")
			})

			dsl.Method("subscribe", func() {
				// Define payload if not none
				if payloadType != DataTypeNone {
					// For array payloads, wrap in object to avoid CLI issues
					if payloadType == DataTypeArray {
						dsl.Payload(func() {
							dsl.Attribute("items", dsl.ArrayOf(dsl.String))
							dsl.Required("items")
						})
					} else {
						dsl.Payload(createTypeExpression(payloadType))
					}
				}

				// SSE always has streaming result
				dsl.StreamingResult(createTypeExpression(resultType))

				dsl.JSONRPC(func() {
					dsl.ServerSentEvents()
				})
			})
		})
	}
}

// createSSERequests creates test requests for SSE scenarios that validate
// server-to-client streaming functionality. The requests establish an SSE
// connection and expect to receive a sequence of events.
//
// Each request includes the initial subscription parameters (if any) and
// a list of expected streaming messages. The test framework validates that
// events are received in order and properly formatted according to the SSE
// specification.
func createSSERequests(payloadType, resultType DataType) []TestRequest {
	var params any
	if payloadType != DataTypeNone {
		params = createTestPayload(payloadType)
	}

	return []TestRequest{
		{
			Method: "subscribe", // Use simple method name from DSL
			Params: params,
			StreamingMessages: []StreamMessage{
				{Direction: DirectionReceive, Data: createSSEData(resultType, 1)},
				{Direction: DirectionReceive, Data: createSSEData(resultType, 2)},
				{Direction: DirectionReceive, Data: createSSEData(resultType, 3)},
				{Direction: DirectionReceive, Data: createSSEData(resultType, 4)},
				{Direction: DirectionReceive, Data: createSSEData(resultType, 5)},
			},
		},
	}
}

// createSSEData creates SSE event data for the specified data type and
// sequence index. This delegates to SSETestData to ensure consistency
// between what tests expect and what servers send.
func createSSEData(dataType DataType, index int) any {
	testData := SSETestData{ResultType: dataType}
	return testData.GenerateData(index)
}

// Error handling DSL creators

// createErrorDSL creates a DSL for error handling scenarios
func createErrorDSL(customErrors bool) func() {
	return func() {
		dsl.API("test", func() {
			dsl.Title("Error Test API")
		})

		dsl.Service("errors", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/jsonrpc")
			})

			if customErrors {
				// Define custom errors
				dsl.Error("ValidationError", func() {
					dsl.Attribute("field", dsl.String)
					dsl.Attribute("message", dsl.String)
					dsl.Required("field", "message")
				})

				dsl.Error("NotFoundError", func() {
					dsl.Attribute("resource", dsl.String)
					dsl.Attribute("id", dsl.String)
					dsl.Required("resource", "id")
				})
			}

			dsl.Method("test_error", func() {
				dsl.Payload(func() {
					dsl.Attribute("trigger", dsl.String)
					dsl.Required("trigger")
				})

				dsl.Result(dsl.String)

				if customErrors {
					dsl.Error("ValidationError")
					dsl.Error("NotFoundError")
				}

				dsl.JSONRPC(func() {
					if customErrors {
						dsl.Response("ValidationError", func() {
							dsl.Code(-32001)
						})
						dsl.Response("NotFoundError", func() {
							dsl.Code(-32002)
						})
					}
				})
			})
		})
	}
}

// createErrorRequests creates test requests for error scenarios
func createErrorRequests(customErrors bool) []TestRequest {
	if customErrors {
		return []TestRequest{
			{
				Method: "test_error", // Use simple method name from DSL
				Params: map[string]any{"trigger": "validation"},
				ExpectedError: &ExpectedError{
					Code:    -32001,
					Message: "validation error",
				},
			},
			{
				Method: "test_error", // Use simple method name from DSL
				Params: map[string]any{"trigger": "notfound"},
				ExpectedError: &ExpectedError{
					Code:    -32002,
					Message: "not found",
				},
			},
			{
				Method:         "test_error", // Use simple method name from DSL
				Params:         map[string]any{"trigger": "success"},
				ExpectedResult: "success",
			},
		}
	}

	// Standard JSON-RPC errors that can be tested at the service level
	return []TestRequest{
		{
			// Test method not found by calling non-existent method
			Method: "nonexistent",
			Params: map[string]any{"trigger": "method"},
			ExpectedError: &ExpectedError{
				Code:    -32601,
				Message: "Method not found",
			},
		},
		{
			// Test invalid params by sending wrong type
			Method: "test_error",
			Params: "not an object", // Should be object with trigger field
			ExpectedError: &ExpectedError{
				Code:    -32602,
				Message: "Invalid params",
			},
		},
		{
			// Test internal error by triggering a generic error
			Method: "test_error",
			Params: map[string]any{"trigger": "internal"},
			ExpectedError: &ExpectedError{
				Code:    -32603,
				Message: "Internal error",
			},
		},
	}
}
