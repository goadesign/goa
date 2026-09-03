// This file verifies transport generators can retain the exact generated
// service and views package preferences chosen by a service plan.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestMethodPackageImportsReturnsPlannedServiceAndViewsPackages catches HTTP
// generators guessing a service or views package from their own output path.
func TestMethodPackageImportsReturnsPlannedServiceAndViewsPackages(t *testing.T) {
	root := expr.RunDSL(t, func() {
		result := dsl.ResultType("application/vnd.storage.item", func() {
			dsl.TypeName("Item")
			dsl.Attribute("name", dsl.String)
			dsl.View("default", func() {
				dsl.Attribute("name")
			})
		})
		dsl.Service("Storage", func() {
			dsl.Method("Show", func() {
				dsl.Result(result)
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)

	servicePackage, viewsPackage, err := plan.MethodPackageImports(root.Service("Storage").Method("Show"))
	require.NoError(t, err)
	require.Equal(t, &codegen.ImportSpec{Name: "storage", Path: "generated.local/gen/storage"}, servicePackage)
	require.Equal(t, &codegen.ImportSpec{Name: "storageviews", Path: "generated.local/gen/storage/views"}, viewsPackage)

	servicePackage, viewsPackage, err = plan.ServicePackageImports(root.Service("Storage"))
	require.NoError(t, err)
	require.Equal(t, &codegen.ImportSpec{Name: "storage", Path: "generated.local/gen/storage"}, servicePackage)
	require.Equal(t, &codegen.ImportSpec{Name: "storageviews", Path: "generated.local/gen/storage/views"}, viewsPackage)
}
