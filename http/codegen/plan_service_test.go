// This file checks that plugins can read only the finalized HTTP service data
// that belongs to the exact service expression used to build a retained plan.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestPlanServiceRequiresLink(t *testing.T) {
	root := serviceLookupRoot(t)
	plan, _, _ := plannedHTTPPlan(t, root, false)

	require.PanicsWithValue(t, "HTTP render model requested before plan linking", func() {
		plan.Service(root.API.HTTP.Services[0])
	})
}

func TestPlanServiceUsesExactExpression(t *testing.T) {
	root := serviceLookupRoot(t)
	plan, generation, servicePlan := plannedHTTPPlan(t, root, false)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plan.Link())

	alpha, ok := plan.Service(root.API.HTTP.Services[0])
	require.True(t, ok)
	require.Equal(t, "Alpha", alpha.Service.Name)

	beta, ok := plan.Service(root.API.HTTP.Services[1])
	require.True(t, ok)
	require.Equal(t, "Beta", beta.Service.Name)
	require.NotSame(t, alpha, beta)

	foreign := serviceLookupRoot(t)
	_, ok = plan.Service(foreign.API.HTTP.Services[0])
	require.False(t, ok)
}

// serviceLookupRoot creates two services so the test can distinguish exact
// service identity from a lookup by a repeated name or position.
func serviceLookupRoot(t *testing.T) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		dsl.Service("Alpha", func() {
			dsl.Method("Read", func() {
				dsl.HTTP(func() { dsl.GET("/alpha") })
			})
		})
		dsl.Service("Beta", func() {
			dsl.Method("Read", func() {
				dsl.HTTP(func() { dsl.GET("/beta") })
			})
		})
	})
}
