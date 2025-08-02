package tests

import (
	"testing"
	
	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
	"goa.design/goa/v3/jsonrpc/integration_tests/scenarios"
)

func TestSingleScenario(t *testing.T) {
	h := harness.New(t)
	
	// Test a specific failing scenario
	matrix := scenarios.GenerateTestMatrix()
	for _, s := range matrix {
		if s.Name == "http_none_payload_map_result" {
			t.Logf("Testing scenario: %s", s.Name)
			
			runner := scenarios.NewScenarioRunner(h)
			if err := runner.Run(s); err != nil {
				t.Fatalf("Scenario failed: %v", err)
			}
			break
		}
	}
}