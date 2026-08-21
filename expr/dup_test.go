package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDupPreservesNonNullableArrayElements(t *testing.T) {
	t.Parallel()

	original := &Array{
		ElemType:         &AttributeExpr{Type: String},
		NonNullableElems: true,
	}

	duplicate, ok := Dup(original).(*Array)
	require.True(t, ok)
	assert.NotSame(t, original, duplicate)
	assert.NotSame(t, original.ElemType, duplicate.ElemType)
	assert.True(t, duplicate.NonNullableElems)
}

func TestDupKeepsRootTypeAndSameNameUnionAliasDistinct(t *testing.T) {
	rootType := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{Type: String},
		TypeName:      "ValueBool",
	}
	unionAlias := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{Type: Boolean},
		TypeName:      "ValueBool",
	}
	attribute := &AttributeExpr{Type: &Object{
		{Name: "root", Attribute: &AttributeExpr{Type: rootType}},
		{Name: "choice", Attribute: &AttributeExpr{Type: &Union{
			TypeName: "Value",
			Values: []*NamedAttributeExpr{{
				Name:      "bool",
				Attribute: &AttributeExpr{Type: unionAlias},
			}},
		}}},
	}}

	duplicate := DupAtt(attribute)
	object := duplicate.Type.(*Object)
	rootCopy := object.Attribute("root").Type.(UserType)
	union := object.Attribute("choice").Type.(*Union)
	aliasCopy := union.Values[0].Attribute.Type.(UserType)
	require.NotSame(t, rootCopy, aliasCopy)
	require.Same(t, rootType, rootCopy.Origin())
	require.Same(t, unionAlias, aliasCopy.Origin())
	require.Equal(t, String, rootCopy.Attribute().Type)
	require.Equal(t, Boolean, aliasCopy.Attribute().Type)
}
