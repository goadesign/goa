// These tests import layouts into new packages with no incidental earlier
// registration, so service, shared and view preferences must come from the binder.
package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestServiceLayoutsRetainImportPreferences(t *testing.T) {
	var located expr.UserType
	root := codegen.RunDSL(t, func() {
		located = dsl.Type("Located", func() {
			dsl.Meta("struct:pkg:path", "APIKeyService")
			dsl.Attribute("value", dsl.String)
		})
		for _, name := range []string{"read-value", "read_value", "read_value2"} {
			dsl.Service(name, func() {
				dsl.Method("read", func() {
					dsl.Payload(func() {
						dsl.Attribute("located", located)
					})
				})
			})
		}
	})
	generation, plans := inheritedPlacementPlans(t, []*expr.RootExpr{root})
	outputs := make(map[*serviceFacts]*codegen.GeneratedPackage)
	for index, facts := range plans[0].facts.services {
		output := mustClaimTestPackage(t, generation, fmt.Sprintf("generated.local/gen/isolated%d", index))
		outputs[facts] = output
		imports := codegen.NewGeneratedImportPlan(output)
		method := facts.methods[0]
		layout, err := plans[0].MethodTypeLayout(method, method.Payload)
		require.NoError(t, err)
		require.NoError(t, imports.AddCompleteType(layout))
	}
	require.NoError(t, generation.Freeze())
	for facts, output := range outputs {
		// An isolated output has no competing service qualifier. Even a service
		// with a suffixed directory keeps the binder's original preference.
		require.Equal(t, facts.packageImport.Name, output.ImportName(facts.packagePath))
		require.Equal(t, "apikeyservice", output.ImportName("generated.local/gen/APIKeyService"))
	}
}

func TestViewLayoutRetainsImportPreference(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		result := dsl.ResultType("application/vnd.record", func() {
			dsl.TypeName("Record")
			dsl.Attribute("value", dsl.String)
			dsl.View("default", func() {
				dsl.Attribute("value")
			})
		})
		dsl.Service("APIKeyService", func() {
			dsl.Method("read", func() {
				dsl.Result(result)
			})
		})
	})
	generation, plans := inheritedPlacementPlans(t, []*expr.RootExpr{root})
	service := root.Services[0]
	method := service.Methods[0]
	projected, err := plans[0].ProjectedResult(method)
	require.NoError(t, err)
	layout, err := plans[0].MethodTypeLayout(method, projected)
	require.NoError(t, err)
	output := mustClaimTestPackage(t, generation, "generated.local/gen/isolated")
	imports := codegen.NewGeneratedImportPlan(output)
	require.NoError(t, imports.AddCompleteType(layout))
	require.NoError(t, generation.Freeze())
	facts := plans[0].facts.serviceByID[service.Name]
	require.Equal(t, facts.viewsImport.Name, output.ImportName(facts.viewsPath))
	require.Equal(t, "apikeyserviceviews", output.ImportName(facts.viewsPath))
}
