// These tests check flattened union defaults, generated examples and exact
// expression copies without rendering an application. The selected object's
// required fields and constraints still belong to its authored type.
package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlattenedUnionDefaultsAndExamples(t *testing.T) {
	for _, test := range []struct {
		name    string
		value   map[string]any
		failure string
	}{
		{"valid", map[string]any{"resultType": "complete", "reference": "done"}, ""},
		{"missing discriminator", map[string]any{"reference": "done"}, "discriminator"},
		{"unknown discriminator", map[string]any{"resultType": "other", "reference": "done"}, "unknown OneOf branch"},
		{"missing branch field", map[string]any{"resultType": "complete"}, "missing required field"},
		{"branch constraint", map[string]any{"resultType": "complete", "reference": ""}, "length"},
		{"unknown branch field", map[string]any{"resultType": "complete", "reference": "done", "extra": true}, "unknown field"},
		{"old envelope", map[string]any{"resultType": "complete", "value": map[string]any{"reference": "done"}}, "missing required field"},
	} {
		t.Run(test.name, func(t *testing.T) {
			attribute := flattenedUnionTestAttribute()
			attribute.DefaultValue = test.value
			failures := attribute.Validate("", attribute)
			if test.failure != "" {
				require.Contains(t, failures.Error(), test.failure)
				return
			}
			require.Empty(t, failures.Errors)
			union := AsUnion(attribute.Type)
			require.True(t, union.IsCompatible(test.value))
			method := &MethodExpr{Name: "finish", Service: &ServiceExpr{Name: "operations"}}
			generator := NewExampleGenerator(NewFakerRandomizerFactory("flat")).At(MethodResultExampleIdentity(method))
			example := union.Example(generator).(map[string]any)
			require.Equal(t, "complete", example["resultType"])
			require.Contains(t, example, "reference")
			require.NotContains(t, example, "value")
		})
	}
}

// flattenedUnionTestAttribute declares one object branch with an inherited
// required string, so removing the value envelope cannot remove its checks.
func flattenedUnionTestAttribute() *AttributeExpr {
	minimum := 1
	branch := &UserTypeExpr{
		TypeName: "Complete",
		AttributeExpr: &AttributeExpr{
			Type:       &Object{{Name: "reference", Attribute: &AttributeExpr{Type: String, Validation: &ValidationExpr{MinLength: &minimum}}}},
			Validation: &ValidationExpr{Required: []string{"reference"}},
		},
	}
	return &AttributeExpr{Type: &Union{
		TypeName: "Outcome", TypeKey: "resultType", Flatten: true,
		Values: []*NamedAttributeExpr{{Name: "complete", Attribute: &AttributeExpr{Type: branch}}},
	}}
}
