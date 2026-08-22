// This file provides strict construction helpers for generator lifecycle tests
// whose package roots and planning claims are deliberately valid.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

// mustTestPlan runs the production declaration, freeze, and link lifecycle for
// focused assembler tests and fails the calling test on any invalid phase.
func mustTestPlan(t *testing.T, genpkg string, roots []eval.Root, planners ...func(*Plan) error) *Plan {
	t.Helper()
	generation, err := codegen.NewGeneration(genpkg, roots)
	require.NoError(t, err)
	plan := &Plan{
		generation:    generation,
		preparedRoots: roots,
		examples:      newExampleGenerators(roots),
	}
	for _, planner := range planners {
		require.NoError(t, planner(plan))
	}
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.link())
	return plan
}

// testServiceFiles renders service files from the retained plan under test.
func testServiceFiles(plan *Plan) ([]*codegen.File, error) {
	return serviceFiles(plan)
}

// testTransportFiles renders transport files from the retained plan under test.
func testTransportFiles(plan *Plan) ([]*codegen.File, error) {
	return transportFiles(plan)
}

// testOpenAPIFiles renders OpenAPI files from the retained plan under test.
func testOpenAPIFiles(plan *Plan) ([]*codegen.File, error) {
	return openAPIFiles(plan)
}

// assembleExampleFilesForTest renders example files from the retained plan.
func assembleExampleFilesForTest(plan *Plan) ([]*codegen.File, error) {
	return exampleFiles(plan)
}
