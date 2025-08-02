package tests

import (
	"testing"

	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
	"goa.design/goa/v3/jsonrpc/integration_tests/scenarios"
	"goa.design/goa/v3/jsonrpc/integration_tests/validators"
)

// TestWebSocketServerStreaming tests server-to-client streaming
func TestWebSocketServerStreaming(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	
	// Create test harness
	h := harness.New(t)
	
	// Get WebSocket scenarios
	matrix := scenarios.GenerateTestMatrix()
	
	// Run server streaming scenarios
	runner := scenarios.NewScenarioRunner(h)
	
	for _, scenario := range matrix {
		if scenario.Transport != scenarios.TransportWebSocket {
			continue
		}
		if scenario.Streaming != scenarios.StreamingServer {
			continue
		}
		
		scenario := scenario // capture range variable
		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel() // Run test cases in parallel
			// Add streaming validators
			// Note: StandardValidators expect JSON-RPC responses, but server streaming sends notifications
			// So we only use the streaming message counter
			streamValidator := validators.NewStreamingValidator(3, true)
			scenario.Validators = []validators.Validator{
				streamValidator,
			}
			
			// Run scenario
			if err := runner.Run(scenario); err != nil {
				t.Fatalf("Server streaming scenario failed: %v", err)
			}
			
			// Verify streaming completed
			if err := streamValidator.Complete(); err != nil {
				t.Fatalf("Streaming validation failed: %v", err)
			}
		})
	}
}

// TestWebSocketClientStreaming tests client-to-server streaming
func TestWebSocketClientStreaming(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	
	// Create test harness
	h := harness.New(t)
	
	// Get WebSocket scenarios
	matrix := scenarios.GenerateTestMatrix()
	
	// Run client streaming scenarios
	runner := scenarios.NewScenarioRunner(h)
	
	for _, scenario := range matrix {
		if scenario.Transport != scenarios.TransportWebSocket {
			continue
		}
		if scenario.Streaming != scenarios.StreamingClient {
			continue
		}
		
		scenario := scenario // capture range variable
		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel() // Run test cases in parallel
			// Add validators
			scenario.Validators = validators.StandardValidators()
			
			// Run scenario
			if err := runner.Run(scenario); err != nil {
				t.Fatalf("Client streaming scenario failed: %v", err)
			}
		})
	}
}

// TestWebSocketBidirectional tests bidirectional streaming
func TestWebSocketBidirectional(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	
	// Create test harness
	h := harness.New(t)
	
	// Get WebSocket scenarios
	matrix := scenarios.GenerateTestMatrix()
	
	// Run bidirectional scenarios
	runner := scenarios.NewScenarioRunner(h)
	
	for _, scenario := range matrix {
		if scenario.Transport != scenarios.TransportWebSocket {
			continue
		}
		if scenario.Streaming != scenarios.StreamingBidirectional {
			continue
		}
		
		scenario := scenario // capture range variable
		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel() // Run test cases in parallel
			// Add validators  
			streamValidator := validators.NewStreamingValidator(3, true) // 3 responses received
			scenario.Validators = append(
				validators.StandardValidators(),
				streamValidator,
			)
			
			// Run scenario
			if err := runner.Run(scenario); err != nil {
				t.Fatalf("Bidirectional streaming scenario failed: %v", err)
			}
		})
	}
}

