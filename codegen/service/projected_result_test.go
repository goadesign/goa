// This file checks that HTTP generation can copy the result fields selected by
// a view before names are assigned to the service package.
package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestPlanProjectedResultBeforeLink(t *testing.T) {
	var viewed, plain *expr.MethodExpr
	root := codegen.RunDSL(t, func() {
		result := dsl.ResultType("application/vnd.projected-result", func() {
			dsl.TypeName("ProjectedResult")
			dsl.Attribute("name", dsl.String)
			dsl.View("default", func() { dsl.Attribute("name") })
			dsl.View("summary", func() { dsl.Attribute("name") })
		})
		dsl.Service("Values", func() {
			viewed = dsl.Method("Viewed", func() { dsl.Result(result) })
			plain = dsl.Method("Plain", func() { dsl.Result(dsl.String) })
		})
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)

	first, err := plan.ProjectedResult(viewed)
	require.NoError(t, err)
	first.Description = "changed by caller"
	second, err := plan.ProjectedResult(viewed)
	require.NoError(t, err)
	require.NotEqual(t, first.Description, second.Description)
	require.NotSame(t, first, second)
	declaration, err := plan.ProjectedResultDeclaration(viewed)
	require.NoError(t, err)
	require.NotNil(t, declaration)

	_, err = plan.ProjectedResult(plain)
	require.EqualError(t, err, `service method "Plain" does not have a viewed result`)
	_, err = plan.ProjectedResultDeclaration(plain)
	require.EqualError(t, err, `service method "Plain" does not have a viewed result`)
	foreign := &expr.MethodExpr{Name: "Foreign"}
	_, err = plan.ProjectedResult(foreign)
	require.EqualError(t, err, `service method "Foreign" is not part of this plan`)
	_, err = plan.ProjectedResultDeclaration(foreign)
	require.EqualError(t, err, `service method "Foreign" is not part of this plan`)

	require.NoError(t, generation.Freeze())
	require.Equal(t, "ProjectedResultView", declaration.Name())
	require.NoError(t, plan.Link())
	linked := plan.Services().Get("Values").Method("Viewed").ViewedResult
	require.NotNil(t, linked)
	require.NotEqual(t, "changed by caller", expr.AsObject(linked.Type).Attribute("projected").Description)
}

func TestPlanHTTPMethodNamesBeforeLink(t *testing.T) {
	var watch *expr.MethodExpr
	root := codegen.RunDSL(t, func() {
		dsl.Service("Values", func() {
			watch = dsl.Method("Watch", func() {
				dsl.StreamingResult(dsl.String)
			})
		})
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)

	names, err := plan.HTTPMethodNames(watch)
	require.NoError(t, err)
	require.Equal(t, "Watch", names.Method)
	require.Equal(t, "WatchServerStream", names.ServerStream)
	require.Equal(t, "WatchClientStream", names.ClientStream)

	_, err = plan.HTTPMethodNames(&expr.MethodExpr{Name: "Foreign"})
	require.EqualError(t, err, `service method "Foreign" is not part of this plan`)
}
