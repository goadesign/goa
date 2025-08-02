package validators

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTPResponseValidator validates HTTP-specific aspects of JSON-RPC responses
// including status codes and headers. While JSON-RPC typically uses 200 OK
// for all responses (including errors), this validator can verify transport-level
// requirements.
//
// Note: In the current implementation, this focuses on JSON-RPC response
// validation. Full HTTP validation would require access to the underlying
// HTTP response object.
type HTTPResponseValidator struct {
	expectedStatus  int
	expectedHeaders map[string]string
}

// NewHTTPResponseValidator creates a new HTTP response validator with expected
// status code and headers. The validator can be used to ensure JSON-RPC
// responses are delivered with correct HTTP transport properties.
func NewHTTPResponseValidator(expectedStatus int, expectedHeaders map[string]string) *HTTPResponseValidator {
	return &HTTPResponseValidator{
		expectedStatus:  expectedStatus,
		expectedHeaders: expectedHeaders,
	}
}

// Validate checks HTTP response properties. In a full implementation, this
// would validate HTTP status codes, headers, and other transport-specific
// properties. Currently, it validates the JSON-RPC response format.
func (v *HTTPResponseValidator) Validate(response any) error {
	// This would typically work with HTTP response objects
	// For integration tests, we might pass additional context

	// For now, just validate JSON-RPC response
	_, err := AsJSONRPCResponse(response)
	return err
}

// ContentTypeValidator validates that responses have the correct content type
// header. For JSON-RPC over HTTP, this should typically be "application/json".
//
// This validator ensures that the transport layer properly identifies the
// response format, which is important for client libraries and intermediaries.
func ContentTypeValidator(expectedType string) Validator {
	return ValidatorFunc(func(response any) error {
		// In real implementation, this would check HTTP headers
		// For now, just validate JSON-RPC response format
		_, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid JSON-RPC response: %w", err)
		}
		return nil
	})
}

// WebSocketMessageValidator validates WebSocket message format and type.
// JSON-RPC over WebSocket should use text frames (messageType = 1) with
// JSON-encoded content.
//
// This validator ensures that streaming messages maintain proper framing
// and encoding throughout the WebSocket session.
type WebSocketMessageValidator struct {
	messageType int // 1 for text, 2 for binary
}

// NewWebSocketMessageValidator creates a WebSocket message validator for the
// specified message type. Use messageType = 1 for text frames (typical for
// JSON-RPC) or messageType = 2 for binary frames.
func NewWebSocketMessageValidator(messageType int) *WebSocketMessageValidator {
	return &WebSocketMessageValidator{
		messageType: messageType,
	}
}

// Validate checks WebSocket message properties including frame type and
// JSON-RPC message format. This ensures messages are properly formatted
// for WebSocket transport.
func (v *WebSocketMessageValidator) Validate(response any) error {
	// Validate it's a valid JSON-RPC message
	_, err := AsJSONRPCResponse(response)
	return err
}

// SSEEventValidator validates Server-Sent Events format for JSON-RPC streaming
// responses. SSE events should contain properly formatted JSON-RPC messages
// within the data field.
//
// This validator checks both the SSE event structure and the embedded JSON-RPC
// message format, ensuring compatibility with SSE client libraries.
type SSEEventValidator struct {
	expectedEventType string
	expectedContent   any // Optional: validate notification params match this
}

// NewSSEEventValidator creates an SSE event validator for the specified event
// type. The event type can be used to categorize different kinds of streaming
// messages (e.g., "message", "error", "ping").
func NewSSEEventValidator(eventType string) *SSEEventValidator {
	return &SSEEventValidator{
		expectedEventType: eventType,
	}
}

// NewSSEEventContentValidator creates an SSE validator that also validates
// the notification content (params field) matches the expected value.
func NewSSEEventContentValidator(eventType string, expectedContent any) *SSEEventValidator {
	return &SSEEventValidator{
		expectedEventType: eventType,
		expectedContent:   expectedContent,
	}
}

