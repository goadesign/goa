package scenarios

import (
	"fmt"
)

// GenerateTestMatrix creates a comprehensive test matrix covering all meaningful
// combinations of transports, data types, and features. The matrix ensures
// thorough coverage of the JSON-RPC implementation by systematically testing:
//   - All transport types (HTTP, WebSocket, SSE)
//   - All payload and result type combinations
//   - Special scenarios (errors, validation, batch requests, views)
//
// The generated scenarios can be filtered and executed selectively by test
// functions based on transport type, features, or other criteria.
func GenerateTestMatrix() []Scenario {
	var scenarios []Scenario

	// Generate HTTP scenarios
	scenarios = append(scenarios, generateHTTPScenarios()...)

	// Generate WebSocket scenarios
	scenarios = append(scenarios, generateWebSocketScenarios()...)

	// Generate SSE scenarios
	scenarios = append(scenarios, generateSSEScenarios()...)

	// Add special scenarios
	scenarios = append(scenarios, generateSpecialScenarios()...)

	return scenarios
}

// generateHTTPScenarios creates all HTTP transport test scenarios by
// systematically combining different payload and result types. This ensures
// comprehensive coverage of type marshaling and unmarshaling across the
// JSON-RPC/HTTP transport.
//
// The function generates scenarios for all payload/result combinations plus
// special notification scenarios (no result). Each scenario includes the
// appropriate DSL function and test requests for validation.
func generateHTTPScenarios() []Scenario {
	var scenarios []Scenario

	// Test all payload/result type combinations
	payloadTypes := []DataType{
		DataTypeNone,
		DataTypePrimitive,
		DataTypeArray,
		DataTypeObject,
		DataTypeMap,
		DataTypeUserType,
	}

	resultTypes := []DataType{
		DataTypePrimitive,
		DataTypeArray,
		DataTypeObject,
		DataTypeMap,
		DataTypeUserType,
	}

	for _, pt := range payloadTypes {
		for _, rt := range resultTypes {
			scenario := Scenario{
				Name:        fmt.Sprintf("http_%s_payload_%s_result", pt, rt),
				Description: fmt.Sprintf("HTTP transport with %s payload and %s result", pt, rt),
				Transport:   TransportHTTP,
				PayloadType: pt,
				ResultType:  rt,
				Streaming:   StreamingNone,
				Features:    []Feature{FeatureCore},
				DSLCode:     GenerateDSLCode(pt, rt),
				Requests:    createHTTPRequests(pt, rt),
			}

			scenarios = append(scenarios, scenario)
		}
	}

	// Add notification scenarios (no result)
	for _, pt := range payloadTypes[1:] { // Skip "none" payload
		scenario := Scenario{
			Name:        fmt.Sprintf("http_%s_notification", pt),
			Description: fmt.Sprintf("HTTP notification with %s payload", pt),
			Transport:   TransportHTTP,
			PayloadType: pt,
			ResultType:  DataTypeNone,
			Streaming:   StreamingNone,
			Features:    []Feature{FeatureCore},
			DSLCode:     GenerateNotificationDSL(pt),
			Requests:    createNotificationRequests(pt),
		}

		scenarios = append(scenarios, scenario)
	}

	return scenarios
}

