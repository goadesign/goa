// These tests compare saved paths with the complete paths observed during a
// traversal and verify that deep mutations still report their exact location.
package generator

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestPreparedDesignSavedPathsMatchTraversal(t *testing.T) {
	type node struct {
		Next  *node
		Value string
	}
	rootNode := &node{}
	leaf := rootNode
	for range 64 {
		leaf.Next = &node{}
		leaf = leaf.Next
	}
	root := &expr.RootExpr{
		Types: []expr.UserType{&expr.UserTypeExpr{
			TypeName: "Record",
			AttributeExpr: &expr.AttributeExpr{
				Type:         expr.String,
				DefaultValue: map[string]any{"deep": rootNode, "quoted\"key": []string{"value"}},
			},
		}},
		Conversions: []*expr.TypeMap{{External: &node{}}},
	}
	roots := []eval.Root{root}
	snapshot, err := snapshotPreparedDesign(roots)
	require.NoError(t, err)
	var paths []string
	walker := &designSnapshotter{
		visited: make(map[designVisit]struct{}),
		record: func(path []byte, _ reflect.Type, _ string) {
			paths = append(paths, string(path))
		},
	}
	require.NoError(t, walker.appendValue([]byte("roots[0]"), reflect.ValueOf(root)))
	require.Len(t, snapshot.states, len(paths))
	for index, state := range snapshot.states {
		require.Equal(t, paths[index], state.path.string())
	}
	leaf.Value = "changed"
	changed, err := snapshot.changedPath(roots)
	require.NoError(t, err)
	require.Equal(t, `roots[0].Types[0].AttributeExpr.DefaultValue["deep"]`+strings.Repeat(".Next", 64)+".Value", changed)
	leaf.Value = ""
	changed, err = snapshot.changedPath(roots)
	require.NoError(t, err)
	require.Empty(t, changed)
}
