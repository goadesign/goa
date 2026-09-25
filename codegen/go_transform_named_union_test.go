package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestTransformUnionCollectionsKeepHooks(t *testing.T) {
	for _, mapped := range []bool{false, true} {
		t.Run(fmt.Sprintf("map=%t", mapped), func(t *testing.T) {
			union := &expr.AttributeExpr{Type: &expr.Union{
				TypeName: "Choice",
				Values: []*expr.NamedAttributeExpr{
					{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
				},
			}}
			attribute := &expr.AttributeExpr{Type: &expr.Array{ElemType: union}}
			if mapped {
				attribute.Type = &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: union}
			}
			var retained *expr.AttributeExpr
			var unwraps, renders int
			hooks := &TransformHooks{
				UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *WrapDirective) {
					unwraps++
					if expr.IsUnion(source.Type) {
						retained = source
						require.NotSame(t, union, source)
					}
					return source, target, nil
				},
				TransformUnion: func(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar bool, sourceParent, targetParent *expr.AttributeExpr, attrs *TransformAttrs) (string, error) {
					renders++
					require.Same(t, retained, source)
					require.Nil(t, sourceParent)
					require.Nil(t, targetParent)
					require.False(t, newVar)
					require.NotNil(t, attrs.SourceCtx)
					require.NotNil(t, attrs.TargetCtx)
					return targetVar + " = " + sourceVar + "\n", nil
				},
			}
			plan, err := NewTransformPlan(attribute, attribute, "", hooks)
			require.NoError(t, err)
			plannedUnwraps := unwraps
			context := NewAttributeContext(false, false, true, "", NewNameScope())
			require.NoError(t, plan.BindContexts(context, context))
			code, helpers, err := plan.Render("source", "target", true)
			require.NoError(t, err)
			require.Empty(t, helpers)
			require.Contains(t, code, "= val")
			require.Equal(t, 1, renders)
			require.Equal(t, plannedUnwraps, unwraps, "render must use the planned unwrap choice")
		})
	}
}