// generateWebSocketScenarios creates WebSocket streaming test scenarios
// covering server streaming, client streaming, and bidirectional streaming
// patterns. Each pattern is tested with various data types to ensure proper
// handling of streaming frames and message sequencing.
//
// The scenarios test the full lifecycle of WebSocket connections including
// connection establishment, message exchange, and graceful closure.
func generateWebSocketScenarios() []Scenario {
	var scenarios []Scenario

	streamingTypes := []StreamingType{
		StreamingServer,
		StreamingClient,
		StreamingBidirectional,
	}

	dataTypes := []DataType{
		DataTypePrimitive,
		DataTypeArray,
		DataTypeObject,
		DataTypeUserType,
		DataTypeComplex,
	}

	for _, st := range streamingTypes {
		for _, dt := range dataTypes {
			var payloadType, resultType DataType

			switch st {
			case StreamingServer:
				payloadType = DataTypeNone // Server streaming has no payload
				resultType = dt
			case StreamingClient:
				payloadType = dt
				resultType = DataTypePrimitive // Simple acknowledgment
			case StreamingBidirectional:
				payloadType = dt
				resultType = dt
			}

			scenario := Scenario{
				Name:        fmt.Sprintf("websocket_%s_%s", st, dt),
				Description: fmt.Sprintf("WebSocket %s streaming with %s data", st, dt),
				Transport:   TransportWebSocket,
				PayloadType: payloadType,
				ResultType:  resultType,
				Streaming:   st,
				Features:    []Feature{FeatureStreaming},
				DSLCode:     GenerateWebSocketDSL(payloadType, resultType, st),
				Requests:    createWebSocketRequests(st, dt),
			}

			scenarios = append(scenarios, scenario)
		}
	}

	return scenarios
}

// generateSSEScenarios creates Server-Sent Events test scenarios
func generateSSEScenarios() []Scenario {
	var scenarios []Scenario

	// SSE only supports server streaming with no payload (GET request)
	resultTypes := []DataType{
		DataTypePrimitive,
		DataTypeArray,
		DataTypeObject,
		DataTypeUserType,
		DataTypeComplex,
	}

	for _, rt := range resultTypes {
		scenario := Scenario{
			Name:        fmt.Sprintf("sse_%s_result", rt),
			Description: fmt.Sprintf("SSE streaming with %s result stream", rt),
			Transport:   TransportSSE,
			PayloadType: DataTypeNone,
			ResultType:  rt,
			Streaming:   StreamingServer,
			Features:    []Feature{FeatureStreaming},
			DSLCode:     GenerateSSEDSL(DataTypeNone, rt),
			Requests:    createSSERequests(DataTypeNone, rt),
		}

		scenarios = append(scenarios, scenario)
	}

	return scenarios
}

// generateSpecialScenarios creates scenarios for specific features
func generateSpecialScenarios() []Scenario {
	return []Scenario{
		// Error handling scenarios
		{
			Name:        "http_error_standard",
			Description: "Standard JSON-RPC error codes",
			Transport:   TransportHTTP,
			PayloadType: DataTypePrimitive,
			ResultType:  DataTypePrimitive,
			Features:    []Feature{FeatureErrors},
			DSLCode:     GenerateErrorDSL(false),
			Requests:    createErrorRequests(false),
		},
		{
			Name:        "http_error_custom",
			Description: "Custom application errors",
			Transport:   TransportHTTP,
			PayloadType: DataTypeObject,
			ResultType:  DataTypeObject,
			Features:    []Feature{FeatureErrors},
			DSLCode:     GenerateErrorDSL(true),
			Requests:    createErrorRequests(true),
		},

		// Validation scenarios
		{
			Name:        "http_validation_required",
			Description: "Required field validation",
			Transport:   TransportHTTP,
			PayloadType: DataTypeObject,
			ResultType:  DataTypeObject,
			Features:    []Feature{FeatureValidation},
			DSLCode:     GenerateValidationDSL("required"),
			Requests:    createValidationRequests("required"),
		},
		{
			Name:        "http_validation_format",
			Description: "Format validation (email, url, etc)",
			Transport:   TransportHTTP,
			PayloadType: DataTypeObject,
			ResultType:  DataTypeObject,
			Features:    []Feature{FeatureValidation},
			DSLCode:     GenerateValidationDSL("format"),
			Requests:    createValidationRequests("format"),
		},

		// Batch request scenario
		{
			Name:        "http_batch_requests",
			Description: "Batch JSON-RPC requests",
			Transport:   TransportHTTP,
			PayloadType: DataTypeArray,
			ResultType:  DataTypeArray,
			Features:    []Feature{FeatureBatch},
			DSLCode:     GenerateBatchDSL(),
			Requests:    createBatchRequests(),
		},

		// Views scenario
		{
			Name:        "http_result_views",
			Description: "Result type views",
			Transport:   TransportHTTP,
			PayloadType: DataTypePrimitive,
			ResultType:  DataTypeUserType,
			Features:    []Feature{FeatureViews},
			DSLCode:     GenerateViewsDSL(),
			Requests:    createViewsRequests(),
		},

		// Complex nested types
		{
			Name:        "http_deeply_nested",
			Description: "Deeply nested data structures",
			Transport:   TransportHTTP,
			PayloadType: DataTypeComplex,
			ResultType:  DataTypeComplex,
			Features:    []Feature{FeatureCore},
			DSLCode:     GenerateComplexDSL(),
			Requests:    createComplexRequests(),
		},

		// Large payload test
		{
			Name:        "http_large_payload",
			Description: "Large payload handling",
			Transport:   TransportHTTP,
			PayloadType: DataTypeArray,
			ResultType:  DataTypeObject,
			Features:    []Feature{FeatureCore},
			DSLCode:     GenerateLargePayloadDSL(),
			Requests:    createLargePayloadRequests(),
		},

		// Unicode handling
		{
			Name:        "http_unicode",
			Description: "Unicode string handling",
			Transport:   TransportHTTP,
			PayloadType: DataTypeObject,
			ResultType:  DataTypeObject,
			Features:    []Feature{FeatureCore},
			DSLCode:     GenerateUnicodeDSL(),
			Requests:    createUnicodeRequests(),
		},
	}
}

