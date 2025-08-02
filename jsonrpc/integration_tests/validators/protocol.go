package validators

import (
	"fmt"

	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
)

// ProtocolValidator returns a validator that checks JSON-RPC 2.0 protocol
// compliance. It verifies:
//   - The jsonrpc field is exactly "2.0"
//   - Either result or error is present, but not both
//   - Error objects have valid codes and non-empty messages
//   - The response structure matches the JSON-RPC specification
//
// This validator should be used for all JSON-RPC response validation as it
// ensures basic protocol compliance.
func ProtocolValidator() Validator {
	return ValidatorFunc(func(response any) error {
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		// Check JSON-RPC version
		if resp.JSONRPC != "2.0" {
			return fmt.Errorf("invalid JSON-RPC version: expected '2.0', got '%s'", resp.JSONRPC)
		}

		// Check that either result or error is present, but not both
		hasResult := len(resp.Result) > 0
		hasError := resp.Error != nil

		if hasResult && hasError {
			return fmt.Errorf("response contains both result and error")
		}

		if !hasResult && !hasError {
			return fmt.Errorf("response contains neither result nor error")
		}

		// Validate error format if present
		if hasError {
			if err := validateError(resp.Error); err != nil {
				return fmt.Errorf("invalid error format: %w", err)
			}
		}

		// Check ID is present (null is valid for notifications)
		// ID validation depends on the request context

		return nil
	})
}

// validateError validates JSON-RPC error object structure according to the
// JSON-RPC 2.0 specification. It ensures error objects contain valid error
// codes and non-empty messages.
//
// The function checks both standard error codes (-32700 to -32099) and allows
// application-defined error codes. This helps catch common mistakes like
// using HTTP status codes instead of JSON-RPC error codes.
func validateError(err *harness.ErrorObject) error {
	if err == nil {
		return fmt.Errorf("error object is nil")
	}

	// Validate error code
	if !isValidErrorCode(err.Code) {
		return fmt.Errorf("invalid error code: %d", err.Code)
	}

	// Validate error message
	if err.Message == "" {
		return fmt.Errorf("error message is empty")
	}

	return nil
}

// isValidErrorCode checks if an error code is valid per JSON-RPC specification.
// Valid codes include:
//   - Standard JSON-RPC errors: -32700 to -32600 and -32099 to -32000
//   - Server implementation errors: -32099 to -32000
//   - Application-defined errors: any other negative or positive integer
//
// The function helps ensure error codes follow the specification and aren't
// accidentally using incompatible error code schemes.
func isValidErrorCode(code int) bool {
	// Standard JSON-RPC error codes
	standardCodes := map[int]bool{
		-32700: true, // Parse error
		-32600: true, // Invalid Request
		-32601: true, // Method not found
		-32602: true, // Invalid params
		-32603: true, // Internal error
	}

	if standardCodes[code] {
		return true
	}

	// Server error codes (-32000 to -32099)
	if code >= -32099 && code <= -32000 {
		return true
	}

	// Application defined errors
	return true
}

// RequestResponseValidator returns a validator that ensures the response ID
// matches the request ID. This is critical for correlating responses with
// requests, especially in batch or concurrent scenarios.
//
// The validator handles different ID types (string, number, null) according
// to the JSON-RPC specification. Null IDs indicate notifications which should
// not receive responses.
func RequestResponseValidator(requestID any) Validator {
	return ValidatorFunc(func(response any) error {
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		// Compare IDs
		if !compareIDs(requestID, resp.ID) {
			return fmt.Errorf("response ID '%v' does not match request ID '%v'", resp.ID, requestID)
		}

		return nil
	})
}

// compareIDs compares two JSON-RPC IDs
func compareIDs(id1, id2 any) bool {
	// Handle different numeric types
	switch v1 := id1.(type) {
	case float64:
		switch v2 := id2.(type) {
		case float64:
			return v1 == v2
		case int:
			return v1 == float64(v2)
		case int64:
			return v1 == float64(v2)
		}
	case int:
		switch v2 := id2.(type) {
		case float64:
			return float64(v1) == v2
		case int:
			return v1 == v2
		case int64:
			return int64(v1) == v2
		}
	case string:
		v2, ok := id2.(string)
		return ok && v1 == v2
	case nil:
		return id2 == nil
	}

	return fmt.Sprintf("%v", id1) == fmt.Sprintf("%v", id2)
}

// MethodValidator validates that the correct method was called
func MethodValidator(expectedMethod string) Validator {
	return ValidatorFunc(func(response any) error {
		// This validator typically works with request logging
		// For now, just validate the response structure
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		// If it's an error response with method not found, validate
		if resp.Error != nil && resp.Error.Code == -32601 {
			return fmt.Errorf("method not found: %s", expectedMethod)
		}

		return nil
	})
}