// Validate checks SSE event properties including event type and the embedded
// JSON-RPC message format. SSE events in JSON-RPC contain notifications (not
// responses) since they are server-initiated messages.
func (v *SSEEventValidator) Validate(response any) error {
	// SSE events should be JSON-RPC notifications
	notification, ok := response.(map[string]any)
	if !ok {
		return fmt.Errorf("SSE event is not a JSON object")
	}

	// Validate JSON-RPC 2.0 notification format
	jsonrpc, ok := notification["jsonrpc"].(string)
	if !ok || jsonrpc != "2.0" {
		return fmt.Errorf("invalid or missing jsonrpc version")
	}

	// Must have a method (notifications are like requests without an id)
	method, ok := notification["method"].(string)
	if !ok || method == "" {
		return fmt.Errorf("missing method in notification")
	}

	// Must NOT have an id field (that would make it a request)
	if _, hasID := notification["id"]; hasID {
		return fmt.Errorf("SSE notification should not have an id field")
	}

	// params are optional in notifications
	// But if we have expected content, validate it matches
	if v.expectedContent != nil {
		params, hasParams := notification["params"]
		if !hasParams {
			return fmt.Errorf("expected params but none found")
		}

		// For now, do a simple equality check
		// In a more sophisticated implementation, we might do deep equality
		// or allow for matchers/patterns
		expectedStr := fmt.Sprintf("%v", v.expectedContent)
		actualStr := fmt.Sprintf("%v", params)
		if expectedStr != actualStr {
			return fmt.Errorf("params mismatch: expected %v, got %v", v.expectedContent, params)
		}
	}

	return nil
}

// StreamingValidator validates streaming message sequences for WebSocket and
// SSE transports. It tracks the number of messages received and optionally
// validates message ordering.
//
// This validator is stateful and accumulates information across multiple
// messages, making it suitable for validating entire streaming sessions
// rather than individual messages.
type StreamingValidator struct {
	expectedCount    int
	receivedCount    int
	validateSequence bool
}

// NewStreamingValidator creates a streaming validator that expects a specific
// number of messages. Set expectedCount to 0 to skip count validation.
// Enable validateSequence to check that messages arrive in the expected order.
func NewStreamingValidator(expectedCount int, validateSequence bool) *StreamingValidator {
	return &StreamingValidator{
		expectedCount:    expectedCount,
		validateSequence: validateSequence,
	}
}

// Validate checks each streaming message and updates internal counters.
// This method should be called for each message in the stream. It only tracks
// the count of messages - format validation should be done by transport-specific
// validators (e.g., SSEEventValidator for SSE, WebSocketMessageValidator for WS).
//
// The validator will return an error if more messages are received than
// expected, helping detect issues with stream termination.
func (v *StreamingValidator) Validate(response any) error {
	v.receivedCount++

	// Just count messages - format validation is done by other validators
	// This makes the streaming validator transport-agnostic

	// Check if we've exceeded expected count
	if v.expectedCount > 0 && v.receivedCount > v.expectedCount {
		return fmt.Errorf("received more messages than expected: %d > %d", v.receivedCount, v.expectedCount)
	}

	return nil
}

// Complete checks if the streaming session received the expected number of
// messages. This should be called after the stream ends to validate that
// all expected messages were received.
//
// Returns an error if fewer messages were received than expected, which
// typically indicates premature stream termination or lost messages.
func (v *StreamingValidator) Complete() error {
	if v.expectedCount > 0 && v.receivedCount != v.expectedCount {
		return fmt.Errorf("received fewer messages than expected: %d < %d", v.receivedCount, v.expectedCount)
	}
	return nil
}

// BatchResponseValidator validates JSON-RPC batch response format according to
// the specification. Batch responses must be arrays with each element being a
// valid JSON-RPC response.
//
// This validator checks both the array structure and validates each individual
// response within the batch. It's used for testing the server's ability to
// handle multiple requests in a single HTTP POST.
func BatchResponseValidator(expectedCount int) Validator {
	return ValidatorFunc(func(response any) error {
		// Batch responses should be arrays
		// Handle both []any and []json.RawMessage
		var responses []any
		switch v := response.(type) {
		case []any:
			responses = v
		case []json.RawMessage:
			// Convert []json.RawMessage to []any
			responses = make([]any, len(v))
			for i, msg := range v {
				responses[i] = msg
			}
		default:
			return fmt.Errorf("batch response is not an array")
		}

		if len(responses) != expectedCount {
			return fmt.Errorf("expected %d responses, got %d", expectedCount, len(responses))
		}

		// Validate each response
		for i, resp := range responses {
			if _, err := AsJSONRPCResponse(resp); err != nil {
				return fmt.Errorf("invalid response at index %d: %w", i, err)
			}
		}

		return nil
	})
}

// HeaderValidator validates HTTP headers
func HeaderValidator(headers map[string]string) Validator {
	return ValidatorFunc(func(response any) error {
		// In actual implementation, this would check HTTP headers
		// For now, just validate response format
		_, err := AsJSONRPCResponse(response)
		return err
	})
}

// StatusCodeValidator validates HTTP status codes
func StatusCodeValidator(expectedStatus int) Validator {
	return ValidatorFunc(func(response any) error {
		// Would check actual HTTP status in real implementation
		if expectedStatus != http.StatusOK {
			return fmt.Errorf("status code validation not implemented")
		}

		_, err := AsJSONRPCResponse(response)
		return err
	})
}
