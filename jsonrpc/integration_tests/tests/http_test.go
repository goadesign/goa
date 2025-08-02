package tests

import (
	"testing"

	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
	"goa.design/goa/v3/jsonrpc/integration_tests/scenarios"
	"goa.design/goa/v3/jsonrpc/integration_tests/validators"
)

// TestHTTPBasic tests basic HTTP JSON-RPC functionality using a small set of
// quick test scenarios. This test ensures fundamental request/response patterns
// work correctly over HTTP transport.
//
// The test validates basic method calls, parameter passing, and result handling
// without exhaustive type coverage. It's designed to catch obvious regressions
// quickly during development.
func TestHTTPBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Get quick test scenarios for HTTP
	quickScenarios := scenarios.QuickTestScenarios()

	// Create runner
	runner := scenarios.NewScenarioRunner(h)

	for _, scenario := range quickScenarios {
		if scenario.Transport != scenarios.TransportHTTP {
			continue
		}

		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel() // Run scenarios in parallel

			// Add standard validators
			scenario.Validators = validators.StandardValidators()

			// Run scenario
			if err := runner.Run(scenario); err != nil {
				t.Fatalf("Scenario failed: %v", err)
			}
		})
	}
}

// TestHTTPMatrix tests all HTTP transport combinations systematically using
// the complete test matrix. This comprehensive test validates every combination
// of payload types and result types to ensure thorough coverage.
//
// The test applies appropriate validators based on scenario features and data
// types. This catches edge cases and ensures all type combinations work correctly.
func TestHTTPMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Generate full test matrix
	matrix := scenarios.GenerateTestMatrix()

	// Run HTTP scenarios
	runner := scenarios.NewScenarioRunner(h)

	for _, scenario := range matrix {
		if scenario.Transport != scenarios.TransportHTTP {
			continue
		}

		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel() // Run scenarios in parallel

			// Add validators based on features
			scenario.Validators = getValidatorsForScenario(scenario)

			// Run scenario
			if err := runner.Run(scenario); err != nil {
				t.Fatalf("Scenario %s failed: %v", scenario.Name, err)
			}
		})
	}
}

// TestHTTPErrors tests error handling over HTTP transport, focusing on scenarios
// that should produce JSON-RPC errors. This validates that the framework properly
// converts service errors to JSON-RPC error responses.
//
// The test checks error codes, messages, and the overall error response structure
// to ensure compliance with the JSON-RPC specification. It covers standard errors
// like invalid params and method not found.
func TestHTTPErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Get error scenarios
	matrix := scenarios.GenerateTestMatrix()

	// Run error scenarios
	runner := scenarios.NewScenarioRunner(h)

	for _, scenario := range matrix {
		if scenario.Transport != scenarios.TransportHTTP {
			continue
		}

		// Only run error scenarios
		hasErrors := false
		for _, feature := range scenario.Features {
			if feature == scenarios.FeatureErrors {
				hasErrors = true
				break
			}
		}
		if !hasErrors {
			continue
		}

		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel() // Run scenarios in parallel

			// Add error validators
			scenario.Validators = append(
				validators.StandardValidators(),
				validators.ErrorCodeRangeValidator(-32768, -32000),
			)

			// Run scenario
			if err := runner.Run(scenario); err != nil {
				t.Fatalf("Error scenario failed: %v", err)
			}
		})
	}
}

// TestHTTPValidation tests input validation for HTTP JSON-RPC requests. This
// ensures that the framework properly validates request parameters according to
// the service definitions and returns appropriate validation errors.
//
// The test covers required fields, format validation, and type checking. It
// verifies that validation failures result in proper JSON-RPC error responses
// with the -32602 (Invalid params) error code.
func TestHTTPValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Get validation scenarios
	matrix := scenarios.GenerateTestMatrix()

	// Run validation scenarios
	runner := scenarios.NewScenarioRunner(h)

	for _, scenario := range matrix {
		if scenario.Transport != scenarios.TransportHTTP {
			continue
		}

		// Only run validation scenarios
		hasValidation := false
		for _, feature := range scenario.Features {
			if feature == scenarios.FeatureValidation {
				hasValidation = true
				break
			}
		}
		if !hasValidation {
			continue
		}

		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel() // Run scenarios in parallel

			// Add validation-specific validators
			scenario.Validators = validators.StandardValidators()

			// Run scenario
			if err := runner.Run(scenario); err != nil {
				t.Fatalf("Validation scenario failed: %v", err)
			}
		})
	}
}

