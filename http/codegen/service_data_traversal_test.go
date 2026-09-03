// This file verifies that HTTP body shaping and traversal visit unrelated
// declarations even when their semantic identifiers or structures match.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestMakeHTTPTypeDistinguishesEqualUIDOrigins(t *testing.T) {
	first := locatedHTTPTraversalType("First", "shared", "first/types")
	second := locatedHTTPTraversalType("Second", "shared", "second/types")
	body := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
	}}

	wire := makeHTTPType(body)
	object := expr.AsObject(wire.Type)
	wireFirst := object.Attribute("first").Type.(expr.UserType)
	wireSecond := object.Attribute("second").Type.(expr.UserType)
	require.NotContains(t, wireFirst.Attribute().Meta, "struct:pkg:path")
	require.NotContains(t, wireSecond.Attribute().Meta, "struct:pkg:path")
}

func TestMakeHTTPTypeStopsRecursiveUnionBranches(t *testing.T) {
	node := &expr.UserTypeExpr{TypeName: "Node", UID: "node"}
	next := &expr.Union{
		TypeName: "Next",
		Values: []*expr.NamedAttributeExpr{
			{Name: "end", Attribute: &expr.AttributeExpr{Type: &expr.Object{}}},
			{Name: "node", Attribute: &expr.AttributeExpr{Type: node}},
		},
	}
	node.AttributeExpr = &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "next", Attribute: &expr.AttributeExpr{Type: next}},
		},
		Meta: expr.MetaExpr{"struct:pkg:path": {"service/types"}},
	}

	wire := makeHTTPType(&expr.AttributeExpr{Type: node})
	wireNode := wire.Type.(expr.UserType)
	wireNext := expr.AsObject(wireNode).Attribute("next").Type.(*expr.Union)
	wireRecursiveNode := wireNext.Values[1].Attribute.Type.(expr.UserType)
	require.Equal(t, "Node", wireRecursiveNode.Name())
	require.NotContains(t, wireNode.Attribute().Meta, "struct:pkg:path")
}

func TestCollectUserTypesDistinguishesEqualUIDOriginsAndStopsRecursion(t *testing.T) {
	first := &expr.UserTypeExpr{TypeName: "First", UID: "shared"}
	firstObject := &expr.Object{}
	first.AttributeExpr = &expr.AttributeExpr{Type: firstObject}
	firstObject.Set("self", &expr.AttributeExpr{Type: first})
	second := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}},
		TypeName:      "Second",
		UID:           "shared",
	}
	outer := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
			{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
		}},
		TypeName: "Outer",
	}

	var names []string
	collectUserTypes(outer, func(userType expr.UserType) {
		names = append(names, userType.Name())
	})
	require.Equal(t, []string{"Outer", "First", "Second"}, names)
}

func TestAddMarshalTagsDistinguishesEqualStructuralOrigins(t *testing.T) {
	first := marshalTagTraversalType()
	second := marshalTagTraversalType()
	outer := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
			{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
		}},
		TypeName: "Outer",
	}

	addMarshalTags(&expr.AttributeExpr{Type: outer})
	firstValue := expr.AsObject(expr.AsObject(first).Attribute("nested").Type).Attribute("value")
	secondValue := expr.AsObject(expr.AsObject(second).Attribute("nested").Type).Attribute("value")
	require.Equal(t, []string{"value"}, firstValue.Meta["struct:tag:json"])
	require.Equal(t, []string{"value"}, secondValue.Meta["struct:tag:json"])
}

// locatedHTTPTraversalType builds an authored declaration whose package path
// must be removed when Goa derives its transport type.
func locatedHTTPTraversalType(name, uid, packagePath string) *expr.UserTypeExpr {
	return &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{},
			Meta: expr.MetaExpr{"struct:pkg:path": {packagePath}},
		},
		TypeName: name,
		UID:      uid,
	}
}

// marshalTagTraversalType builds a declaration with an inline object whose
// fields must receive transport serialization tags.
func marshalTagTraversalType() *expr.UserTypeExpr {
	return &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "nested", Attribute: &expr.AttributeExpr{Type: &expr.Object{
				{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
			}}},
		}},
		TypeName: "Shared",
		UID:      "shared",
	}
}
