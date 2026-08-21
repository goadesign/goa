// This file verifies canonical comparison of reusable transport error value
// contracts independently of authored ordering and explicit default storage.
package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEquivalentErrorAttributesUseEffectiveUnionKeys(t *testing.T) {
	branches := []*NamedAttributeExpr{
		{Name: "text", Attribute: &AttributeExpr{Type: String}},
	}
	implicit := &AttributeExpr{Type: &Union{TypeName: "Value", Values: branches}}
	explicit := &AttributeExpr{Type: &Union{
		TypeName: "Value",
		TypeKey:  "type",
		ValueKey: "value",
		Values:   branches,
	}}

	require.True(t, equivalentErrorAttributes(implicit, explicit))
}

func TestEquivalentErrorAttributesIgnoreRequiredOrder(t *testing.T) {
	first := requiredObject("first", "second")
	second := requiredObject("second", "first")

	require.True(t, equivalentErrorAttributes(first, second))
}

func TestEquivalentErrorAttributesMaterializeBases(t *testing.T) {
	base := &UserTypeExpr{
		TypeName: "BaseError",
		AttributeExpr: &AttributeExpr{
			Type: &Object{
				{Name: "message", Attribute: &AttributeExpr{
					Type:         String,
					DefaultValue: "invalid",
					Meta:         MetaExpr{"struct:field:name": {"Message"}},
				}},
			},
			Validation: &ValidationExpr{Required: []string{"message"}},
		},
	}
	composed := &AttributeExpr{Type: &Object{}, Bases: []DataType{base}}
	explicit := &AttributeExpr{
		Type: &Object{
			{Name: "message", Attribute: &AttributeExpr{
				Type:         String,
				DefaultValue: "invalid",
				Meta:         MetaExpr{"struct:field:name": {"Message"}},
			}},
		},
		Validation: &ValidationExpr{Required: []string{"message"}},
	}

	require.True(t, equivalentErrorAttributes(composed, explicit))
	require.Len(t, composed.Bases, 1)
	require.False(t, composed.finalized)
	require.Empty(t, *AsObject(composed.Type))
	require.False(t, base.AttributeExpr.finalized)
}

func TestEquivalentErrorAttributesRejectDifferentEffectiveBases(t *testing.T) {
	stringBase := errorBase("Base", String)
	integerBase := errorBase("Base", Int)
	first := &AttributeExpr{Type: &Object{}, Bases: []DataType{stringBase}}
	second := &AttributeExpr{Type: &Object{}, Bases: []DataType{integerBase}}

	require.False(t, equivalentErrorAttributes(first, second))
}

func TestEquivalentErrorAttributesMaterializeReferences(t *testing.T) {
	reference := &UserTypeExpr{
		TypeName: "ReferenceError",
		AttributeExpr: &AttributeExpr{
			Type: &Object{
				{Name: "message", Attribute: &AttributeExpr{
					Type:         String,
					DefaultValue: "invalid",
				}},
			},
			Validation: &ValidationExpr{Required: []string{"message"}},
		},
	}
	referenced := &AttributeExpr{
		Type: &Object{
			{Name: "message", Attribute: &AttributeExpr{Type: String}},
		},
		References: []DataType{reference},
	}
	explicit := &AttributeExpr{
		Type: &Object{
			{Name: "message", Attribute: &AttributeExpr{
				Type:         String,
				DefaultValue: "invalid",
			}},
		},
		Validation: &ValidationExpr{Required: []string{"message"}},
	}

	require.True(t, equivalentErrorAttributes(referenced, explicit))
	require.Len(t, referenced.References, 1)
	require.Nil(t, referenced.Find("message").Validation)
	require.Nil(t, referenced.Find("message").DefaultValue)
}

func TestEquivalentErrorAttributesRejectReferenceContractDrift(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*AttributeExpr)
	}{
		{"validation", func(attribute *AttributeExpr) {
			minimum := 3
			attribute.Validation.MinLength = &minimum
		}},
		{"default", func(attribute *AttributeExpr) { attribute.DefaultValue = "different" }},
		{"metadata", func(attribute *AttributeExpr) {
			attribute.Meta["struct:field:name"] = []string{"Different"}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			referenced, explicit := referencedErrorContracts()
			test.mutate(explicit.Find("message"))

			require.False(t, equivalentErrorAttributes(referenced, explicit))
		})
	}
}

func TestEquivalentErrorAttributesCompareUnionBranchesPositionally(t *testing.T) {
	first := &AttributeExpr{Type: &Union{TypeName: "Value", Values: []*NamedAttributeExpr{
		{Name: "text", Attribute: &AttributeExpr{Type: String}},
		{Name: "count", Attribute: &AttributeExpr{Type: Int}},
	}}}
	second := &AttributeExpr{Type: &Union{TypeName: "Value", Values: []*NamedAttributeExpr{
		{Name: "count", Attribute: &AttributeExpr{Type: Int}},
		{Name: "text", Attribute: &AttributeExpr{Type: String}},
	}}}

	require.False(t, equivalentErrorAttributes(first, second))
}

