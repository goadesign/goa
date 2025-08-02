package harness

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// TestHandler provides a generic handler for integration test scenarios.
// It implements common test patterns like error triggering, validation,
// and streaming behavior based on method names and parameters.
type TestHandler struct{}

// HandleMethod processes a method call and returns appropriate responses
// for integration testing scenarios.
func (h *TestHandler) HandleMethod(ctx context.Context, method string, payload any) (any, error) {
	// Handle error scenarios
	if strings.Contains(method, "test_error") || strings.Contains(method, "error") {
		return h.handleErrorMethod(payload)
	}

	// Handle validation scenarios
	if strings.Contains(method, "validate") {
		return h.handleValidationMethod(payload)
	}

	// Handle standard echo/call methods
	if strings.Contains(method, "call") || strings.Contains(method, "echo") {
		return h.handleEchoMethod(payload)
	}

	// Default: echo back the payload
	return payload, nil
}

// handleErrorMethod returns errors based on trigger parameter
func (h *TestHandler) handleErrorMethod(payload any) (any, error) {
	// Extract trigger from payload
	var trigger string
	switch p := payload.(type) {
	case string:
		trigger = p
	case map[string]any:
		if t, ok := p["trigger"].(string); ok {
			trigger = t
		}
	}

	// Return appropriate error based on trigger
	switch trigger {
	case "parse":
		return nil, &JSONRPCError{Code: -32700, Message: "parse error"}
	case "invalid":
		return nil, &JSONRPCError{Code: -32600, Message: "invalid request"}
	case "method":
		return nil, &JSONRPCError{Code: -32601, Message: "method not found"}
	case "params":
		return nil, &JSONRPCError{Code: -32602, Message: "invalid params"}
	case "internal":
		return nil, &JSONRPCError{Code: -32603, Message: "internal error"}
	case "validation":
		return nil, &JSONRPCError{
			Code:    -32001,
			Message: "validation error",
			Data:    map[string]any{"field": "email", "message": "invalid format"},
		}
	case "notfound":
		return nil, &JSONRPCError{
			Code:    -32002,
			Message: "not found",
			Data:    map[string]any{"resource": "user", "id": "123"},
		}
	case "success":
		return "success", nil
	default:
		return nil, &JSONRPCError{Code: -32603, Message: "internal error"}
	}
}

// handleValidationMethod validates input and returns errors for invalid data
func (h *TestHandler) handleValidationMethod(payload any) (any, error) {
	params, ok := payload.(map[string]any)
	if !ok {
		return nil, &JSONRPCError{Code: -32602, Message: "invalid params"}
	}

	// Check required fields
	if required, exists := params["required_field"]; !exists || required == nil || required == "" {
		return nil, &JSONRPCError{
			Code:    -32602,
			Message: "invalid params",
			Data:    map[string]any{"error": "required_field is required"},
		}
	}

	// Check email format
	if email, exists := params["email"]; exists && email != nil {
		emailStr, ok := email.(string)
		if !ok || !strings.Contains(emailStr, "@") {
			return nil, &JSONRPCError{
				Code:    -32602,
				Message: "invalid params",
				Data:    map[string]any{"error": "email must be a valid email address"},
			}
		}
	}

	// Check URL format
	if url, exists := params["url"]; exists && url != nil {
		urlStr, ok := url.(string)
		if !ok || (!strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://")) {
			return nil, &JSONRPCError{
				Code:    -32602,
				Message: "invalid params",
				Data:    map[string]any{"error": "url must be a valid URL"},
			}
		}
	}

	return map[string]any{"status": "valid"}, nil
}

// handleEchoMethod echoes back the payload with some transformation
func (h *TestHandler) handleEchoMethod(payload any) (any, error) {
	// For object payloads, add a response field
	if params, ok := payload.(map[string]any); ok {
		result := make(map[string]any)
		for k, v := range params {
			result[k] = v
		}
		result["echoed"] = true
		return result, nil
	}

	// For primitive payloads, just echo back
	return payload, nil
}

// JSONRPCError represents a JSON-RPC error response
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error implements the error interface
func (e *JSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// IsJSONRPCError checks if an error is a JSONRPCError
func IsJSONRPCError(err error) (*JSONRPCError, bool) {
	var jErr *JSONRPCError
	if errors.As(err, &jErr) {
		return jErr, true
	}
	return nil, false
}
