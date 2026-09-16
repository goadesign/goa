// This file checks that helper sharing uses the generated value inside every
// collection wrapper, including wrappers nested below other collections.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestTransformHelperRegistryFollowsNestedWrappers(t *testing.T) {
	for _, wrapTarget := range []bool{false, true} {
		t.Run(fmt.Sprintf("wrap-target=%t", wrapTarget), func(t *testing.T) {
			rootPlan, source, target := rootWrappedTransformPlan(t, wrapTarget)
			registry := NewTransformHelperRegistry()
			for depth := range 3 {
				plan, err := rootPlan.program.Plan(source, target, "")
				require.NoError(t, err)
				require.NoError(t, registry.Collect(
					plan,
					transformTestLayout(t, source, GoLayoutPolicy{UseDefault: true}),
					transformTestLayout(t, target, GoLayoutPolicy{UseDefault: true}),
					transformTestOrderFactory(fmt.Sprintf("depth-%d", depth)).order,
				))
				source = nestedTransformCollection(source)
				target = nestedTransformCollection(target)
			}
			groups, err := registry.Finalize()
			require.NoError(t, err)
			require.Len(t, groups, 1, "wrapping a value must not change its conversion helper")
		})
	}
}

// nestedTransformCollection puts a value inside both an array and a map, so
// the next helper lookup must cross both collections before opening its wrapper.
func nestedTransformCollection(value *expr.AttributeExpr) *expr.AttributeExpr {
	return &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: &expr.Map{
		KeyType:  &expr.AttributeExpr{Type: expr.String},
		ElemType: value,
	}}}}
}
