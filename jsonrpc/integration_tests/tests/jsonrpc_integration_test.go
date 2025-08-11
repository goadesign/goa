package tests

import (
	"path/filepath"
	"testing"

	"goa.design/goa/v3/jsonrpc/integration_tests/framework"
)

// TestJSONRPC is the single entry point for all JSON-RPC integration tests.
// All test scenarios are defined in ../scenarios/scenarios.yaml
func TestJSONRPC(t *testing.T) {
	runner, err := framework.NewRunner(
		filepath.Join("..", "scenarios", "scenarios.yaml"),
		framework.WithParallel(true),
	)
	if err != nil {
		t.Fatalf("Failed to create test runner: %v", err)
	}

	runner.Run(t)
}
