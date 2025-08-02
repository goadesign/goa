package validators

import (
	"encoding/json"

	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
)

// Validator is the interface for response validators that verify JSON-RPC
// responses meet expected criteria. Validators can check protocol compliance,
// data integrity, error formats, or any custom validation logic.
//
// Each validator focuses on a specific aspect of the response, allowing
// tests to compose multiple validators for comprehensive verification.
type Validator interface {
	Validate(response any) error
}

// ValidatorFunc is a function adapter for the Validator interface, allowing
// simple validation functions to be used wherever a Validator is needed.
// This simplifies creating one-off validators for specific test scenarios
// without defining new types.
type ValidatorFunc func(response any) error

// Validate implements the Validator interface by calling the wrapped function.
// This allows ValidatorFunc to satisfy the Validator interface.
func (f ValidatorFunc) Validate(response any) error {
	return f(response)
}

// CompositeValidator combines multiple validators into a single validator
// that runs each validator in sequence. This enables building complex
// validation logic from simpler, reusable components.
//
// Validation stops at the first error, making error messages more focused
// and debugging easier.
type CompositeValidator struct {
	validators []Validator
}

// NewCompositeValidator creates a new composite validator from the provided
// validators. The validators are executed in the order provided, with
// validation stopping at the first error.
//
// This is useful for combining standard validators with scenario-specific
// ones to create comprehensive test assertions.
func NewCompositeValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{
		validators: validators,
	}
}

// Validate runs all validators in sequence, stopping at the first error.
// This ensures that error messages are specific to the first validation
// failure, making test failures easier to diagnose.
//
// The response parameter is passed to each validator unchanged, allowing
// different validators to examine different aspects of the same response.
func (v *CompositeValidator) Validate(response any) error {
	for _, validator := range v.validators {
		if err := validator.Validate(response); err != nil {
			return err
		}
	}
	return nil
}

// StandardValidators returns a set of validators that should be applied to
// most JSON-RPC responses. This includes protocol compliance validation and
// basic data integrity checks.
//
// Tests typically start with these validators and add scenario-specific ones
// as needed.
func StandardValidators() []Validator {
	return []Validator{
		ProtocolValidator(),
		DataIntegrityValidator(),
	}
}

// Helper functions for working with responses

// AsJSONRPCResponse converts a generic response to a typed JSON-RPC response
// structure. This helper handles various input types including already-typed
// responses, raw JSON data, and generic interfaces.
//
// The function is useful in validators that need to examine specific JSON-RPC
// fields like error codes or result structures. It provides a consistent way
// to access response data regardless of how it was originally captured.
func AsJSONRPCResponse(response any) (*harness.Response, error) {
	switch r := response.(type) {
	case *harness.Response:
		return r, nil
	case harness.Response:
		return &r, nil
	case json.RawMessage:
		// Handle json.RawMessage directly
		var resp harness.Response
		if err := json.Unmarshal(r, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	default:
		// Try to unmarshal if it's raw JSON
		data, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}
		var resp harness.Response
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}
}
