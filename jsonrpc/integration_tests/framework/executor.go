package framework

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
)

// executor handles test scenario execution
type executor struct {
	serverURL string
	config    executorConfig
}

// newExecutor creates a new test executor
func newExecutor(serverURL string, opts ...executorOption) *executor {
	config := executorConfig{
		WebSocketTimeout: 30 * time.Second,
		Debug:            false,
	}

	for _, opt := range opts {
		opt(&config)
	}

	return &executor{
		serverURL: serverURL,
		config:    config,
	}
}

// Execute runs a test scenario
func (e *executor) Execute(t *testing.T, scenario Scenario) {
	t.Helper()

	// Handle different scenario types
	switch {
	case len(scenario.Sequence) > 0:
		e.executeStreaming(t, scenario)
	case len(scenario.Batch) > 0:
		e.executeBatch(t, scenario)
	case scenario.RawRequest != "":
		e.executeRaw(t, scenario)
	default:
		e.executeSimple(t, scenario)
	}
}

// executeSimple handles basic request/response scenarios
func (e *executor) executeSimple(t *testing.T, scenario Scenario) {
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
		require.Failf(t, "Unknown transport", "Unknown transport: %s", scenario.Transport)
	}
}

// executeHTTP handles HTTP transport scenarios
func (e *executor) executeHTTP(ctx context.Context, t *testing.T, scenario Scenario) {
	t.Helper()

	// Create client
	client, err := harness.NewClient(e.serverURL, nil)
	require.NoError(t, err, "Failed to create client")

	// Build request
	method := scenario.Request.GetMethod(scenario.Method)

	// Try CLI client first for non-streaming scenarios
	// Skip CLI if custom JSONRPC field is specified
	if e.config.WorkDir != "" && scenario.Request.JSONRPC == "" {
		cliClient, err := harness.NewCLIClient(e.config.WorkDir, e.serverURL)
		if err != nil {
		} else if cliClient.CanHandle(method, scenario.Request.Params) {
			// For CLI, we need to separate service and method
			// Default to "test" service if no dot in method name
			service := "test"
			methodName := method
			if parts := strings.Split(method, "."); len(parts) == 2 {
				service = parts[0]
				methodName = parts[1]
			}

			result, err := cliClient.CallMethod(ctx, service, methodName, scenario.Request.Params)
			if err != nil {
				if scenario.Expect.Error != nil {
					// Expected error - validate it
					e.validateError(t, err, scenario.Expect.Error)
					return
				}
				require.NoError(t, err, "CLI call failed")
			}

			// With verbose flag, CLI now returns the raw transport-level response
			if result != nil {
				// Wrap in JSON-RPC envelope
				response := map[string]any{
					"jsonrpc": "2.0",
					"id":      scenario.Request.ID,
					"result":  result,
				}
				e.validateJSONRPCResponse(t, response, scenario.Expect)
			} else if !scenario.Expect.NoResponse {
				assert.Fail(t, "Expected response but got none")
			}
			return
		}
	}

	// Fall back to direct client

	req := harness.JSONRPCRequest{
		Method: method,
		Params: scenario.Request.Params,
		ID:     scenario.Request.ID,
	}
	// Handle JSONRPC field:
	// - Not specified (empty string) → Use default "2.0"
	// - "-" → Omit the field entirely
	// - Any other value → Use that value
	if scenario.Request.JSONRPC == "-" {
		// Special value to omit the field
		emptyStr := ""
		req.JSONRPC = &emptyStr
	} else if scenario.Request.JSONRPC != "" {
		// Custom value specified
		req.JSONRPC = &scenario.Request.JSONRPC
	}
	// If JSONRPC is empty string (not specified), req.JSONRPC remains nil and defaults to "2.0"
	result, err := client.CallHTTP(ctx, req)
	if err != nil {
		if scenario.Expect.Error != nil {
			e.validateError(t, err, scenario.Expect.Error)
			return
		}
		require.NoError(t, err, "HTTP call failed")
	}

	// Handle notification case
	if scenario.Expect.NoResponse {
		assert.Nil(t, result, "Expected no response for notification")
		return
	}

	// Parse response
	if result != nil {
		var resp any
		err := json.Unmarshal(result, &resp)
		require.NoError(t, err, "Failed to parse response")
		e.validateJSONRPCResponse(t, resp, scenario.Expect)
	} else if !scenario.Expect.NoResponse {
		assert.Fail(t, "Expected response but got none")
	}
}

