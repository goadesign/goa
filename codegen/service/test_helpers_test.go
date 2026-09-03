// This file provides strict construction helpers for service code-generation
// tests whose package roots and planner claims are deliberately valid.
package service

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// mustTestGeneration creates one generation or fails the calling test.
func mustTestGeneration(t *testing.T, genpkg string, roots []eval.Root) *codegen.Generation {
	t.Helper()
	generation, err := codegen.NewGeneration(genpkg, roots)
	require.NoError(t, err)
	return generation
}

// mustClaimTestPackage claims one valid planner path or fails the calling test.
func mustClaimTestPackage(t *testing.T, generation *codegen.Generation, path string) *codegen.GeneratedPackage {
	t.Helper()
	generatedPackage, err := generation.ClaimPackage(path)
	require.NoError(t, err)
	return generatedPackage
}

// planTestServices collects one retained service plan when the test exercises
// declaration collection separately from post-freeze linking.
func planTestServices(root *expr.RootExpr, generation *codegen.Generation) error {
	_, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	return err
}

// servicePackagePath returns the natural generated package path used by tests
// that build one noncolliding service package directly.
func servicePackagePath(genpkg string, service *expr.ServiceExpr) string {
	return path.Join(genpkg, servicePackageName(service.Name))
}
