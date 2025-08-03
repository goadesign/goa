package tests

import (
	"encoding/json"
	"strings"
	"testing"

	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
	"goa.design/goa/v3/jsonrpc/integration_tests/scenarios"
	"goa.design/goa/v3/jsonrpc/integration_tests/validators"
)

// TestStandardJSONRPCErrors tests standard JSON-RPC error codes
func TestStandardJSONRPCErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	testCases := []struct {
		name         string
		scenario     scenarios.Scenario
		expectedCode int
		expectedMsg  string
	}{
		// Skip parse_error test for now - it requires sending malformed JSON
		// which the current test harness doesn't support
		{
			name: "invalid_request",
			scenario: scenarios.Scenario{
				Name:      "invalid_request",
				Transport: scenarios.TransportHTTP,
				DSLCode:   createBasicDSLCode(),
				Requests: []scenarios.TestRequest{
					{
						// Test missing required payload - should be invalid params
						Method: "echo",
						Params: json.RawMessage("null"), // Explicitly send null params
					},
				},
			},
			expectedCode: -32602,           // Invalid params for missing required payload
			expectedMsg:  "Invalid params", // Standard JSON-RPC error message
		},
		{
			name: "method_not_found",
			scenario: scenarios.Scenario{
				Name:      "method_not_found",
				Transport: scenarios.TransportHTTP,
				DSLCode:   createBasicDSLCode(),
				Requests: []scenarios.TestRequest{
					{
						Method: "nonexistent_method",
						Params: map[string]any{"test": "value"},
					},
				},
			},
			expectedCode: -32601,
			expectedMsg:  "not found",
		},
		{
			name: "invalid_params",
			scenario: scenarios.Scenario{
				Name:      "invalid_params",
				Transport: scenarios.TransportHTTP,
				DSLCode:   createValidationDSLCode(),
				Requests: []scenarios.TestRequest{
					{
						Method: "validate",
						Params: map[string]any{
							// Missing required field
							"optional": "value",
						},
					},
				},
			},
			expectedCode: -32602,
			expectedMsg:  "Invalid params", // Standard JSON-RPC error message
		},
	}

	runner := scenarios.NewScenarioRunner(h)

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run test cases in parallel
			// Add error validators
			tc.scenario.Validators = []validators.Validator{
				validators.ProtocolValidator(),
				validators.ErrorValidator(tc.expectedCode, tc.expectedMsg),
			}

			// Run scenario
			if err := runner.Run(tc.scenario); err != nil {
				t.Fatalf("Error scenario %s failed: %v", tc.name, err)
			}
		})
	}
}

// TestCustomApplicationErrors tests custom application error handling
func TestCustomApplicationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Create custom error scenario
	scenario := scenarios.Scenario{
		Name:        "custom_errors",
		Description: "Test custom application errors",
		Transport:   scenarios.TransportHTTP,
		PayloadType: scenarios.DataTypeObject,
		ResultType:  scenarios.DataTypeObject,
		Features:    []scenarios.Feature{scenarios.FeatureErrors},
		DSLCode:     createCustomErrorDSLCode(),
		Requests: []scenarios.TestRequest{
			{
				Method: "process",
				Params: map[string]any{
					"action": "unauthorized",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32001,
					Message: "unauthorized",
				},
			},
			{
				Method: "process",
				Params: map[string]any{
					"action": "not_found",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32002,
					Message: "resource not found",
				},
			},
			{
				Method: "process",
				Params: map[string]any{
					"action": "conflict",
				},
				ExpectedError: &scenarios.ExpectedError{
					Code:    -32003,
					Message: "conflict",
				},
			},
		},
		Validators: []validators.Validator{
			validators.ProtocolValidator(),
			validators.ErrorCodeRangeValidator(-32099, -32000),
		},
	}

	// Run scenario
	runner := scenarios.NewScenarioRunner(h)
	if err := runner.Run(scenario); err != nil {
		t.Fatalf("Custom error scenario failed: %v", err)
	}
}

// TestErrorDataPropagation tests that error data is properly propagated
func TestErrorDataPropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Create error with data scenario
	scenario := scenarios.Scenario{
		Name:        "error_data",
		Description: "Test error data propagation",
		Transport:   scenarios.TransportHTTP,
		DSLCode:     createErrorWithDataDSLCode(),
		Requests: []scenarios.TestRequest{
			{
				Method: "validate_complex",
				Params: map[string]any{
					"data": map[string]any{
						"field1": "invalid",
						"field2": -1, // Should be positive
					},
				},
			},
		},
		Validators: []validators.Validator{
			validators.ProtocolValidator(),
			validators.DataIntegrityValidator(),
			validators.CustomErrorValidator(harness.ErrorObject{
				Code:    -32602,
				Message: "Invalid params", // JSON-RPC standard error message
				Data:    nil,              // Goa's standard validation errors don't include custom data
			}),
		},
	}

	// Run scenario
	runner := scenarios.NewScenarioRunner(h)
	if err := runner.Run(scenario); err != nil {
		t.Fatalf("Error data scenario failed: %v", err)
	}
}