// QuickTestScenarios returns a representative subset of test scenarios suitable
// for quick feedback during development. These scenarios cover the essential
// functionality of each transport type without running the full matrix.
//
// Quick tests typically complete in under 30 seconds and are useful for
// verifying basic functionality before running the comprehensive test suite.
func QuickTestScenarios() []Scenario {
	return []Scenario{
		{
			Name:        "http_basic",
			Description: "Basic HTTP request/response",
			Transport:   TransportHTTP,
			PayloadType: DataTypeObject,
			ResultType:  DataTypeObject,
			Features:    []Feature{FeatureCore},
			DSLCode:     GenerateHTTPDSL(DataTypeObject, DataTypeObject),
			Requests:    createHTTPRequests(DataTypeObject, DataTypeObject),
		},
		{
			Name:        "websocket_basic",
			Description: "Basic WebSocket streaming",
			Transport:   TransportWebSocket,
			PayloadType: DataTypePrimitive,
			ResultType:  DataTypePrimitive,
			Streaming:   StreamingServer,
			Features:    []Feature{FeatureStreaming},
			DSLCode:     GenerateWebSocketDSL(DataTypePrimitive, DataTypePrimitive, StreamingServer),
			Requests:    createWebSocketRequests(StreamingServer, DataTypePrimitive),
		},
		{
			Name:        "sse_basic",
			Description: "Basic SSE streaming",
			Transport:   TransportSSE,
			PayloadType: DataTypeNone,
			ResultType:  DataTypePrimitive,
			Streaming:   StreamingServer,
			Features:    []Feature{FeatureStreaming},
			DSLCode:     GenerateSSEDSL(DataTypeNone, DataTypePrimitive),
			Requests:    createSSERequests(DataTypeNone, DataTypePrimitive),
		},
		{
			Name:        "http_errors",
			Description: "Error handling",
			Transport:   TransportHTTP,
			PayloadType: DataTypePrimitive,
			ResultType:  DataTypePrimitive,
			Features:    []Feature{FeatureErrors},
			DSLCode:     GenerateErrorDSL(false),
			Requests:    createErrorRequests(false),
		},
	}
}
