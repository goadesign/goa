// These tests keep helper layout queries tied to the saved conversion paths.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestTransformPlanHelperLayoutsReturnRegistryNodes(t *testing.T) {
	plan := siblingTransformPlan(t)
	source := plan.sourceCopier.Original(plan.source)
	target := plan.targetCopier.Original(plan.target)
	policy := GoLayoutPolicy{UseDefault: true}
	sourceLayout := transformTestLayout(t, source, policy)
	targetLayout := transformTestLayout(t, target, policy)
	helpers := plan.Helpers()
	require.Len(t, helpers, 2)

	registry := NewTransformHelperRegistry()
	require.NoError(t, registry.Collect(plan, sourceLayout, targetLayout, transformTestOrderFactory("query").order))
	for index, helper := range helpers {
		selectedSource, selectedTarget, err := plan.HelperLayouts(helper.ID, sourceLayout, targetLayout)
		require.NoError(t, err)
		require.Same(t, registry.candidates[index].sourceLayout, selectedSource)
		require.Same(t, registry.candidates[index].targetLayout, selectedTarget)
		require.True(t, selectedSource.MatchesOccurrence(plan.sourceCopier.Original(plan.helpers[index].Source)))
		require.True(t, selectedTarget.MatchesOccurrence(plan.targetCopier.Original(plan.helpers[index].Target)))

		// Helpers exposes detached descriptions; changing one cannot redirect an ID.
		helper.Source.Type = expr.String
		helper.Target.Type = expr.Int
		againSource, againTarget, err := plan.HelperLayouts(helper.ID, sourceLayout, targetLayout)
		require.NoError(t, err)
		require.Same(t, selectedSource, againSource)
		require.Same(t, selectedTarget, againTarget)
		require.Nil(t, plan.Helpers()[index].Declaration)
	}
	require.NotSame(t, registry.candidates[0].sourceLayout, registry.candidates[1].sourceLayout)
}

func TestTransformPlanHelperLayoutsAcceptCompilerCopies(t *testing.T) {
	original := siblingTransformPlan(t)
	source := original.sourceCopier.Original(original.source)
	target := original.targetCopier.Original(original.target)
	// Transport normalization and layout planning can each copy the same input.
	plan, err := NewTransformPlan(expr.DupAtt(source), expr.DupAtt(target), "", nil)
	require.NoError(t, err)
	policy := GoLayoutPolicy{UseDefault: true}
	sourceLayout := transformTestLayout(t, expr.DupAtt(source), policy)
	targetLayout := transformTestLayout(t, expr.DupAtt(target), policy)
	helpers := plan.Helpers()
	require.Len(t, helpers, 2)
	for _, helper := range helpers {
		selectedSource, selectedTarget, err := plan.HelperLayouts(helper.ID, sourceLayout, targetLayout)
		require.NoError(t, err)
		require.Equal(t, "Recursive", selectedSource.fixedName)
		require.Equal(t, "Recursive", selectedTarget.fixedName)
	}
}

func TestTransformPlanHelperLayoutsRejectWrongInputs(t *testing.T) {
	plan := siblingTransformPlan(t)
	source := plan.sourceCopier.Original(plan.source)
	target := plan.targetCopier.Original(plan.target)
	policy := GoLayoutPolicy{UseDefault: true}
	sourceLayout := transformTestLayout(t, source, policy)
	targetLayout := transformTestLayout(t, target, policy)
	id := plan.Helpers()[0].ID

	foreign := siblingTransformPlan(t).Helpers()[0].ID
	_, _, err := plan.HelperLayouts(foreign, sourceLayout, targetLayout)
	require.EqualError(t, err, "transform helper does not belong to this plan")
	_, _, err = plan.HelperLayouts(TransformHelperID{}, sourceLayout, targetLayout)
	require.EqualError(t, err, "transform helper does not belong to this plan")
	_, _, err = plan.HelperLayouts(id, nil, targetLayout)
	require.EqualError(t, err, "transform helper layouts must not be nil")
	_, _, err = plan.HelperLayouts(id, sourceLayout, nil)
	require.EqualError(t, err, "transform helper layouts must not be nil")

	// Sharing a type is insufficient: the supplied layout must describe this root.
	wrongSource := transformTestLayout(t, &expr.AttributeExpr{Type: source.Type}, policy)
	_, _, err = plan.HelperLayouts(id, wrongSource, targetLayout)
	require.EqualError(t, err, "transform helper source layout describes another root")
	wrongTarget := transformTestLayout(t, &expr.AttributeExpr{Type: target.Type}, policy)
	_, _, err = plan.HelperLayouts(id, sourceLayout, wrongTarget)
	require.EqualError(t, err, "transform helper target layout describes another root")

	incomplete := *sourceLayout
	incomplete.value = nil
	_, _, err = plan.HelperLayouts(id, &incomplete, targetLayout)
	require.ErrorContains(t, err, "does not select a generated struct")

	changed := expr.DupAtt(target)
	fields := expr.AsObject(changed.Type)
	*fields = (*fields)[:1]
	changedLayout := transformTestLayout(t, changed, policy)
	_, _, err = plan.HelperLayouts(id, sourceLayout, changedLayout)
	require.ErrorContains(t, err, "object has 2 design fields and 1 generated fields")
}

// Existing wrapper directives select a field while the rendering scope keeps
// the complete enclosing value until it enters the actual helper parameter.
func TestTransformHelperContextsFollowExistingWrappers(t *testing.T) {
	for _, wrapTarget := range []bool{false, true} {
		t.Run(fmt.Sprintf("wrap-target=%t", wrapTarget), func(t *testing.T) {
			root, source, target := rootWrappedTransformPlan(t, wrapTarget)
			for range 3 {
				plan, err := root.program.Plan(source, target, "")
				require.NoError(t, err)
				policy := GoLayoutPolicy{UseDefault: true, SumType: true}
				sourceLayout := transformTestLayout(t, source, policy)
				targetLayout := transformTestLayout(t, target, policy)
				context := NewAttributeContext(false, false, true, "", NewNameScope())
				sourceContext, err := context.WithGoTypeLayout(sourceLayout.Link(sourceLayout.Owner(), nil))
				require.NoError(t, err)
				targetContext, err := context.WithGoTypeLayout(targetLayout.Link(targetLayout.Owner(), nil))
				require.NoError(t, err)
				require.NoError(t, plan.BindContexts(sourceContext, targetContext))
				require.Len(t, plan.helpers, 1)
				helper := plan.helpers[0]
				selectedSource, selectedTarget, err := plan.HelperLayouts(helper.ID, sourceLayout, targetLayout)
				require.NoError(t, err)
				enteredSource, err := plan.helperContext(plan.sourceCtx, helper, false)
				require.NoError(t, err)
				enteredTarget, err := plan.helperContext(plan.targetCtx, helper, true)
				require.NoError(t, err)
				require.Equal(t, selectedSource.Link(sourceLayout.Owner(), nil).Ref(), enteredSource.Scope.Ref(helper.Source, ""))
				require.Equal(t, selectedTarget.Link(targetLayout.Owner(), nil).Ref(), enteredTarget.Scope.Ref(helper.Target, ""))
				source = nestedTransformCollection(source)
				target = nestedTransformCollection(target)
			}
		})
	}
}
