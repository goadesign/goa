package validators

import (
	"encoding/json"
	"fmt"
	"reflect"

	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
)

// DataIntegrityValidator returns a validator that performs basic data integrity
// checks on JSON-RPC responses. It ensures:
//   - Result data (if present) is valid JSON
//   - Error data fields (if present) contain valid JSON
//   - No data corruption occurred during transport
//
// This validator is part of StandardValidators() and provides a baseline check
// that response data can be properly decoded.
func DataIntegrityValidator() Validator {
	return ValidatorFunc(func(response any) error {
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		// If error response, validate error data integrity
		if resp.Error != nil {
			return validateErrorData(resp.Error)
		}

		// For result, basic validation that it's valid JSON
		if len(resp.Result) > 0 {
			var data any
			if err := json.Unmarshal(resp.Result, &data); err != nil {
				return fmt.Errorf("result is not valid JSON: %w", err)
			}
		}

		return nil
	})
}

// validateErrorData validates the optional data field in JSON-RPC error objects.
// The data field can contain additional information about the error and must be
// valid JSON if present.
//
// This validation ensures that error data can be properly decoded by clients,
// preventing issues with malformed error details that could break error handling
// logic.
func validateErrorData(errObj *harness.ErrorObject) error {
	if errObj.Data != nil {
		// Data can be any JSON value
		if dataBytes, ok := errObj.Data.([]byte); ok {
			var data any
			if err := json.Unmarshal(dataBytes, &data); err != nil {
				return fmt.Errorf("error data is not valid JSON: %w", err)
			}
		} else if dataBytes, ok := errObj.Data.(json.RawMessage); ok {
			var data any
			if err := json.Unmarshal(dataBytes, &data); err != nil {
				return fmt.Errorf("error data is not valid JSON: %w", err)
			}
		}
		// If it's already unmarshaled, that's fine too
	}
	return nil
}

// TypeValidator returns a validator that checks if the response result matches
// the expected type structure. It performs recursive type checking to ensure
// that the JSON-decoded result has the correct types for all fields.
//
// The expectedType parameter should be a value with the expected structure,
// for example:
//   - "string" for primitive string results
//   - []any{} for array results
//   - map[string]any{"field": "string"} for object results
//
// The validator accounts for JSON type conversions (e.g., all numbers become
// float64) and validates nested structures recursively.
func TypeValidator(expectedType any) Validator {
	return ValidatorFunc(func(response any) error {
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		if resp.Error != nil {
			return fmt.Errorf("expected result but got error: %s", resp.Error.Message)
		}

		// Unmarshal result
		var result any
		if err := json.Unmarshal(resp.Result, &result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}

		// Validate type structure
		return validateTypeStructure(result, expectedType)
	})
}

// validateTypeStructure recursively validates that the actual data structure
// matches the expected type structure. This function handles JSON type
// conversions (e.g., all numbers become float64) and validates nested
// structures.
//
// The validation is structural rather than value-based - it checks that
// fields exist and have the correct types, not that values match exactly.
// This makes it suitable for validating response shapes in integration tests.
func validateTypeStructure(actual, expected any) error {
	actualType := reflect.TypeOf(actual)

	// Handle nil
	if expected == nil {
		if actual != nil {
			return fmt.Errorf("expected nil, got %T", actual)
		}
		return nil
	}

	// Special handling for JSON numbers (float64)
	if isNumeric(expected) && isNumeric(actual) {
		return nil // JSON numbers are always float64
	}

	// Check basic type compatibility
	switch expected.(type) {
	case string:
		if _, ok := actual.(string); !ok {
			return fmt.Errorf("expected string, got %T", actual)
		}

	case bool:
		if _, ok := actual.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", actual)
		}

	case []any:
		actualSlice, ok := actual.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", actual)
		}

		// If expected has elements, validate first element type
		expectedSlice := expected.([]any)
		if len(expectedSlice) > 0 && len(actualSlice) > 0 {
			return validateTypeStructure(actualSlice[0], expectedSlice[0])
		}

	case map[string]any:
		actualMap, ok := actual.(map[string]any)
		if !ok {
			return fmt.Errorf("expected object, got %T", actual)
		}

		// Validate each expected field
		expectedMap := expected.(map[string]any)
		for key, expectedValue := range expectedMap {
			actualValue, exists := actualMap[key]
			if !exists && expectedValue != nil {
				return fmt.Errorf("missing expected field: %s", key)
			}
			if exists {
				if err := validateTypeStructure(actualValue, expectedValue); err != nil {
					return fmt.Errorf("field %s: %w", key, err)
				}
			}
		}

	default:
		// For other types, just check they're not nil
		if actualType == nil {
			return fmt.Errorf("expected %T, got nil", expected)
		}
	}

	return nil
}

// isNumeric checks if a value is numeric (any integer or float type).
// This helper is used to handle JSON's number representation where all
// numbers are decoded as float64, regardless of their original type.
//
// The function helps the type validator accept any numeric type when
// comparing expected vs actual values, avoiding false negatives due to
// JSON's type system limitations.
func isNumeric(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	}
	return false
}

// RequiredFieldsValidator returns a validator that checks if all required
// fields are present in the response result. This is useful for validating
// that generated code properly includes all required fields defined in the DSL.
//
// The validator only checks object results; it silently passes for non-object
// results. Missing required fields cause validation to fail with a descriptive
// error message.
func RequiredFieldsValidator(requiredFields []string) Validator {
	return ValidatorFunc(func(response any) error {
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		if resp.Error != nil {
			return nil // Skip for error responses
		}

		// Unmarshal result as object
		var result map[string]any
		if err := json.Unmarshal(resp.Result, &result); err != nil {
			// Not an object, can't validate fields
			return nil
		}

		// Check required fields
		for _, field := range requiredFields {
			if _, exists := result[field]; !exists {
				return fmt.Errorf("missing required field: %s", field)
			}
		}

		return nil
	})
}

// RangeValidator validates numeric values are within expected ranges
func RangeValidator(field string, min, max float64) Validator {
	return ValidatorFunc(func(response any) error {
		resp, err := AsJSONRPCResponse(response)
		if err != nil {
			return fmt.Errorf("invalid response type: %w", err)
		}

		if resp.Error != nil {
			return nil // Skip for error responses
		}

		// Unmarshal result
		var result map[string]any
		if err := json.Unmarshal(resp.Result, &result); err != nil {
			return nil // Not an object
		}

		// Get field value
		value, exists := result[field]
		if !exists {
			return nil // Field doesn't exist, skip
		}

		// Convert to float64
		var numValue float64
		switch v := value.(type) {
		case float64:
			numValue = v
		case int:
			numValue = float64(v)
		default:
			return fmt.Errorf("field %s is not numeric: %T", field, value)
		}

		// Validate range
		if numValue < min || numValue > max {
			return fmt.Errorf("field %s value %f is outside range [%f, %f]", field, numValue, min, max)
		}

		return nil
	})
}
