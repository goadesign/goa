// This file provides strict construction helpers for generator lifecycle tests
// whose package roots and planning claims are deliberately valid.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

// mustTestGeneration creates one generation or fails the calling test.
func mustTestGeneration(t *testing.T, genpkg string, roots []eval.Root) *codegen.Generation {
	t.Helper()
	generation, err := codegen.NewGeneration(genpkg, roots)
	require.NoError(t, err)
	return generation
}

// testServiceFiles adapts the private plan-owned service assembler for package tests.
func testServiceFiles(generation *codegen.Generation) ([]*codegen.File, error) {
	return serviceFiles(testPlan(generation))
}

// testTransportFiles adapts the private plan-owned transport assembler for package tests.
func testTransportFiles(generation *codegen.Generation) ([]*codegen.File, error) {
	return transportFiles(testPlan(generation))
}

// testOpenAPIFiles adapts the private plan-owned OpenAPI assembler for package tests.
func testOpenAPIFiles(generation *codegen.Generation) ([]*codegen.File, error) {
	return openAPIFiles(testPlan(generation))
}

// assembleExampleFilesForTest adapts the private plan-owned example assembler
// for package tests.
func assembleExampleFilesForTest(generation *codegen.Generation) ([]*codegen.File, error) {
	return exampleFiles(testPlan(generation))
}

// testPlan creates the run-only state needed by a focused assembler test.
func testPlan(generation *codegen.Generation) *Plan {
	return &Plan{
		generation: generation,
		examples:   newExampleGenerators(generation.Roots()),
	}
}
