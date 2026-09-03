// This file builds complete generator plans for tests.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

// mustTestPlan chooses package names, finishes each selected generator, and
// fails the calling test if any step is invalid.
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

// testServiceFiles returns service files from the plan under test.
func testServiceFiles(plan *Plan) ([]*codegen.File, error) {
	return serviceFiles(plan)
}

// testTransportFiles returns transport files from the plan under test.
func testTransportFiles(plan *Plan) ([]*codegen.File, error) {
	return transportFiles(plan)
}

// testOpenAPIFiles returns OpenAPI files from the plan under test.
func testOpenAPIFiles(plan *Plan) ([]*codegen.File, error) {
	return openAPIFiles(plan)
}

// assembleExampleFilesForTest returns example files from the plan under test.
func assembleExampleFilesForTest(plan *Plan) ([]*codegen.File, error) {
	return exampleFiles(plan)
}
