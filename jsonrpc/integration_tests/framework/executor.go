package framework

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
)

// Executor handles test scenario execution
type Executor struct {
	serverURL string
	config    executorConfig
}

// NewExecutor creates a new test executor
func NewExecutor(serverURL string, opts ...ExecutorOption) *Executor {
	config := executorConfig{
		WebSocketTimeout: 30 * time.Second,
		Debug:            false,
	}
	
	for _, opt := range opts {
		opt(&config)
	}
	
	return &Executor{
		serverURL: serverURL,
		config:    config,
	}
}

// Execute runs a test scenario
func (e *Executor) Execute(t *testing.T, scenario Scenario) {
	t.Helper()
	
	if e.config.Debug {
		t.Logf("Executing scenario: %s", scenario.Name)
		t.Logf("Transport: %s, Method: %s", scenario.Transport, scenario.Method)
	}
	
	// Handle different scenario types
	if len(scenario.Sequence) > 0 {
		e.executeStreaming(t, scenario)
	} else if len(scenario.Batch) > 0 {
		e.executeBatch(t, scenario)
	} else if scenario.RawRequest != "" {
		e.executeRaw(t, scenario)
	} else {
		e.executeSimple(t, scenario)
	}
}

// executeSimple handles basic request/response scenarios
func (e *Executor) executeSimple(t *testing.T, scenario Scenario) {
	t.Helper()
	
	ctx := context.Background()
	
	// Create client based on transport
	switch scenario.Transport {
	case TransportHTTP:
		e.executeHTTP(ctx, t, scenario)
	case TransportWebSocket:
		e.executeWebSocket(ctx, t, scenario)
	case TransportSSE:
		e.executeSSE(ctx, t, scenario)
	default:
		t.Fatalf("Unknown transport: %s", scenario.Transport)
	}
}

// executeHTTP handles HTTP transport scenarios
func (e *Executor) executeHTTP(ctx context.Context, t *testing.T, scenario Scenario) {
	t.Helper()
	
	// Create client
	client, err := harness.NewClient(e.serverURL, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Build request
	method := scenario.Request.GetMethod(scenario.Method)
	
	// Try CLI client first (disabled for now)
	// TODO: Re-enable when CLI client is implemented
	/*
	cliClient, err := harness.NewCLIClient(workDir, e.serverURL)
	if err == nil && cliClient.CanHandle(method, scenario.Request.Params) {
		if e.config.Debug {
			t.Logf("Using CLI client for method: %s", method)
		}
		
		result, err := cliClient.Call(ctx, method, scenario.Request.Params, scenario.Request.ID)
		if err != nil {
			if scenario.Expect.Error != nil {
				// Expected error - validate it
				e.validateError(t, err, scenario.Expect.Error)
				return
			}
			t.Fatalf("CLI call failed: %v", err)
		}
		
		// Validate result
		e.validateResult(t, result, scenario.Expect)
		return
	}
	*/
	
	// Fall back to direct client
	if e.config.Debug {
		t.Logf("Using direct client for method: %s", method)
	}
	
	req := harness.JSONRPCRequest{
		Method: method,
		Params: scenario.Request.Params,
		ID:     scenario.Request.ID,
	}
	result, err := client.CallHTTP(ctx, req)
	if err != nil {
		if scenario.Expect.Error != nil {
			e.validateError(t, err, scenario.Expect.Error)
			return
		}
		t.Fatalf("HTTP call failed: %v", err)
	}
	
	// Handle notification case
	if scenario.Expect.NoResponse {
		if result != nil {
			t.Errorf("Expected no response for notification, got: %v", result)
		}
		return
	}
	
	// Parse response
	if result != nil {
		var resp interface{}
		if err := json.Unmarshal(result, &resp); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}
		e.validateJSONRPCResponse(t, resp, scenario.Expect)
	} else if !scenario.Expect.NoResponse {
		t.Errorf("Expected response but got none")
	}
}

// executeWebSocket handles WebSocket transport scenarios
func (e *Executor) executeWebSocket(ctx context.Context, t *testing.T, scenario Scenario) {
	t.Helper()
	
	// WebSocket scenarios always use sequence
	if len(scenario.Sequence) > 0 {
		e.executeWebSocketSequence(ctx, t, scenario)
		return
	}
	
	// If no sequence, create a simple send/receive sequence from request/expect
	if scenario.Request.Params != nil {
		// Pass method, params, and id as separate fields
		data := map[string]any{
			"method": scenario.Method,
			"params": scenario.Request.Params,
		}
		if scenario.Request.ID != nil {
			data["id"] = scenario.Request.ID
		}
		
		scenario.Sequence = []Action{
			{Type: "send", Data: data},
			{Type: "receive", Expect: scenario.Expect},
		}
		e.executeWebSocketSequence(ctx, t, scenario)
	}
}