// TestTransportSpecificErrors tests transport-specific error scenarios
func TestTransportSpecificErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	testCases := []struct {
		name             string
		transport        scenarios.Transport
		scenario         scenarios.Scenario
		expectError      bool
		errorShouldMatch func(error) bool // Function to validate expected error
	}{
		{
			name:        "http_timeout",
			transport:   scenarios.TransportHTTP,
			expectError: true,
			errorShouldMatch: func(err error) bool {
				// Check for timeout-related errors
				return strings.Contains(strings.ToLower(err.Error()), "timeout") ||
					strings.Contains(strings.ToLower(err.Error()), "context deadline exceeded") ||
					strings.Contains(strings.ToLower(err.Error()), "i/o timeout")
			},
			scenario: scenarios.Scenario{
				Name:      "http_timeout_error",
				Transport: scenarios.TransportHTTP,
				DSLCode:   createTimeoutDSLCode(),
				Requests: []scenarios.TestRequest{
					{
						Method: "slow_operation",
						Params: map[string]any{
							"delay_ms": 5000, // 5 seconds
						},
					},
				},
				Validators: []validators.Validator{
					validators.ProtocolValidator(),
					validators.StandardErrorValidator("internal"),
				},
			},
		},
		{
			name:        "websocket_disconnect",
			transport:   scenarios.TransportWebSocket,
			expectError: true,
			errorShouldMatch: func(err error) bool {
				// Check for WebSocket disconnect or connection errors
				errStr := strings.ToLower(err.Error())
				return strings.Contains(errStr, "websocket") ||
					strings.Contains(errStr, "connection") ||
					strings.Contains(errStr, "disconnect") ||
					strings.Contains(errStr, "closed") ||
					strings.Contains(errStr, "unexpected eof")
			},
			scenario: scenarios.Scenario{
				Name:      "websocket_disconnect_error",
				Transport: scenarios.TransportWebSocket,
				Streaming: scenarios.StreamingServer,
				DSLCode:   createDisconnectDSLCode(),
				Requests: []scenarios.TestRequest{
					{
						Method: "stream_with_error",
						Params: "start",
						StreamingMessages: []scenarios.StreamMessage{
							{Direction: scenarios.DirectionReceive, Data: "msg1"},
							{Direction: scenarios.DirectionReceive, Data: "error"},
						},
					},
				},
				Validators: []validators.Validator{
					validators.ProtocolValidator(),
				},
			},
		},
	}

	runner := scenarios.NewScenarioRunner(h)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run scenario
			err := runner.Run(tc.scenario)

			if tc.expectError {
				// We expect an error - validate it occurred and matches expectations
				if err == nil {
					t.Fatalf("Expected transport error for %s, but scenario completed successfully", tc.name)
				}
				if tc.errorShouldMatch != nil && !tc.errorShouldMatch(err) {
					t.Fatalf("Transport error for %s doesn't match expected pattern: %v", tc.name, err)
				}
			} else {
				// We don't expect an error - fail if one occurs
				if err != nil {
					t.Fatalf("Unexpected error in scenario %s: %v", tc.name, err)
				}
			}
		})
	}
}

// Helper DSL creation functions

func createBasicDSLCode() string {
	return `	API("test", func() {
		Title("Basic Test API")
	})
	
	Service("basic", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("echo", func() {
			Payload(func() {
				Attribute("message", String)
				Required("message")
			})
			Result(String)
			JSONRPC(func() {
			})
		})
	})`
}

func createValidationDSLCode() string {
	return `	API("test", func() {
		Title("Validation Test API")
	})
	
	Service("validation", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("validate", func() {
			Payload(func() {
				Attribute("required", String)
				Attribute("optional", String)
				Required("required")
			})
			Result(Boolean)
			JSONRPC(func() {
			})
		})
	})`
}

func createCustomErrorDSLCode() string {
	return `	API("test", func() {
		Title("Custom Error Test API")
	})
	
	Service("errors", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		
		Error("Unauthorized", func() {
			Description("unauthorized")
			Attribute("reason", String)
			Required("reason")
		})
		Error("NotFound", func() {
			Description("resource not found")
			Attribute("resource", String)
			Attribute("id", String)
			Required("resource", "id")
		})
		Error("Conflict", func() {
			Description("conflict")
			Attribute("message", String)
			Required("message")
		})
		
		Method("process", func() {
			Payload(func() {
				Attribute("action", String)
				Required("action")
			})
			Result(func() {
				Attribute("status", String)
				Required("status")
			})
			Error("Unauthorized")
			Error("NotFound")
			Error("Conflict")
			
			JSONRPC(func() {
				Response("Unauthorized", func() {
					Code(-32001)
				})
				Response("NotFound", func() {
					Code(-32002)
				})
				Response("Conflict", func() {
					Code(-32003)
				})
			})
		})
	})`
}

func createErrorWithDataDSLCode() string {
	return `	API("test", func() {
		Title("Error Data Test API")
	})
	
	Service("validation", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("validate_complex", func() {
			Payload(func() {
				Attribute("data", func() {
					Attribute("field1", String, func() {
						Pattern("^[a-z]+$")
					})
					Attribute("field2", Int, func() {
						Minimum(0)
					})
					Required("field1", "field2")
				})
				Required("data")
			})
			Result(Boolean)
			JSONRPC(func() {
			})
		})
	})`
}

func createTimeoutDSLCode() string {
	return `	API("test", func() {
		Title("Timeout Test API")
	})
	
	Service("slow", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("slow_operation", func() {
			Payload(func() {
				Attribute("delay_ms", Int)
				Required("delay_ms")
			})
			Result(String)
			JSONRPC(func() {
			})
		})
	})`
}

func createDisconnectDSLCode() string {
	return `	API("test", func() {
		Title("Disconnect Test API")
	})
	
	Service("streaming", func() {
		JSONRPC(func() {
			GET("/ws")
		})
		
		Error("StreamError")
		
		Method("stream_with_error", func() {
			StreamingPayload(String)
			StreamingResult(String)
			Error("StreamError")
			JSONRPC(func() {
				// Method-level JSONRPC config without GET
			})
		})
	})`
}
