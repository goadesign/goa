// These tests compare direct audits with the original full-snapshot comparison.
package generator

import (
	"runtime"
	"testing"
	"weak"

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
		{"spare slice capacity", func(root *expr.RootExpr) {
			root.Types[:cap(root.Types)][2] = &expr.UserTypeExpr{
				TypeName: "Hidden", AttributeExpr: &expr.AttributeExpr{Type: expr.String},
			}
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
			types := make([]expr.UserType, 2, 3)
			types[0], types[1] = node, node
			root := &expr.RootExpr{Types: types}
			roots := []eval.Root{root}
			before, err := snapshotPreparedDesign(roots)
			require.NoError(t, err)
			test.mutate(root)
			runtime.GC()
			path, gotErr := before.changedPath(roots)
			after, wantErr := snapshotPreparedDesign(roots)
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

func TestPreparedDesignSnapshotRetainsRemovedObjects(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.String, Description: "original"}
	removed := weak.Make(attribute)
	root := &expr.RootExpr{Types: []expr.UserType{&expr.UserTypeExpr{
		TypeName: "Record", AttributeExpr: attribute,
	}}}
	roots := []eval.Root{root}
	snapshot, err := snapshotPreparedDesign(roots)
	require.NoError(t, err)
	root.Types = nil

	for range 3 {
		runtime.GC()
		// The saved snapshot must retain removed objects so their addresses cannot
		// be reused by later design values and hide a replacement.
		require.NotNil(t, removed.Value())
		changed, err := snapshot.changedPath(roots)
		require.NoError(t, err)
		require.Equal(t, "roots[0].Types", changed)
	}
	runtime.KeepAlive(snapshot)
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
