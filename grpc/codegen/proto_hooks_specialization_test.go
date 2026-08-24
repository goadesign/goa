// This file verifies that gRPC collection conversions use the recorded nesting
// level to choose loop variables instead of examining generated Go expression
// text.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// protobufOneofSnapshotAttributor records the exact union branch used to
	// resolve a protobuf wrapper name.
	protobufOneofSnapshotAttributor struct {
		codegen.Attributor
		branches *[]*expr.AttributeExpr
	}
)

func TestRenderArrayTransformUsesTraversalDepthForLoopVariable(t *testing.T) {
	source := &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Int}}
	target := &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Int}}
	sourceContext := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope())
	targetContext := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope())
	attributes := &codegen.TransformAttrs{
		SourceCtx: sourceContext,
		TargetCtx: targetContext,
		Hooks:     protoHooks(true),
	}

	generated, err := renderArrayTransform(source, target, "source", "target[key]", false, true, attributes)
	require.NoError(t, err)
	require.Contains(t, generated, "for i, val := range source")
	require.Contains(t, generated, "target[key][i] =")
}

func TestRenderArrayTransformPreservesPrimitiveAliasForElementConversion(t *testing.T) {
	alias := &expr.UserTypeExpr{
		TypeName:      "Alias",
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
	}
	source := &expr.Array{ElemType: &expr.AttributeExpr{Type: alias}}
	target := &expr.Array{ElemType: &expr.AttributeExpr{Type: alias}}
	sourceContext := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope())
	targetContext := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope())
	var convertedTarget *expr.AttributeExpr
	attributes := &codegen.TransformAttrs{
		SourceCtx: sourceContext,
		TargetCtx: targetContext,
		Hooks: &codegen.TransformHooks{
			ConvertPrimitive: func(_ *expr.AttributeExpr, target *expr.AttributeExpr, _ string, _, _ bool, _ *codegen.TransformAttrs) (string, bool) {
				convertedTarget = target
				return "string(val)", true
			},
		},
	}

	_, err := renderArrayTransform(source, target, "source", "target", true, true, attributes)
	require.NoError(t, err)
	require.Same(t, target.ElemType, convertedTarget)
}

func TestTransformPlanOneofLookupUsesOriginalBranch(t *testing.T) {
	sourceBranch := &expr.AttributeExpr{Type: expr.String}
	targetBranch := &expr.AttributeExpr{Type: expr.String}
	source := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "SourceChoice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "text", Attribute: sourceBranch},
		},
	}}
	target := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "TargetChoice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "text", Attribute: targetBranch},
		},
	}}
	plan, err := codegen.NewTransformPlan(source, target, "", protoHooks(true))
	require.NoError(t, err)
	require.Empty(t, plan.Helpers())

	sourceContext := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope())
	targetContext := codegen.NewAttributeContext(false, false, true, "", codegen.NewNameScope())
	var branches []*expr.AttributeExpr
	targetContext.Scope = &protobufOneofSnapshotAttributor{
		Attributor: targetContext.Scope,
		branches:   &branches,
	}
	require.NoError(t, plan.BindContexts(sourceContext, targetContext))
	generated, definitions, err := plan.Render("source", "target", true)
	require.NoError(t, err)
	require.Empty(t, definitions)
	require.Contains(t, generated, "TargetChoice_Text")
	require.Len(t, branches, 1)
	require.Same(t, targetBranch, branches[0])
}

// Enter preserves protobuf wrapper lookup while entering a copied union.
func (a *protobufOneofSnapshotAttributor) Enter(attribute *expr.AttributeExpr) codegen.Attributor {
	return &protobufOneofSnapshotAttributor{
		Attributor: a.Attributor.Enter(attribute),
		branches:   a.branches,
	}
}

// IsSumType reports that this test resolver uses protobuf oneof wrappers.
func (*protobufOneofSnapshotAttributor) IsSumType() bool {
	return false
}

// OneofWrapper records the branch and returns its planned protobuf type.
func (a *protobufOneofSnapshotAttributor) OneofWrapper(attribute *expr.AttributeExpr) string {
	*a.branches = append(*a.branches, attribute)
	return "TargetChoice_Text"
}