// TestWebSocketConnectionLifecycle tests connection management
func TestWebSocketConnectionLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	
	// Create test harness
	h := harness.New(t)
	
	// Create a simple WebSocket scenario
	scenario := scenarios.Scenario{
		Name:        "websocket_lifecycle",
		Description: "Test WebSocket connection lifecycle",
		Transport:   scenarios.TransportWebSocket,
		PayloadType: scenarios.DataTypePrimitive,
		ResultType:  scenarios.DataTypePrimitive,
		Streaming:   scenarios.StreamingBidirectional,
		Features:    []scenarios.Feature{scenarios.FeatureStreaming},
		DSLCode:     createLifecycleDSLCode(),
		Requests: []scenarios.TestRequest{
			{
				Method: "test_stream",
				StreamingMessages: []scenarios.StreamMessage{
					{Direction: scenarios.DirectionSend, Data: map[string]any{"id": "req-1", "data": "message 1"}},
					{Direction: scenarios.DirectionReceive, Data: map[string]any{"id": "req-1", "data": "message 1"}},
					{Direction: scenarios.DirectionSend, Data: map[string]any{"id": "req-2", "data": "message 2"}},
					{Direction: scenarios.DirectionReceive, Data: map[string]any{"id": "req-2", "data": "message 2"}},
				},
			},
		},
		Validators: validators.StandardValidators(),
	}
	
	// Run scenario
	runner := scenarios.NewScenarioRunner(h)
	if err := runner.Run(scenario); err != nil {
		t.Fatalf("Lifecycle test failed: %v", err)
	}
}

// TestWebSocketErrorHandling tests error propagation in WebSocket
func TestWebSocketErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	
	// Create test harness
	h := harness.New(t)
	
	// Create error scenario
	scenario := scenarios.Scenario{
		Name:        "websocket_errors",
		Description: "Test WebSocket error handling",
		Transport:   scenarios.TransportWebSocket,
		PayloadType: scenarios.DataTypePrimitive,
		ResultType:  scenarios.DataTypePrimitive,
		Streaming:   scenarios.StreamingBidirectional,
		Features:    []scenarios.Feature{scenarios.FeatureStreaming, scenarios.FeatureErrors},
		DSLCode:     createWebSocketErrorDSLCode(),
		Requests: []scenarios.TestRequest{
			{
				Method: "error_stream",
				StreamingMessages: []scenarios.StreamMessage{
					{Direction: scenarios.DirectionSend, Data: map[string]any{"id": "req-1", "data": "trigger_error"}},
					// No DirectionReceive - we expect an error response, not a successful result
				},
			},
		},
		Validators: []validators.Validator{
			validators.ProtocolValidator(),
			validators.ErrorValidator(-32603, "internal error"),
		},
	}
	
	// Run scenario
	runner := scenarios.NewScenarioRunner(h)
	if err := runner.Run(scenario); err != nil {
		t.Fatalf("Error handling test failed: %v", err)
	}
}

// createLifecycleDSLCode creates a DSL for lifecycle testing
func createLifecycleDSLCode() string {
	// Return a WebSocket DSL with proper JSON-RPC streaming objects
	return `	API("test", func() {
		Title("WebSocket Lifecycle Test")
	})
	
	Service("lifecycle", func() {
		HTTP(func() {
			Path("/api")  // HTTP path for service
		})
		
		Method("test_stream", func() {
			// JSON-RPC streaming requires objects with request ID metadata
			StreamingPayload(func() {
				Attribute("id", String, func() {
					Meta("jsonrpc:id")
				})
				Attribute("data", String)
				Required("id", "data")
			})
			
			StreamingResult(func() {
				Attribute("id", String, func() {
					Meta("jsonrpc:id")
				})
				Attribute("data", String)
				Required("id", "data")
			})
			
			JSONRPC(func() {
				GET("/jsonrpc/ws")  // Method-level WebSocket endpoint
			})
		})
	})`
}

// createWebSocketErrorDSLCode creates a DSL for error testing
func createWebSocketErrorDSLCode() string {
	return `	API("test", func() {
		Title("WebSocket Error Test")
	})
	
	Service("errors", func() {
		HTTP(func() {
			Path("/api")  // HTTP path for service
		})
		
		Error("StreamError")
		
		Method("error_stream", func() {
			// Bidirectional streaming like the working lifecycle test
			StreamingPayload(func() {
				Attribute("id", String, func() {
					Meta("jsonrpc:id")
				})
				Attribute("data", String)
				Required("id", "data")
			})
			
			StreamingResult(func() {
				Attribute("id", String, func() {
					Meta("jsonrpc:id")
				})
				Attribute("data", String)
				Required("id", "data")
			})
			
			Error("StreamError")
			
			JSONRPC(func() {
				GET("/jsonrpc/ws")  // Method-level WebSocket endpoint
			})
		})
	})`
}