// executeWebSocket handles WebSocket transport scenarios
func (e *executor) executeWebSocket(ctx context.Context, t *testing.T, scenario Scenario) {
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
func (e *executor) executeSSE(_ context.Context, t *testing.T, _ Scenario) {
	t.Helper()

	// SSE implementation would go here
	// For now, just a placeholder
	t.Skip("SSE transport not yet implemented")
}

// executeStreaming handles streaming scenarios with sequences
func (e *executor) executeStreaming(t *testing.T, scenario Scenario) {
	t.Helper()

	ctx := context.Background()

	// Only WebSocket and SSE support streaming
	switch scenario.Transport {
	case TransportWebSocket:
		e.executeWebSocketSequence(ctx, t, scenario)
	case TransportSSE:
		e.executeSSESequence(ctx, t, scenario)
	default:
		require.Failf(t, "Unsupported transport", "Transport %s does not support streaming", scenario.Transport)
	}
}

// executeWebSocketSequence handles WebSocket streaming sequences
func (e *executor) executeWebSocketSequence(ctx context.Context, t *testing.T, scenario Scenario) {
	t.Helper()

	client, err := harness.NewClient(e.serverURL, nil)
	require.NoError(t, err, "Failed to create client")

	// Execute sequence steps
	for i, step := range scenario.Sequence {
		switch step.Type {
		case "connect":
			err := client.ConnectWebSocket(ctx)
			require.NoErrorf(t, err, "Step %d: failed to connect WebSocket", i)

		case "send":
			// Auto-connect if not connected
			if !client.IsConnected() {
				err := client.ConnectWebSocket(ctx)
				require.NoErrorf(t, err, "Step %d: failed to auto-connect WebSocket", i)
			}

			require.NotNilf(t, step.Data, "Step %d: send step requires data", i)

			// Extract method, params, and id from the data
			reqData, ok := step.Data.(map[string]any)
			require.Truef(t, ok, "Step %d: invalid request data format", i)

			req := harness.JSONRPCRequest{
				Method: reqData["method"].(string),
				Params: reqData["params"],
				ID:     reqData["id"],
			}

			// Mark HasID when id key present (even if it's null)
			if _, hasID := reqData["id"]; hasID {
				req.HasID = true
			}

			// Handle custom jsonrpc field if specified
			if jsonrpcVal, ok := reqData["jsonrpc"]; ok {
				if jsonrpcStr, ok := jsonrpcVal.(string); ok {
					if jsonrpcStr == "-" {
						// Special value to omit the field
						emptyStr := ""
						req.JSONRPC = &emptyStr
					} else {
						req.JSONRPC = &jsonrpcStr
					}
				}
			}
			// If not specified, JSONRPC remains nil and defaults to "2.0"

			err := client.SendWebSocket(ctx, req)
			require.NoErrorf(t, err, "Step %d: failed to send", i)

		case "receive":
			msg, err := client.ReceiveWebSocket(ctx)
			require.NoErrorf(t, err, "Step %d: failed to receive", i)

			var response map[string]any
			err = json.Unmarshal(msg, &response)
			require.NoErrorf(t, err, "Step %d: failed to unmarshal response", i)

			// Compare the response with expected
			if expected, ok := step.Expect.(map[string]any); ok {
				e.compareJSONRPCMessages(t, response, expected)
			} else {
				require.Failf(t, "Invalid expected value", "Step %d: expected value must be a map", i)
			}

		case "close":
			err := client.CloseWebSocket()
			require.NoErrorf(t, err, "Step %d: failed to close WebSocket", i)

		default:
			require.Failf(t, "Unknown step type", "Step %d: unknown step type: %s", i, step.Type)
		}

		// Apply delay if specified
		if step.Delay > 0 {
			time.Sleep(step.Delay)
		}
	}
}

// executeSSESequence handles SSE streaming sequences
func (e *executor) executeSSESequence(ctx context.Context, t *testing.T, scenario Scenario) {
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
	// Handle JSONRPC field:
	// - Not specified (empty string) → Use default "2.0"
	// - "-" → Omit the field entirely
	// - Any other value → Use that value
	if scenario.Request.JSONRPC == "-" {
		// Special value to omit the field
		emptyStr := ""
		req.JSONRPC = &emptyStr
	} else if scenario.Request.JSONRPC != "" {
		// Custom value specified
		req.JSONRPC = &scenario.Request.JSONRPC
	}
	// If JSONRPC is empty string (not specified), req.JSONRPC remains nil and defaults to "2.0"
	events, err := client.CallSSE(ctx, req)
	require.NoError(t, err, "SSE request failed")

	// Validate sequence
	require.Len(t, events, len(scenario.Sequence), "Event count mismatch")

	for i, step := range scenario.Sequence {
		require.Equalf(t, "receive", step.Type, "SSE only supports 'receive' steps, got %s", step.Type)

		require.Lessf(t, i, len(events), "Expected event at step %d, but no more events", i)

		// Parse and validate the event
		var response map[string]any
		err := json.Unmarshal(events[i], &response)
		require.NoErrorf(t, err, "Failed to unmarshal event %d", i)

		// For SSE streaming, step.Expect contains the full expected JSON-RPC message
		expectedMsg, ok := step.Expect.(map[string]any)
		require.True(t, ok, "Step %d: invalid expect format", i)

		// Compare the messages
		e.compareJSONRPCMessages(t, response, expectedMsg)
	}
}

// executeBatch handles batch request scenarios
func (e *executor) executeBatch(t *testing.T, scenario Scenario) {
	t.Helper()

	// Batch requests only work with HTTP
	require.Equal(t, TransportHTTP, scenario.Transport, "Batch requests only supported on HTTP transport")

	ctx := context.Background()
	client, err := harness.NewClient(e.serverURL, nil)
	require.NoError(t, err, "Failed to create client")

	// Build batch request
	batch := make([]any, 0, len(scenario.Batch))
	for _, req := range scenario.Batch {
		method := req.GetMethod(scenario.Method)
		jsonReq := map[string]any{
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
	require.NoError(t, err, "Failed to marshal batch")

	responseJSON, err := client.CallHTTPRaw(ctx, batchJSON)
	require.NoError(t, err, "Batch call failed")

	// Parse batch response
	var responses []json.RawMessage
	err = json.Unmarshal(responseJSON, &responses)
	require.NoError(t, err, "Failed to parse batch response")

	// Validate responses
	require.Len(t, responses, len(scenario.ExpectBatch), "Response count mismatch")

	for i, respJSON := range responses {
		var resp map[string]any
		err := json.Unmarshal(respJSON, &resp)
		require.NoErrorf(t, err, "Failed to parse response %d", i)

		e.validateBatchResponse(t, i, resp, scenario.ExpectBatch[i])
	}
}

// executeRaw handles raw request scenarios
func (e *executor) executeRaw(t *testing.T, scenario Scenario) {
	t.Helper()

	// Raw requests only work with HTTP
	require.Equal(t, TransportHTTP, scenario.Transport, "Raw requests only supported on HTTP transport")

	ctx := context.Background()
	client, err := harness.NewClient(e.serverURL, nil)
	require.NoError(t, err, "Failed to create client")

	// Send raw request
	responseJSON, err := client.CallHTTPRaw(ctx, []byte(scenario.RawRequest))
	if err != nil {
		if scenario.Expect.Error != nil {
			// Expected error
			return
		}
		require.NoError(t, err, "Raw call failed")
	}

	// Parse response
	var resp any
	err = json.Unmarshal(responseJSON, &resp)
	require.NoError(t, err, "Failed to parse response")

	// Validate response
	e.validateRawResponse(t, resp, scenario.Expect)
}

// Validation methods

func (e *executor) validateJSONRPCResponse(t *testing.T, response any, expect Expect) {
	t.Helper()

	respMap, ok := response.(map[string]any)
	require.True(t, ok, "Expected map response, got %T", response)

	// Check ID
	if expect.ID != nil {
		assert.EqualValues(t, expect.ID, respMap["id"], "ID mismatch")
	}

	// Check result or error
	if expect.Error != nil {
		// Expecting error
		errObj, ok := respMap["error"].(map[string]any)
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
func (e *executor) compareJSONRPCMessages(t *testing.T, actual, expected map[string]any) {
	t.Helper()

	// Compare jsonrpc version
	if actualVersion, ok := actual["jsonrpc"].(string); ok {
		expectedVersion, _ := expected["jsonrpc"].(string)
		require.Equal(t, expectedVersion, actualVersion, "JSON-RPC version mismatch")
	}

	// Compare method only if explicitly expected
	if _, expHas := expected["method"]; expHas {
		actualMethod, ok := actual["method"].(string)
		require.True(t, ok, "Expected method in response")
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

func (e *executor) validateBatchResponse(t *testing.T, _ int, response map[string]any, expect Expect) {
	t.Helper()

	// Batch responses are validated the same way
	e.validateJSONRPCResponse(t, response, expect)
}

func (e *executor) validateRawResponse(t *testing.T, response any, expect Expect) {
	t.Helper()

	// Raw responses might not be standard JSON-RPC
	if expect.Error != nil {
		// For raw requests, we might get non-standard errors
		return
	}

	// Try to validate as JSON-RPC response
	if respMap, ok := response.(map[string]any); ok {
		e.validateJSONRPCResponse(t, respMap, expect)
	} else {
		// Just compare directly
		assert.EqualValues(t, expect.Result, response, "Raw response mismatch")
	}
}

func (e *executor) validateError(t *testing.T, _ error, _ *ExpectError) {
	t.Helper()

	// For CLI errors, we need to extract the error details
	// This is a simplified version - real implementation would parse the error

	// TODO: Parse error and validate code/message
}

func (e *executor) validateErrorObject(t *testing.T, errObj map[string]any, expect *ExpectError) {
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
func (e *executor) compareValues(t *testing.T, actual, expected any, path string) {
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
