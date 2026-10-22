// This file verifies that collecting shared HTTP definitions once preserves
// separate copied shapes, union declarations, and per-call transport policies.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestWireCollectionSharedDefinition(t *testing.T) {
	choice := &expr.AttributeExpr{Type: &expr.Union{TypeName: "Choice", Values: []*expr.NamedAttributeExpr{
		{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}}}
	leaf := &expr.UserTypeExpr{TypeName: "Leaf", UID: "leaf", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
		{Name: "choice", Attribute: choice},
	}}}
	body := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: leaf}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: leaf}},
	}}
	catalog := newWireTypeCatalog()
	catalog.collect(body, wireResponseBody, wireTypePolicy{})
	require.Len(t, catalog.records, 1)
	require.Len(t, catalog.unionOccurrences, 1)
	require.Equal(t, []string{"LeafResponseBody"}, catalog.records[0].releasedNames)

	// A separate request collection must still add its spelling and validator,
	// even when the same definition was already collected for a response.
	catalog.collect(body, wireRequestBody, wireTypePolicy{validate: true})
	require.Len(t, catalog.records, 1)
	require.True(t, catalog.records[0].needsValidator)
	require.Equal(t, []string{"LeafRequestBody", "LeafResponseBody"}, catalog.records[0].releasedNames)
	require.Len(t, catalog.unionOccurrences, 2)
	require.Equal(t, wireRequestBody, catalog.unionOccurrences[1].use.role)
}

func TestWireCollectionSeparateCopiesWithSharedOrigin(t *testing.T) {
	for _, changedShape := range []bool{false, true} {
		name := "equal shapes with distinct unions"
		if changedShape {
			name = "different descendant shapes"
		}
		t.Run(name, func(t *testing.T) {
			original := wireCatalogType("Parent", "parent", "value", false)
			first := expr.DupAtt(&expr.AttributeExpr{Type: original}).Type.(expr.UserType)
			second := expr.DupAtt(&expr.AttributeExpr{Type: original}).Type.(expr.UserType)
			require.Same(t, first.Origin(), second.Origin())
			sharedChild := wireCatalogType("Child", "child", "value", false)
			for index, copied := range []expr.UserType{first, second} {
				child := sharedChild
				if changedShape && index == 1 {
					child = wireCatalogType("OtherChild", "other-child", "value", false)
				}
				copied.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
					{Name: "choice", Attribute: &expr.AttributeExpr{Type: &expr.Union{TypeName: "Choice", Values: []*expr.NamedAttributeExpr{
						{Name: "child", Attribute: &expr.AttributeExpr{Type: child}},
					}}}},
				}})
			}
			body := &expr.AttributeExpr{Type: &expr.Object{
				{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
				{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
			}}
			catalog := newWireTypeCatalog()
			catalog.collect(body, wireResponseBody, wireTypePolicy{})
			require.Len(t, catalog.unionOccurrences, 2)
			firstUnion := catalog.unionOccurrences[0].attribute.AuthoredAttribute()
			secondUnion := catalog.unionOccurrences[1].attribute.AuthoredAttribute()
			require.NotSame(t, firstUnion, secondUnion)
			firstRecord := catalog.find(newWireTypeIdentity(&expr.AttributeExpr{Type: first}, wireAttribute, wireTypePolicy{}, "Parent"))
			secondRecord := catalog.find(newWireTypeIdentity(&expr.AttributeExpr{Type: second}, wireAttribute, wireTypePolicy{}, "Parent"))
			require.NotNil(t, firstRecord)
			require.NotNil(t, secondRecord)
			if !changedShape {
				require.Same(t, firstRecord, secondRecord)
			}
			if changedShape {
				names := make(map[string]bool)
				for _, record := range catalog.records {
					names[record.identity.preferred] = true
				}
				require.True(t, names["Child"])
				require.True(t, names["OtherChild"])
			}
		})
	}
}

func TestWireCollectionRevisitsDefinitionAfterAncestorStoppedItsChildren(t *testing.T) {
	ancestor := wireCatalogType("Ancestor", "ancestor", "value", false)
	copy := expr.DupAtt(&expr.AttributeExpr{Type: ancestor}).Type.(expr.UserType)
	child := wireCatalogType("Child", "child", "value", false)
	copy.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "child", Attribute: &expr.AttributeExpr{Type: child}},
	}})
	middle := &expr.UserTypeExpr{TypeName: "Middle", UID: "middle", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
		{Name: "ancestor", Attribute: &expr.AttributeExpr{Type: copy}},
	}}}
	ancestor.SetAttribute(&expr.AttributeExpr{Type: &expr.Object{
		{Name: "middle", Attribute: &expr.AttributeExpr{Type: middle}},
	}})
	body := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: ancestor}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: middle}},
	}}
	catalog := newWireTypeCatalog()
	catalog.collect(body, wireResponseBody, wireTypePolicy{})
	found := false
	for _, record := range catalog.records {
		if record.identity.preferred == "Child" {
			found = true
		}
	}
	require.True(t, found, "the second occurrence must collect the child skipped beneath its ancestor")
}
