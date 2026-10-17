// These tests preserve authored default representations while mapping explicit
// HTTP property names, including named and pointer union discriminators.
package openapi

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestProjectDefaultPreservesScalarRepresentations(t *testing.T) {
	text := "abc"
	for _, value := range []any{text, &text, []byte(text)} {
		require.Equal(t, value, ProjectDefault(&expr.AttributeExpr{Type: expr.Bytes}, value))
	}
	object := &expr.Object{
		{Name: "display_name:displayName", Attribute: &expr.AttributeExpr{Type: expr.String}},
		{Name: "content", Attribute: &expr.AttributeExpr{Type: expr.Bytes}},
	}
	for _, value := range []any{text, &text, []byte(text)} {
		input := map[string]any{"display_name": "first", "content": value}
		output := ProjectDefault(&expr.AttributeExpr{Type: object}, &input)
		require.Equal(t, map[string]any{"displayName": "first", "content": value}, output)
		require.Contains(t, input, "display_name")
	}
}

func TestMappedUnionValuesPreserveDiscriminators(t *testing.T) {
	type branchName string
	name := branchName("first")
	branch := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "display_name:displayName", Attribute: &expr.AttributeExpr{Type: expr.String}},
	}}
	union := &expr.Union{Values: []*expr.NamedAttributeExpr{{Name: "first", Attribute: branch}}}
	attribute := &expr.AttributeExpr{Type: union}
	for _, discriminator := range []any{name, &name} {
		input := map[string]any{"type": discriminator, "value": map[string]any{"display_name": "first"}}
		expected := map[string]any{"type": discriminator, "value": map[string]any{"displayName": "first"}}
		require.Equal(t, expected, ProjectDefault(attribute, &input))
		example := ProjectExample(attribute, &input).(map[string]any)
		require.Equal(t, map[string]any{"displayName": "first"}, example["value"])
	}
	plain := &expr.AttributeExpr{Type: &expr.Union{Values: []*expr.NamedAttributeExpr{{Name: "first", Attribute: &expr.AttributeExpr{Type: expr.String}}}}}
	value := map[string]any{"type": &name, "value": "unchanged"}
	require.Equal(t, value, ProjectExample(plain, value))
	require.Equal(t, value, ProjectDefault(plain, value))
}

func TestProjectDefaultStructFields(t *testing.T) {
	type (
		childDefault struct {
			DisplayName string
			CustomLabel string
		}
		parentDefault struct {
			Child *childDefault
		}
	)
	child := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "display_name:displayName", Attribute: &expr.AttributeExpr{Type: expr.String}},
		{Name: "label:label", Attribute: &expr.AttributeExpr{Type: expr.String, Meta: expr.MetaExpr{"struct:field:name": {"custom_label"}}}},
	}}
	value := childDefault{DisplayName: "example", CustomLabel: "custom"}
	expected := map[string]any{"displayName": "example", "label": "custom"}
	for _, input := range []any{value, &value} {
		require.Equal(t, expected, ProjectDefault(child, input))
	}
	parent := &expr.AttributeExpr{Type: &expr.Object{{Name: "child", Attribute: child}}}
	require.Equal(t, map[string]any{"child": expected}, ProjectDefault(parent, &parentDefault{Child: &value}))
}

func TestMappedUnionExamplesPreserveUnmappedBranches(t *testing.T) {
	object := &expr.AttributeExpr{Type: &expr.Object{{Name: "display_name:displayName", Attribute: &expr.AttributeExpr{Type: expr.String}}}}
	for _, untagged := range []bool{false, true} {
		union := &expr.Union{
			Untagged: untagged,
			Values: []*expr.NamedAttributeExpr{
				{Name: "object", Attribute: object},
				{Name: "bytes", Attribute: &expr.AttributeExpr{Type: expr.Bytes}},
			},
		}
		attribute := &expr.AttributeExpr{Type: union}
		var value any = "abc"
		if !untagged {
			value = map[string]any{"type": "bytes", "value": value}
		}
		require.Equal(t, value, ProjectExample(attribute, value))
		require.Equal(t, value, ProjectDefault(attribute, value))
	}
}