// TestHTTPBatch tests batch request handling over HTTP transport. According to
// the JSON-RPC specification, clients can send multiple requests in a single
// HTTP POST as an array.
//
// This test validates that the server correctly processes batch requests,
// returning an array of responses that correspond to each request in the batch.
// It also tests error handling within batches and mixed success/failure scenarios.
func TestHTTPBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// Create test harness
	h := harness.New(t)

	// Get batch scenarios
	matrix := scenarios.GenerateTestMatrix()

	// Run batch scenarios
	runner := scenarios.NewScenarioRunner(h)

	for _, scenario := range matrix {
		if scenario.Transport != scenarios.TransportHTTP {
			continue
		}

		// Only run batch scenarios
		hasBatch := false
		for _, feature := range scenario.Features {
			if feature == scenarios.FeatureBatch {
				hasBatch = true
				break
			}
		}
		if !hasBatch {
			continue
		}

		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel() // Run scenarios in parallel

			// Add batch validators
			// For batch responses, don't use standard validators as they expect single responses
			scenario.Validators = []validators.Validator{
				validators.BatchResponseValidator(2), // Expecting 2 responses
			}

			// Run scenario
			if err := runner.Run(scenario); err != nil {
				t.Fatalf("Batch scenario failed: %v", err)
			}
		})
	}
}

// getValidatorsForScenario returns appropriate validators based on scenario features
// and data types. This helper function builds a comprehensive set of validators
// tailored to each test scenario's specific requirements.
//
// The function starts with standard validators (protocol compliance, data integrity)
// then adds feature-specific validators for errors, validation, batch requests, and
// views. Finally, it adds type-specific validators based on the expected result type.
// This modular approach ensures each scenario is validated thoroughly without
// unnecessary checks.
func getValidatorsForScenario(scenario scenarios.Scenario) []validators.Validator {
	// Check if this is a batch scenario
	isBatch := false
	for _, feature := range scenario.Features {
		if feature == scenarios.FeatureBatch {
			isBatch = true
			break
		}
	}

	// For batch scenarios, don't use standard validators as they expect single responses
	var vals []validators.Validator
	if !isBatch {
		vals = validators.StandardValidators()
	}

	// Add feature-specific validators
	for _, feature := range scenario.Features {
		switch feature {
		case scenarios.FeatureErrors:
			vals = append(vals, validators.ErrorCodeRangeValidator(-32768, -32000))

		case scenarios.FeatureValidation:
			// For validation scenarios, don't add a blanket error validator
			// The scenario runner will validate each request individually based on ExpectedError
			// Just add the error code range validator to check error format when errors do occur
			vals = append(vals, validators.ErrorCodeRangeValidator(-32768, -32000))

		case scenarios.FeatureBatch:
			vals = append(vals, validators.BatchResponseValidator(2))

		case scenarios.FeatureViews:
			// Views might have specific field requirements
			// Use JSON field names (lowercase) not Go struct field names (uppercase)
			vals = append(vals, validators.RequiredFieldsValidator([]string{"id", "name"}))
		}
	}

	// Add data type validators only for scenarios that have consistent result types
	// Error and validation scenarios should not validate result types since they have mixed responses
	hasErrors := false
	hasValidation := false
	for _, feature := range scenario.Features {
		if feature == scenarios.FeatureErrors {
			hasErrors = true
		}
		if feature == scenarios.FeatureValidation {
			hasValidation = true
		}
	}

	if !hasErrors && !hasValidation && !isBatch {
		switch scenario.ResultType {
		case scenarios.DataTypePrimitive:
			vals = append(vals, validators.TypeValidator("string"))

		case scenarios.DataTypeArray:
			vals = append(vals, validators.TypeValidator([]any{}))

		case scenarios.DataTypeObject:
			vals = append(vals, validators.TypeValidator(map[string]any{}))

		case scenarios.DataTypeUserType:
			// Use JSON field names (lowercase) not Go struct field names (uppercase)
			vals = append(vals, validators.RequiredFieldsValidator([]string{"id", "name"}))
		}
	}

	return vals
}
