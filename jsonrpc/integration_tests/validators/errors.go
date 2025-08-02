package validators

import (
	"encoding/json"
	"fmt"
	"strings"

	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
)

// ErrorValidator returns a validator that checks if an error response matches
// the expected error code and message. This is used to verify that services
// properly return errors in the correct JSON-RPC format.
//
// The expectedMessage parameter can be a substring; the validator will check
// if the actual error message contains this text. This allows for flexible
// matching when exact error messages may vary.
func ErrorValidator(expectedCode int, expectedMessage string) Validator {
	return ValidatorFunc(func(response any) error {
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		if resp.Error == nil {
			return fmt.Errorf("expected error response but got result")
		}

		// Validate error code
		if resp.Error.Code != expectedCode {
			return fmt.Errorf("expected error code %d, got %d", expectedCode, resp.Error.Code)
		}

		// Validate error message (partial match)
		if expectedMessage != "" && !strings.Contains(resp.Error.Message, expectedMessage) {
			return fmt.Errorf("expected error message to contain '%s', got '%s'", expectedMessage, resp.Error.Message)
		}

		return nil
	})
}

// StandardErrorValidator returns a validator for standard JSON-RPC error codes.
// It accepts an error type string and validates the corresponding error code:
//   - "parse": -32700 (Parse error)
//   - "invalid": -32600 (Invalid Request)
//   - "method": -32601 (Method not found)
//   - "params": -32602 (Invalid params)
//   - "internal": -32603 (Internal error)
//
// This simplifies testing of standard JSON-RPC errors without needing to
// remember specific error codes.
func StandardErrorValidator(errorType string) Validator {
	errorCodes := map[string]int{
		"parse":    -32700,
		"invalid":  -32600,
		"method":   -32601,
		"params":   -32602,
		"internal": -32603,
	}

	code, exists := errorCodes[errorType]
	if !exists {
		return ValidatorFunc(func(response any) error {
			return fmt.Errorf("unknown standard error type: %s", errorType)
		})
	}

	return ErrorValidator(code, "")
}

// CustomErrorValidator validates application-specific errors with exact matching
// of error code, message, and optional data fields. This validator is stricter
// than ErrorValidator, requiring exact message matches and supporting data field
// validation.
//
// Use this validator when testing custom application errors that include
// structured data in the error response, such as validation details or
// debug information.
func CustomErrorValidator(expectedError harness.ErrorObject) Validator {
	return ValidatorFunc(func(response any) error {
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		if resp.Error == nil {
			return fmt.Errorf("expected error response but got result")
		}

		// Validate error code
		if resp.Error.Code != expectedError.Code {
			return fmt.Errorf("expected error code %d, got %d", expectedError.Code, resp.Error.Code)
		}

		// Validate error message
		if expectedError.Message != "" && resp.Error.Message != expectedError.Message {
			return fmt.Errorf("expected error message '%s', got '%s'", expectedError.Message, resp.Error.Message)
		}

		// Validate error data if present
		if expectedError.Data != nil {
			if resp.Error.Data == nil {
				return fmt.Errorf("expected error data but none present")
			}

			// Compare data structures - handle different types
			var expectedData, actualData any

			// Handle expected data
			switch v := expectedError.Data.(type) {
			case []byte:
				if err := json.Unmarshal(v, &expectedData); err != nil {
					return fmt.Errorf("failed to unmarshal expected error data: %w", err)
				}
			case json.RawMessage:
				if err := json.Unmarshal(v, &expectedData); err != nil {
					return fmt.Errorf("failed to unmarshal expected error data: %w", err)
				}
			default:
				expectedData = v
			}

			// Handle actual data
			switch v := resp.Error.Data.(type) {
			case []byte:
				if err := json.Unmarshal(v, &actualData); err != nil {
					return fmt.Errorf("failed to unmarshal actual error data: %w", err)
				}
			case json.RawMessage:
				if err := json.Unmarshal(v, &actualData); err != nil {
					return fmt.Errorf("failed to unmarshal actual error data: %w", err)
				}
			default:
				actualData = v
			}

			// Basic comparison - could be enhanced
			if fmt.Sprintf("%v", expectedData) != fmt.Sprintf("%v", actualData) {
				return fmt.Errorf("error data mismatch")
			}
		}

		return nil
	})
}

// ValidationErrorValidator returns a validator that checks for input validation
// errors (typically -32602 Invalid params). If expectedField is provided, it
// verifies that the error message or data mentions the specific field that
// failed validation.
//
// This validator is specialized for testing parameter validation failures,
// ensuring that validation errors are properly reported with appropriate
// error codes and field-specific information when available. It's particularly
// useful for testing that validation errors provide helpful information
// about which field caused the validation failure.
func ValidationErrorValidator(expectedField string) Validator {
	return ValidatorFunc(func(response any) error {
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		if resp.Error == nil {
			return fmt.Errorf("expected validation error but got result")
		}

		// Check for invalid params error code
		if resp.Error.Code != -32602 {
			return fmt.Errorf("expected invalid params error (-32602), got %d", resp.Error.Code)
		}

		// Check if error mentions the expected field
		if expectedField != "" && !strings.Contains(resp.Error.Message, expectedField) {
			// Also check error data
			if resp.Error.Data != nil {
				var data any
				// Handle different data types
				switch v := resp.Error.Data.(type) {
				case []byte:
					json.Unmarshal(v, &data)
				case json.RawMessage:
					json.Unmarshal(v, &data)
				default:
					data = v
				}
				dataStr := fmt.Sprintf("%v", data)
				if !strings.Contains(dataStr, expectedField) {
					return fmt.Errorf("validation error should mention field '%s'", expectedField)
				}
			} else {
				return fmt.Errorf("validation error should mention field '%s'", expectedField)
			}
		}

		return nil
	})
}

// ErrorCodeRangeValidator validates that error codes fall within an expected
// range. This is useful for ensuring that custom application errors use
// appropriate error code ranges and don't conflict with reserved JSON-RPC
// error codes.
//
// For example, the JSON-RPC specification reserves -32768 to -32000 for
// predefined errors. Application errors should typically use other ranges
// to avoid conflicts. This validator helps enforce such conventions.
func ErrorCodeRangeValidator(minCode, maxCode int) Validator {
	return ValidatorFunc(func(response any) error {
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		if resp.Error == nil {
			return nil // Not an error response
		}

		if resp.Error.Code < minCode || resp.Error.Code > maxCode {
			return fmt.Errorf("error code %d outside expected range [%d, %d]", resp.Error.Code, minCode, maxCode)
		}

		return nil
	})
}
