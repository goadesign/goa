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