func TestEquivalentErrorAttributesCopiesRecursiveDeclarations(t *testing.T) {
	first := recursiveErrorType("RecursiveError")
	second := recursiveErrorType("RecursiveError")

	require.True(t, equivalentErrorAttributes(
		&AttributeExpr{Type: first},
		&AttributeExpr{Type: second},
	))
	require.False(t, first.AttributeExpr.finalized)
	require.Same(t, first, first.Find("next").Type)
}

func TestEffectiveErrorAttributeSharesNoMutableContractValues(t *testing.T) {
	minimum := 2
	source := &AttributeExpr{
		Type: String,
		Docs: &DocsExpr{Description: "source"},
		Validation: &ValidationExpr{
			MinLength: &minimum,
			Values:    []any{"first", "second"},
		},
		DefaultValue: []any{map[string]any{"message": "invalid"}},
		UserExamples: []*ExampleExpr{{Value: map[string]any{"message": "invalid"}}},
	}

	effective := effectiveErrorAttribute(source)
	*effective.Validation.MinLength = 5
	effective.Validation.Values[0] = "changed"
	effective.DefaultValue.([]any)[0].(map[string]any)["message"] = "changed"
	effective.Docs.Description = "changed"
	effective.UserExamples[0].Value.(map[string]any)["message"] = "changed"

	require.Equal(t, 2, *source.Validation.MinLength)
	require.Equal(t, "first", source.Validation.Values[0])
	require.Equal(t, "invalid", source.DefaultValue.([]any)[0].(map[string]any)["message"])
	require.Equal(t, "source", source.Docs.Description)
	require.Equal(t, "invalid", source.UserExamples[0].Value.(map[string]any)["message"])
}

func TestEffectiveErrorAttributeReconnectsCopiesByOrigin(t *testing.T) {
	source := recursiveErrorType("RecursiveError")
	first := Dup(source).(UserType)
	second := Dup(source).(UserType)
	root := &AttributeExpr{Type: &Object{
		{Name: "first", Attribute: &AttributeExpr{Type: first}},
		{Name: "second", Attribute: &AttributeExpr{Type: second}},
	}}

	effective := effectiveErrorAttribute(root)
	firstCopy := effective.Find("first").Type.(UserType)
	secondCopy := effective.Find("second").Type.(UserType)

	require.Same(t, firstCopy, secondCopy)
	require.Same(t, firstCopy, firstCopy.Origin())
	require.NotSame(t, source, firstCopy.Origin())
}

// errorBase returns an object declaration whose only field has the given
// primitive type. The shared type name proves comparison uses effective shape.
func errorBase(name string, fieldType DataType) *UserTypeExpr {
	return &UserTypeExpr{
		TypeName: name,
		AttributeExpr: &AttributeExpr{Type: &Object{
			{Name: "value", Attribute: &AttributeExpr{Type: fieldType}},
		}},
	}
}

// recursiveErrorType returns an unfinalized self-referential declaration.
func recursiveErrorType(name string) *UserTypeExpr {
	result := &UserTypeExpr{TypeName: name}
	result.AttributeExpr = &AttributeExpr{Type: &Object{
		{Name: "message", Attribute: &AttributeExpr{Type: String}},
		{Name: "next", Attribute: &AttributeExpr{Type: result}},
	}}
	return result
}

// referencedErrorContracts returns one inherited and one explicit error with
// the same field validation, default, and generated Go name.
func referencedErrorContracts() (*AttributeExpr, *AttributeExpr) {
	referenceMinimum := 2
	field := &AttributeExpr{
		Type:         String,
		Validation:   &ValidationExpr{MinLength: &referenceMinimum},
		DefaultValue: "invalid",
		Meta:         MetaExpr{"struct:field:name": {"Message"}},
	}
	reference := &UserTypeExpr{
		TypeName: "ReferenceError",
		AttributeExpr: &AttributeExpr{Type: &Object{
			{Name: "message", Attribute: field},
		}},
	}
	referenced := &AttributeExpr{
		Type: &Object{
			{Name: "message", Attribute: &AttributeExpr{Type: String}},
		},
		References: []DataType{reference},
	}
	explicitMinimum := 2
	explicit := &AttributeExpr{Type: &Object{
		{Name: "message", Attribute: &AttributeExpr{
			Type:         String,
			Validation:   &ValidationExpr{MinLength: &explicitMinimum},
			DefaultValue: "invalid",
			Meta:         MetaExpr{"struct:field:name": {"Message"}},
		}},
	}}
	return referenced, explicit
}

// requiredObject returns the same two-field error object with the requested
// validation order so the test distinguishes authored order from semantics.
func requiredObject(required ...string) *AttributeExpr {
	return &AttributeExpr{
		Type: &Object{
			{Name: "first", Attribute: &AttributeExpr{Type: String}},
			{Name: "second", Attribute: &AttributeExpr{Type: String}},
		},
		Validation: &ValidationExpr{Required: required},
	}
}
