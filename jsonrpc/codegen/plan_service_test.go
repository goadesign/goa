// This file checks that plugins can read only the finalized JSON-RPC service
// data that belongs to the exact service expression used to build a plan.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

func TestPlanServiceUsesExactExpressionAfterLink(t *testing.T) {
	generation, roots, services, jsonPlans, applicationPlans := jsonRPCPlanningInputs(t)
	plans, err := NewPlans(
		generation,
		PlanInput{
			Root:            roots[0],
			Service:         services[0],
			HTTP:            jsonPlans[0],
			ApplicationHTTP: applicationPlans[0],
		},
		PlanInput{
			Root:            roots[1],
			Service:         services[1],
			HTTP:            jsonPlans[1],
			ApplicationHTTP: applicationPlans[1],
		},
	)
	require.NoError(t, err)
	require.PanicsWithValue(t, "JSON-RPC files requested before Plan.Link", func() {
		plans[0].Service(roots[0].API.JSONRPC.Services[0])
	})

	require.NoError(t, generation.Freeze())
	for _, servicePlan := range services {
		require.NoError(t, servicePlan.Link())
	}
	for _, httpPlan := range jsonPlans {
		require.NoError(t, httpPlan.Link())
	}
	for _, plan := range plans {
		require.NoError(t, plan.Link())
	}

	data, ok := plans[0].Service(roots[0].API.JSONRPC.Services[0])
	require.True(t, ok)
	require.Equal(t, "First", data.Service.Name)
	require.NotEmpty(t, data.ClientStructDeclaration.Name())
	require.NotEmpty(t, data.ServerStructDeclaration.Name())

	foreign := expr.RunDSL(t, jsonRPCPlanningRootDSL("First", "/first"))
	data, ok = plans[0].Service(foreign.API.JSONRPC.Services[0])
	require.False(t, ok)
	require.Empty(t, data)
}

// TestPlanServiceReturnsDetachedSnapshot verifies that changing nested values
// returned to one plugin cannot change a later read or the files already
// prepared for generation.
func TestPlanServiceReturnsDetachedSnapshot(t *testing.T) {
	_, plan := linkedJSONRPCPlan(t, viewedJSONRPCPlanDSL)
	service := plan.root.API.JSONRPC.Services[0]
	before := renderJSONRPCFiles(t, plan.ClientFiles())

	first, ok := plan.Service(service)
	require.True(t, ok)
	require.NotEmpty(t, first.Endpoints)
	require.NotNil(t, first.Endpoints[0].Result)
	first.Endpoints[0].Result.Ref = "changed.Result"

	second, ok := plan.Service(service)
	require.True(t, ok)
	require.NotEqual(t, "changed.Result", second.Endpoints[0].Result.Ref)
	require.Equal(t, before, renderJSONRPCFiles(t, plan.ClientFiles()))
}

// renderJSONRPCFiles writes file sections without changing the plan.
func renderJSONRPCFiles(t *testing.T, files []*codegen.File) string {
	t.Helper()
	var source strings.Builder
	for _, file := range files {
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(&source))
		}
	}
	return source.String()
}
