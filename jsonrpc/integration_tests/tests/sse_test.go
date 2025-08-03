package tests

import (
	"testing"

	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
	"goa.design/goa/v3/jsonrpc/integration_tests/scenarios"
	"goa.design/goa/v3/jsonrpc/integration_tests/validators"
)

// TestSSEStreaming tests Server-Sent Events streaming
func TestSSEStreaming(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	
	// Create test harness
	h := harness.New(t)
	
	// Get SSE scenarios
	matrix := scenarios.GenerateTestMatrix()
	
	// Run SSE scenarios
	runner := scenarios.NewScenarioRunner(h)
	
	for _, scenario := range matrix {
		if scenario.Transport != scenarios.TransportSSE {
			continue
		}
		
		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel() // Run scenarios in parallel
			
			// Add SSE validators
			streamValidator := validators.NewStreamingValidator(5, false) // Expect 5 events
			scenario.Validators = []validators.Validator{
				streamValidator,
				validators.NewSSEEventValidator(""),
			}
			
			// Run scenario
			if err := runner.Run(scenario); err != nil {
				t.Fatalf("SSE scenario failed: %v", err)
			}
			
			// Verify streaming completed
			if err := streamValidator.Complete(); err != nil {
				t.Fatalf("SSE streaming validation failed: %v", err)
			}
		})
	}
}

// TestSSENoPayload tests SSE with no initial payload
func TestSSENoPayload(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	
	// Create test harness
	h := harness.New(t)
	
	// Create no-payload scenario
	scenario := scenarios.Scenario{
		Name:        "sse_no_payload",
		Description: "SSE streaming without initial payload",
		Transport:   scenarios.TransportSSE,
		PayloadType: scenarios.DataTypeNone,
		ResultType:  scenarios.DataTypePrimitive,
		Streaming:   scenarios.StreamingServer,
		Features:    []scenarios.Feature{scenarios.FeatureStreaming},
		DSLCode:     createSSENoPayloadDSLCode(),
		Requests: []scenarios.TestRequest{
			{
				Method: "subscribe",
				Params: nil,
				StreamingMessages: []scenarios.StreamMessage{
					{Direction: scenarios.DirectionReceive, Data: map[string]any{"value": "event 1"}},
					{Direction: scenarios.DirectionReceive, Data: map[string]any{"value": "event 2"}},
					{Direction: scenarios.DirectionReceive, Data: map[string]any{"value": "event 3"}},
				},
			},
		},
		Validators: []validators.Validator{
			validators.NewSSEEventValidator(""),
			validators.DataIntegrityValidator(),
		},
	}
	
	// Run scenario
	runner := scenarios.NewScenarioRunner(h)
	if err := runner.Run(scenario); err != nil {
		t.Fatalf("SSE no-payload scenario failed: %v", err)
	}
}

// TestSSEComplexTypes tests SSE with complex data types
func TestSSEComplexTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	
	// Create test harness
	h := harness.New(t)
	
	// Get standard SSE scenarios from the test matrix
	matrix := scenarios.GenerateTestMatrix()
	
	// Find SSE scenarios with complex types
	for _, scenario := range matrix {
		if scenario.Transport == scenarios.TransportSSE && 
			(scenario.ResultType == scenarios.DataTypeComplex || 
			 scenario.ResultType == scenarios.DataTypeObject ||
			 scenario.ResultType == scenarios.DataTypeUserType) {
			
			t.Run(scenario.Name, func(t *testing.T) {
				// Replace validators with SSE-specific ones
				scenario.Validators = []validators.Validator{
					validators.NewSSEEventValidator(""),
					validators.DataIntegrityValidator(),
				}
				
				// Run scenario
				runner := scenarios.NewScenarioRunner(h)
				if err := runner.Run(scenario); err != nil {
					t.Fatalf("SSE scenario %s failed: %v", scenario.Name, err)
				}
			})
		}
	}
}

// TestSSEConnectionHandling tests SSE connection lifecycle
func TestSSEConnectionHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	
	// Create test harness
	h := harness.New(t)
	
	// Get standard SSE scenarios from the test matrix
	matrix := scenarios.GenerateTestMatrix()
	
	// Find SSE scenarios with primitive types to test basic connection
	for _, scenario := range matrix {
		if scenario.Transport == scenarios.TransportSSE && 
			scenario.PayloadType == scenarios.DataTypePrimitive &&
			scenario.ResultType == scenarios.DataTypePrimitive {
			
			t.Run("sse_connection", func(t *testing.T) {
				// Replace validators with SSE-specific ones
				scenario.Validators = []validators.Validator{
					validators.NewSSEEventValidator(""),
					validators.DataIntegrityValidator(),
				}
				
				// Run scenario
				runner := scenarios.NewScenarioRunner(h)
				if err := runner.Run(scenario); err != nil {
					t.Fatalf("SSE connection test failed: %v", err)
				}
			})
			break // Only run one scenario for this test
		}
	}
}

// createSSENoPayloadDSLCode creates a DSL for SSE without payload
func createSSENoPayloadDSLCode() string {
	return `	API("test", func() {
		Title("SSE No Payload Test")
	})

	Service("events", func() {
		JSONRPC(func() {
			POST("/jsonrpc/sse")
			ServerSentEvents()
		})
		Method("subscribe", func() {
			// No payload
			StreamingResult(func() {
				Attribute("value", String, "The streamed value")
				Required("value")
			})
			
			JSONRPC(func() {
			})
		})
	})`
}