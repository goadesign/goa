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

// TestDupGeneratedResultTypes verifies that ordinary copies do not change the
// evaluated design and DSL copies register the result types whose DSL must run.
func TestDupGeneratedResultTypes(t *testing.T) {
	ResetDSL(t)

	resultType := NewResultTypeExpr("Item", "application/vnd.item", func() {})
	GeneratedResultTypes.Append(resultType)
	attribute := &AttributeExpr{Type: &Array{ElemType: &AttributeExpr{Type: resultType}}}

	duplicate := DupAtt(attribute)
	require.Len(t, *GeneratedResultTypes, 1)
	require.NotSame(t, resultType, duplicate.Type.(*Array).ElemType.Type)

	dslDuplicate := DupAttForDSL(attribute)
	require.Len(t, *GeneratedResultTypes, 2)
	require.Same(t, dslDuplicate.Type.(*Array).ElemType.Type, (*GeneratedResultTypes)[1])
}

func TestIsErrorResultRecognizesCopies(t *testing.T) {
	duplicate := DupAtt(&AttributeExpr{Type: ErrorResult}).Type

	require.True(t, IsErrorResult(ErrorResult))
	require.True(t, IsErrorResult(duplicate))
	require.False(t, IsErrorResult(&UserTypeExpr{
		AttributeExpr: &AttributeExpr{Type: String},
		TypeName:      "error",
	}))
	require.False(t, IsErrorResult(String))
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

// TestDupAttributeKeepsAuthoredAttribute verifies that every transport copy
// can find the exact attribute written in the design, including after more
// than one copy.
func TestDupAttributeKeepsAuthoredAttribute(t *testing.T) {
	originalChild := &AttributeExpr{Type: String}
	original := &AttributeExpr{Type: &Object{
		{Name: "child", Attribute: originalChild},
	}}

	first := DupAtt(original)
	second := DupAtt(first)

	require.Same(t, original, first.AuthoredAttribute())
	require.Same(t, original, second.AuthoredAttribute())
	require.Same(t, originalChild, first.Type.(*Object).Attribute("child").AuthoredAttribute())
	require.Same(t, originalChild, second.Type.(*Object).Attribute("child").AuthoredAttribute())
}

func TestDupSchemeKeepsAuthoredScheme(t *testing.T) {
	authored := &SchemeExpr{SchemeName: "key"}
	first := DupScheme(authored)
	second := DupScheme(first)

	require.Same(t, authored, first.AuthoredScheme())
	require.Same(t, authored, second.AuthoredScheme())
}
