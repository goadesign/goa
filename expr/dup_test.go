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

// TestDupAttFinalizeResolvesTransitiveBasesAndReferences verifies that DupAtt
// deep-copies transitive Bases and References so finalizing a detached copy
// resolves inherited fields without mutating the original unfinalized types.
func TestDupAttFinalizeResolvesTransitiveBasesAndReferences(t *testing.T) {
	refType := &UserTypeExpr{
		TypeName: "ReferenceDoc",
		AttributeExpr: &AttributeExpr{
			Type: &Object{
				{Name: "doc_meta", Attribute: &AttributeExpr{
					Type:        String,
					Description: "documentation metadata",
				}},
			},
		},
	}

	minLen := 2
	ancestor := &UserTypeExpr{
		TypeName: "Ancestor",
		AttributeExpr: &AttributeExpr{
			Type: &Object{
				{Name: "query", Attribute: &AttributeExpr{
					Type:       String,
					Validation: &ValidationExpr{MinLength: &minLen},
				}},
				{Name: "doc_meta", Attribute: &AttributeExpr{
					Type: String,
				}},
			},
			References: []DataType{refType},
			Validation: &ValidationExpr{Required: []string{"query"}},
		},
	}

	base := &UserTypeExpr{
		TypeName: "Base",
		AttributeExpr: &AttributeExpr{
			Type:  &Object{},
			Bases: []DataType{ancestor},
		},
	}

	derived := &UserTypeExpr{
		TypeName: "Derived",
		AttributeExpr: &AttributeExpr{
			Type:  &Object{},
			Bases: []DataType{base},
		},
	}

	node := &UserTypeExpr{
		TypeName: "Node",
		AttributeExpr: &AttributeExpr{
			Type: &Object{
				{Name: "self", Attribute: &AttributeExpr{Type: String}},
				{Name: "nested", Attribute: &AttributeExpr{
					Type:  &Object{},
					Bases: []DataType{derived},
				}},
			},
		},
	}

	// Verify original types are unfinalized with bases/references intact
	require.False(t, ancestor.Attribute().finalized)
	require.False(t, ancestor.Attribute().inheritanceResolved)
	require.False(t, base.Attribute().finalized)
	require.False(t, base.Attribute().inheritanceResolved)
	require.False(t, derived.Attribute().finalized)
	require.False(t, derived.Attribute().inheritanceResolved)
	require.False(t, node.Attribute().finalized)
	require.False(t, node.Attribute().inheritanceResolved)
	require.Len(t, ancestor.Attribute().References, 1)
	require.Len(t, base.Attribute().Bases, 1)
	require.Len(t, derived.Attribute().Bases, 1)
	require.Len(t, AsObject(node.Type).Attribute("nested").Bases, 1)

	// Dup node attribute and finalize the detached duplicate
	dupNodeAtt := DupAtt(node.Attribute())
	dupNodeAtt.Finalize()

	// Verify duplicate resolved inheritance and finalization
	require.True(t, dupNodeAtt.finalized)
	dupNested := AsObject(dupNodeAtt.Type).Attribute("nested")
	require.NotNil(t, dupNested)
	require.True(t, dupNested.finalized)
	require.True(t, dupNested.inheritanceResolved)
	require.Nil(t, dupNested.Bases, "bases should be cleared after merge")

	dupNestedObj := AsObject(dupNested.Type)
	require.NotNil(t, dupNestedObj)
	queryAtt := dupNestedObj.Attribute("query")
	require.NotNil(t, queryAtt, "detached copy must have query attribute merged from Ancestor")
	assert.Equal(t, String, queryAtt.Type)
	require.NotNil(t, queryAtt.Validation)
	require.NotNil(t, queryAtt.Validation.MinLength)
	assert.Equal(t, 2, *queryAtt.Validation.MinLength)
	assert.True(t, dupNested.IsRequired("query"))

	docMetaAtt := dupNestedObj.Attribute("doc_meta")
	require.NotNil(t, docMetaAtt, "detached copy must have doc_meta inherited from ReferenceDoc")
	assert.Equal(t, "documentation metadata", docMetaAtt.Description)

	// Verify original unfinalized types were NOT mutated
	assert.False(t, ancestor.Attribute().finalized)
	assert.False(t, ancestor.Attribute().inheritanceResolved)
	assert.False(t, base.Attribute().finalized)
	assert.False(t, base.Attribute().inheritanceResolved)
	assert.False(t, derived.Attribute().finalized)
	assert.False(t, derived.Attribute().inheritanceResolved)
	assert.False(t, node.Attribute().finalized)
	assert.False(t, node.Attribute().inheritanceResolved)
	assert.Len(t, ancestor.Attribute().References, 1)
	assert.Len(t, base.Attribute().Bases, 1)
	assert.Len(t, derived.Attribute().Bases, 1)
	assert.Len(t, AsObject(node.Type).Attribute("nested").Bases, 1)
}
