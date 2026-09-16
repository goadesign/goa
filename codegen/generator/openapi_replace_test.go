// This file verifies that a plugin can replace the OpenAPI documents for the
// exact application design before generation names become final.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
	"goa.design/goa/v3/http/codegen/openapi"
	httpdata "goa.design/goa/v3/http/codegen/testdata"
)

func TestReplaceOpenAPI(t *testing.T) {
	root := codegen.RunDSL(t, httpdata.SimpleDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	plan := &Plan{
		generation:    generation,
		preparedRoots: []eval.Root{root},
		examples:      newExampleGenerators([]eval.Root{root}),
	}
	require.NoError(t, planOpenAPIData(plan))
	replacement, err := httpcodegen.NewOpenAPIPlan(root, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	second, err := httpcodegen.NewOpenAPIPlanFromSpecs(
		root,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		[]openapi.Spec{{Version: openapi.Version20, Path: "localized/openapi"}},
		openapi.Values{},
	)
	require.NoError(t, err)

	require.NoError(t, plan.ReplaceOpenAPI(root, replacement, second))
	files, err := openAPIFiles(plan)
	require.NoError(t, err)
	require.Equal(t, append(replacement.Files(), second.Files()...), files)
}

func TestReplaceOpenAPIRejectsInvalidOwnerPhaseAndPlans(t *testing.T) {
	root := codegen.RunDSL(t, httpdata.SimpleDSL)
	other := codegen.RunDSL(t, httpdata.SimpleDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	plan := &Plan{
		generation:    generation,
		preparedRoots: []eval.Root{root},
		examples:      newExampleGenerators([]eval.Root{root}),
	}
	require.NoError(t, planOpenAPIData(plan))
	replacement, err := httpcodegen.NewOpenAPIPlan(root, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	original, err := openAPIFiles(plan)
	require.NoError(t, err)

	require.ErrorContains(t, plan.ReplaceOpenAPI(nil, replacement), "root is nil")
	require.ErrorContains(t, plan.ReplaceOpenAPI(other, replacement), "not the application design root")
	require.ErrorContains(t, plan.ReplaceOpenAPI(root), "at least one")
	require.ErrorContains(t, plan.ReplaceOpenAPI(root, nil), "plan 0 is nil")
	require.ErrorContains(t, plan.ReplaceOpenAPI(root, replacement, replacement), "same output path")
	upper, err := httpcodegen.NewOpenAPIPlanFromSpecs(
		root,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		[]openapi.Spec{{Version: openapi.Version20, Path: "docs/API"}},
		openapi.Values{},
	)
	require.NoError(t, err)
	lower, err := httpcodegen.NewOpenAPIPlanFromSpecs(
		root,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		[]openapi.Spec{{Version: openapi.Version30, Path: "docs/api"}},
		openapi.Values{},
	)
	require.NoError(t, err)
	require.ErrorContains(t, plan.ReplaceOpenAPI(root, upper, lower), "case-insensitive filesystem")
	unchanged, err := openAPIFiles(plan)
	require.NoError(t, err)
	require.Equal(t, original, unchanged)

	require.NoError(t, generation.Freeze())
	require.ErrorContains(t, plan.ReplaceOpenAPI(root, replacement), "after generation freeze")
}
