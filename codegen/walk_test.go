// This file verifies that attribute traversal distinguishes unrelated design
// declarations while terminating when a declaration refers to itself.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestWalkDistinguishesEqualUIDOrigins(t *testing.T) {
	firstLeaf := &expr.AttributeExpr{Type: expr.String}
	secondLeaf := &expr.AttributeExpr{Type: expr.Int}
	first := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "first", Attribute: firstLeaf},
		}},
		TypeName: "First",
		UID:      "shared",
	}
	second := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "second", Attribute: secondLeaf},
		}},
		TypeName: "Second",
		UID:      "shared",
	}
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
	}}

	visited := make(map[*expr.AttributeExpr]bool)
	require.NoError(t, Walk(root, func(att *expr.AttributeExpr) error {
		visited[att] = true
		return nil
	}))
	require.True(t, visited[firstLeaf])
	require.True(t, visited[secondLeaf])
}

func TestWalkPreservesDynamicResultTypeOrigin(t *testing.T) {
	baseLeaf := &expr.AttributeExpr{Type: expr.String}
	base := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "base", Attribute: baseLeaf},
		}},
		TypeName: "Base",
	}
	resultLeaf := &expr.AttributeExpr{Type: expr.Int}
	embedded := base.Dup(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "result", Attribute: resultLeaf},
	}}).(*expr.UserTypeExpr)
	result := &expr.ResultTypeExpr{
		UserTypeExpr: embedded,
		Identifier:   "application/vnd.result",
	}
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "base", Attribute: &expr.AttributeExpr{Type: base}},
		{Name: "result", Attribute: &expr.AttributeExpr{Type: result}},
	}}

	visited := make(map[*expr.AttributeExpr]bool)
	require.NoError(t, Walk(root, func(att *expr.AttributeExpr) error {
		visited[att] = true
		return nil
	}))
	require.True(t, visited[baseLeaf])
	require.True(t, visited[resultLeaf])
}

func TestWalkTypeTerminatesRecursiveCopy(t *testing.T) {
	recursive := &expr.UserTypeExpr{TypeName: "Recursive", UID: "recursive"}
	object := &expr.Object{}
	recursive.AttributeExpr = &expr.AttributeExpr{Type: object}
	self := &expr.AttributeExpr{Type: recursive}
	object.Set("self", self)
	copy := expr.Dup(recursive).(expr.UserType)

	visits := 0
	require.NoError(t, WalkType(copy, func(*expr.AttributeExpr) error {
		visits++
		return nil
	}))
	require.Equal(t, 2, visits)
}
