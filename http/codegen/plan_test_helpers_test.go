// This file prepares HTTP plans for tests through the same name assignment and
// file-building steps used by the generator.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// linkedHTTPPlanForRoot builds the HTTP files for root after every generated
// package has received its final names.
func linkedHTTPPlanForRoot(t *testing.T, root *expr.RootExpr) *Plan {
	t.Helper()
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())
	return plans[0]
}

// linkedHTTPExamplePlanForRoot builds an HTTP plan whose copied server data
// belongs to the same service plan.
func linkedHTTPExamplePlanForRoot(t *testing.T, root *expr.RootExpr) *ExamplePlan {
	t.Helper()
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	examplePlan, err := example.NewPlan(generation, servicePlan)
	require.NoError(t, err)
	examples, err := NewExamplePlan(plans[0], examplePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())
	return examples
}