// executeSSE handles Server-Sent Events scenarios
func (e *Executor) executeSSE(ctx context.Context, t *testing.T, scenario Scenario) {
	t.Helper()
	
	// SSE implementation would go here
	// For now, just a placeholder
	t.Skip("SSE transport not yet implemented")
}

// executeStreaming handles streaming scenarios with sequences
func (e *Executor) executeStreaming(t *testing.T, scenario Scenario) {
	t.Helper()
	
	ctx := context.Background()
	
	// Only WebSocket and SSE support streaming
	switch scenario.Transport {
	case TransportWebSocket:
		e.executeWebSocketSequence(ctx, t, scenario)
	case TransportSSE:
		e.executeSSESequence(ctx, t, scenario)
	default:
		t.Fatalf("Transport %s does not support streaming", scenario.Transport)
	}
}

// executeWebSocketSequence handles WebSocket streaming sequences
func (e *Executor) executeWebSocketSequence(ctx context.Context, t *testing.T, scenario Scenario) {
	t.Helper()
	
	client, err := harness.NewClient(e.serverURL, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Execute sequence steps
	for i, step := range scenario.Sequence {
		if e.config.Debug {
			t.Logf("Executing step %d: %s", i, step.Type)
		}
		switch step.Type {
		case "connect":
			if err := client.ConnectWebSocket(ctx); err != nil {
				t.Fatalf("Step %d: failed to connect WebSocket: %v", i, err)
			}
			
		case "send":
			// Auto-connect if not connected
			if !client.IsConnected() {
				if err := client.ConnectWebSocket(ctx); err != nil {
					t.Fatalf("Step %d: failed to auto-connect WebSocket: %v", i, err)
				}
			}
			
			if step.Data == nil {
				t.Fatalf("Step %d: send step requires data", i)
			}
			
			// Extract method, params, and id from the data
			reqData, ok := step.Data.(map[string]any)
			if !ok {
				t.Fatalf("Step %d: invalid request data format", i)
			}
			
			req := harness.JSONRPCRequest{
				Method: reqData["method"].(string),
				Params: reqData["params"],
				ID:     reqData["id"],
			}
			
			if err := client.SendWebSocket(ctx, req); err != nil {
				t.Fatalf("Step %d: failed to send: %v", i, err)
			}
			
		case "receive":
			if e.config.Debug {
				t.Logf("Step %d: waiting to receive", i)
			}
			msg, err := client.ReceiveWebSocket(ctx)
			if err != nil {
				t.Fatalf("Step %d: failed to receive: %v", i, err)
			}
			if e.config.Debug {
				t.Logf("Step %d: received: %s", i, string(msg))
			}
			
			var response map[string]interface{}
			if err := json.Unmarshal(msg, &response); err != nil {
				t.Fatalf("Step %d: failed to unmarshal response: %v", i, err)
			}
			
			// Compare the response with expected
			if expected, ok := step.Expect.(map[string]interface{}); ok {
				e.compareJSONRPCMessages(t, response, expected)
			} else {
				t.Fatalf("Step %d: expected value must be a map", i)
			}
			
		case "close":
			if err := client.CloseWebSocket(); err != nil {
				t.Fatalf("Step %d: failed to close WebSocket: %v", i, err)
			}
			
		default:
			t.Fatalf("Step %d: unknown step type: %s", i, step.Type)
		}
		
		// Apply delay if specified
		if step.Delay > 0 {
			time.Sleep(step.Delay)
		}
	}
}

// executeSSESequence handles SSE streaming sequences
func (e *Executor) executeSSESequence(ctx context.Context, t *testing.T, scenario Scenario) {
	t.Helper()
	
	// SSE only supports server-to-client streaming
	require.True(t, scenario.Request.Params != nil || scenario.Request.ID != nil, "SSE requires an initial request")
	
	client, err := harness.NewClient(e.serverURL, nil)
	require.NoError(t, err, "Failed to create client")
	
	// Send request and get SSE events
	method := scenario.Request.GetMethod(scenario.Method)
	req := harness.JSONRPCRequest{
		Method: method,
		Params: scenario.Request.Params,
		ID:     scenario.Request.ID,
	}
	events, err := client.CallSSE(ctx, req)
	if err != nil {
		t.Fatalf("SSE request failed: %v", err)
	}
	
	// Validate sequence
	if len(events) != len(scenario.Sequence) {
		t.Fatalf("Expected %d events, got %d", len(scenario.Sequence), len(events))
	}
	
	for i, step := range scenario.Sequence {
		if step.Type != "receive" {
			t.Fatalf("SSE only supports 'receive' steps, got %s", step.Type)
		}
		
		if i >= len(events) {
			t.Fatalf("Expected event at step %d, but no more events", i)
		}
		
		// Parse and validate the event
		var response map[string]interface{}
		if err := json.Unmarshal(events[i], &response); err != nil {
			t.Fatalf("Failed to unmarshal event %d: %v", i, err)
		}
		
		// For SSE streaming, step.Expect contains the full expected JSON-RPC message
		expectedMsg, ok := step.Expect.(map[string]interface{})
		require.True(t, ok, "Step %d: invalid expect format", i)
		
		// Compare the messages
		e.compareJSONRPCMessages(t, response, expectedMsg)
	}
}

// executeBatch handles batch request scenarios
func (e *Executor) executeBatch(t *testing.T, scenario Scenario) {
	t.Helper()
	
	// Batch requests only work with HTTP
	if scenario.Transport != TransportHTTP {
		t.Fatalf("Batch requests only supported on HTTP transport")
	}
	
	ctx := context.Background()
	client, err := harness.NewClient(e.serverURL, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Build batch request
	var batch []interface{}
	for _, req := range scenario.Batch {
		method := req.GetMethod(scenario.Method)
		jsonReq := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  method,
			"params":  req.Params,
		}
		if req.ID != nil {
			jsonReq["id"] = req.ID
		}
		batch = append(batch, jsonReq)
	}
	
	// Send batch
	batchJSON, err := json.Marshal(batch)
	if err != nil {
		t.Fatalf("Failed to marshal batch: %v", err)
	}
	
	responseJSON, err := client.CallHTTPRaw(ctx, batchJSON)
	if err != nil {
		t.Fatalf("Batch call failed: %v", err)
	}
	
	// Parse batch response
	var responses []json.RawMessage
	if err := json.Unmarshal(responseJSON, &responses); err != nil {
		t.Fatalf("Failed to parse batch response: %v", err)
	}
	
	// Validate responses
	if len(responses) != len(scenario.ExpectBatch) {
		t.Fatalf("Expected %d responses, got %d", len(scenario.ExpectBatch), len(responses))
	}
	
	for i, respJSON := range responses {
		var resp map[string]interface{}
		if err := json.Unmarshal(respJSON, &resp); err != nil {
			t.Fatalf("Failed to parse response %d: %v", i, err)
		}
		
		e.validateBatchResponse(t, i, resp, scenario.ExpectBatch[i])
	}
}

// executeRaw handles raw request scenarios
func (e *Executor) executeRaw(t *testing.T, scenario Scenario) {
	t.Helper()
	
	// Raw requests only work with HTTP
	if scenario.Transport != TransportHTTP {
		t.Fatalf("Raw requests only supported on HTTP transport")
	}
	
	ctx := context.Background()
	client, err := harness.NewClient(e.serverURL, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Send raw request
	responseJSON, err := client.CallHTTPRaw(ctx, []byte(scenario.RawRequest))
	if err != nil {
		if scenario.Expect.Error != nil {
			// Expected error
			return
		}
		t.Fatalf("Raw call failed: %v", err)
	}
	
	// Parse response
	var resp interface{}
	if err := json.Unmarshal(responseJSON, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Validate response
	e.validateRawResponse(t, resp, scenario.Expect)
}

// Validation methods

func (e *Executor) validateResult(t *testing.T, result interface{}, expect Expect) {
	t.Helper()
	
	// Parse result as JSON-RPC response
	respMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map response, got %T", result)
	}
	
	e.validateJSONRPCResponse(t, respMap, expect)
}

func (e *Executor) validateJSONRPCResponse(t *testing.T, response interface{}, expect Expect) {
	t.Helper()
	
	respMap, ok := response.(map[string]interface{})
	require.True(t, ok, "Expected map response, got %T", response)
	
	// Check ID
	if expect.ID != nil {
		assert.EqualValues(t, expect.ID, respMap["id"], "ID mismatch")
	}
	
	// Check result or error
	if expect.Error != nil {
		// Expecting error
		errObj, ok := respMap["error"].(map[string]interface{})
		require.True(t, ok, "Expected error response, got result: %v", respMap["result"])
		
		e.validateErrorObject(t, errObj, expect.Error)
	} else {
		// Expecting result
		_, hasError := respMap["error"]
		require.False(t, hasError, "Expected result, got error: %v", respMap["error"])
		
		// Use JSONEq for complex types or EqualValues for simple types
		expectedJSON, errExp := json.Marshal(expect.Result)
		actualJSON, errAct := json.Marshal(respMap["result"])
		if errExp == nil && errAct == nil {
			assert.JSONEq(t, string(expectedJSON), string(actualJSON), "Result mismatch")
		} else {
			assert.EqualValues(t, expect.Result, respMap["result"], "Result mismatch")
		}
	}
}

// compareJSONRPCMessages compares two JSON-RPC messages (used for SSE/WebSocket validation)
func (e *Executor) compareJSONRPCMessages(t *testing.T, actual, expected map[string]interface{}) {
	t.Helper()
	
	// Compare jsonrpc version
	if actualVersion, ok := actual["jsonrpc"].(string); ok {
		expectedVersion, _ := expected["jsonrpc"].(string)
		require.Equal(t, expectedVersion, actualVersion, "JSON-RPC version mismatch")
	}
	
	// Compare method
	if actualMethod, ok := actual["method"].(string); ok {
		expectedMethod, _ := expected["method"].(string)
		require.Equal(t, expectedMethod, actualMethod, "Method mismatch")
	}
	
	// Compare params
	if expectedParams, ok := expected["params"]; ok {
		actualParams, ok := actual["params"]
		require.True(t, ok, "Expected params in response")
		e.compareValues(t, actualParams, expectedParams, "params")
	}
	
	// Compare result
	if expectedResult, ok := expected["result"]; ok {
		actualResult, ok := actual["result"]
		require.True(t, ok, "Expected result in response")
		e.compareValues(t, actualResult, expectedResult, "result")
	}
	
	// Compare error
	if expectedError, ok := expected["error"]; ok {
		actualError, ok := actual["error"]
		require.True(t, ok, "Expected error in response")
		e.compareValues(t, actualError, expectedError, "error")
	}
	
	// Compare id
	if expectedID, ok := expected["id"]; ok {
		actualID, ok := actual["id"]
		require.True(t, ok, "Expected id in response")
		e.compareValues(t, actualID, expectedID, "id")
	}
}

func (e *Executor) validateWebSocketResponse(t *testing.T, response interface{}, expect Expect) {
	t.Helper()
	
	// WebSocket responses are the same as JSON-RPC responses
	e.validateJSONRPCResponse(t, response, expect)
}

func (e *Executor) validateBatchResponse(t *testing.T, index int, response map[string]interface{}, expect Expect) {
	t.Helper()
	
	// Batch responses are validated the same way
	e.validateJSONRPCResponse(t, response, expect)
}

func (e *Executor) validateRawResponse(t *testing.T, response interface{}, expect Expect) {
	t.Helper()
	
	// Raw responses might not be standard JSON-RPC
	if expect.Error != nil {
		// For raw requests, we might get non-standard errors
		t.Logf("Raw error response: %v", response)
		return
	}
	
	// Try to validate as JSON-RPC response
	if respMap, ok := response.(map[string]interface{}); ok {
		e.validateJSONRPCResponse(t, respMap, expect)
	} else {
		// Just compare directly
		assert.EqualValues(t, expect.Result, response, "Raw response mismatch")
	}
}

func (e *Executor) validateError(t *testing.T, err error, expect *ExpectError) {
	t.Helper()
	
	// For CLI errors, we need to extract the error details
	// This is a simplified version - real implementation would parse the error
	t.Logf("Got error: %v", err)
	
	// TODO: Parse error and validate code/message
}

func (e *Executor) validateErrorObject(t *testing.T, errObj map[string]interface{}, expect *ExpectError) {
	t.Helper()
	
	// Check error code
	code, ok := errObj["code"].(float64)
	require.True(t, ok, "Missing or invalid error code")
	assert.EqualValues(t, expect.Code, int(code), "Error code mismatch")
	
	// Check error message
	msg, ok := errObj["message"].(string)
	require.True(t, ok, "Missing or invalid error message")
	assert.Equal(t, expect.Message, msg, "Error message mismatch")
	
	// Check error data if expected
	if expect.Data != nil {
		assert.Equal(t, expect.Data, errObj["data"], "Error data mismatch")
	}
}

// compareValues compares two values, handling both simple and complex types
func (e *Executor) compareValues(t *testing.T, actual, expected interface{}, path string) {
	t.Helper()
	
	// Try JSON comparison first for complex types
	expectedJSON, errExp := json.Marshal(expected)
	actualJSON, errAct := json.Marshal(actual)
	if errExp == nil && errAct == nil {
		assert.JSONEq(t, string(expectedJSON), string(actualJSON), "%s mismatch", path)
	} else {
		// Fall back to direct comparison
		assert.EqualValues(t, expected, actual, "%s mismatch", path)
	}
}

