// These tests compare direct audits with the original full-snapshot comparison.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestPreparedDesignDirectComparisonMatchesSnapshots(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*expr.RootExpr)
	}{
		{"unchanged", func(*expr.RootExpr) {
		}},
		{"scalar", func(root *expr.RootExpr) {
			root.Types[0].Attribute().Description = "changed"
		}},
		{"map entry", func(root *expr.RootExpr) {
			root.Types[0].Attribute().Meta["first"][0] = "changed"
		}},
		{"map replacement", func(root *expr.RootExpr) {
			root.Types[0].Attribute().Meta = expr.MetaExpr{"second": {"new"}}
		}},
		{"nil collection", func(root *expr.RootExpr) {
			root.Types = nil
		}},
		{"empty collection", func(root *expr.RootExpr) {
			root.Types = []expr.UserType{}
		}},
		{"alias replacement", func(root *expr.RootExpr) {
			copied := *root.Types[0].(*expr.UserTypeExpr)
			root.Types[1] = &copied
		}},
		{"unsupported after mutation", func(root *expr.RootExpr) {
			root.Types[0].(*expr.UserTypeExpr).TypeName = "Changed"
			root.Types[0].Attribute().DefaultValue = make(chan int)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := &expr.UserTypeExpr{TypeName: "Node"}
			node.AttributeExpr = &expr.AttributeExpr{
				Type: &expr.Object{{Name: "next", Attribute: &expr.AttributeExpr{Type: node}}},
				Meta: expr.MetaExpr{"first": {"original"}},
			}
			root := &expr.RootExpr{Types: []expr.UserType{node, node}}
			roots := []eval.Root{root}
			before, err := snapshotPreparedDesign(roots)
			require.NoError(t, err)
			test.mutate(root)
			after, wantErr := snapshotPreparedDesign(roots)
			path, gotErr := before.changedPath(roots)
			if wantErr != nil {
				require.EqualError(t, gotErr, wantErr.Error())
				require.Empty(t, path)
				return
			}
			require.NoError(t, gotErr)
			require.Equal(t, firstSnapshotDifference(before, after), path)
		})
	}
}

func TestPreparedDesignDirectComparisonRootCount(t *testing.T) {
	root := &expr.RootExpr{}
	for _, test := range []struct {
		name          string
		before, after []eval.Root
	}{
		{"added", nil, []eval.Root{root}},
		{"removed", []eval.Root{root}, nil},
		{"empty", nil, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			before, err := snapshotPreparedDesign(test.before)
			require.NoError(t, err)
			after, err := snapshotPreparedDesign(test.after)
			require.NoError(t, err)
			path, err := before.changedPath(test.after)
			require.NoError(t, err)
			require.Equal(t, firstSnapshotDifference(before, after), path)
		})
	}
}

// firstSnapshotDifference preserves the previous comparison as a test oracle.
func firstSnapshotDifference(before, after *designSnapshot) string {
	common := min(len(before.states), len(after.states))
	for index := range common {
		if before.states[index] != after.states[index] {
			return after.states[index].path
		}
	}
	if len(before.states) > common {
		return before.states[common].path
	}
	if len(after.states) > common {
		return after.states[common].path
	}
	return ""
}
