// This file verifies that private expression snapshots remain exact copies of
// the authored graph while callers may safely mutate either graph afterward.
package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	attributeGraphPublicValue struct {
		Values []string
	}

	attributeGraphPrivateValue struct {
		values []string
	}
)

func TestAttributeGraphCopierPreservesCycles(t *testing.T) {
	node := &UserTypeExpr{TypeName: "Node", UID: "node"}
	nodeObject := &Object{}
	node.SetAttribute(&AttributeExpr{Type: nodeObject})
	nodeObject.Set("next", &AttributeExpr{Type: node})
	original := &AttributeExpr{Type: node}

	copied := NewAttributeGraphCopier().Copy(original)
	copiedNode := copied.Type.(UserType)
	copiedNext := AsObject(copiedNode).Attribute("next").Type.(UserType)

	require.NotSame(t, node, copiedNode)
	require.Same(t, copiedNode, copiedNext)
}

func TestAttributeGraphCopierKeepsDistinctCopiesOfOneUserType(t *testing.T) {
	authored := &UserTypeExpr{
		AttributeExpr: &AttributeExpr{Type: String},
		TypeName:      "Value",
		UID:           "value",
	}
	first := authored.Dup(&AttributeExpr{Type: String})
	second := authored.Dup(&AttributeExpr{Type: Int})
	original := &AttributeExpr{Type: &Object{
		{Name: "first", Attribute: &AttributeExpr{Type: first}},
		{Name: "second", Attribute: &AttributeExpr{Type: second}},
	}}

	copied := NewAttributeGraphCopier().Copy(original)
	copiedObject := copied.Type.(*Object)
	copiedFirst := copiedObject.Attribute("first").Type.(UserType)
	copiedSecond := copiedObject.Attribute("second").Type.(UserType)

	require.NotSame(t, copiedFirst, copiedSecond)
	require.Same(t, authored, copiedFirst.Origin())
	require.Same(t, authored, copiedSecond.Origin())
	require.Equal(t, String, copiedFirst.Attribute().Type)
	require.Equal(t, Int, copiedSecond.Attribute().Type)
}

func TestAttributeGraphCopierDetachesResultTypeViews(t *testing.T) {
	viewAttribute := &AttributeExpr{
		Type: &Object{{
			Name: "name",
			Attribute: &AttributeExpr{
				Type:        String,
				Description: "input field",
			},
		}},
		Description: "input view",
		Meta:        MetaExpr{"marker": {"input"}},
	}
	result := NewResultTypeExpr("Item", "application/vnd.item", nil)
	result.ContentType = "application/json"
	result.Views = []*ViewExpr{{
		AttributeExpr: viewAttribute,
		Name:          "default",
		Parent:        result,
	}}
	input := &AttributeExpr{Type: result}

	working := NewAttributeGraphCopier().Copy(input).Type.(*ResultTypeExpr)
	baseline := NewAttributeGraphCopier().Copy(&AttributeExpr{Type: working}).Type.(*ResultTypeExpr)
	workingView := working.Views[0]
	baselineView := baseline.Views[0]

	require.Equal(t, "application/json", working.ContentType)
	require.Equal(t, "application/json", baseline.ContentType)
	require.NotSame(t, result, working)
	require.NotSame(t, result.Views[0], workingView)
	require.NotSame(t, result.Views[0].AttributeExpr, workingView.AttributeExpr)
	require.NotSame(t, workingView, baselineView)
	require.NotSame(t, workingView.AttributeExpr, baselineView.AttributeExpr)
	require.Same(t, working, workingView.Parent)
	require.Same(t, baseline, baselineView.Parent)

	working.ContentType = "text/plain"
	workingView.Name = "working"
	workingView.Description = "working view"
	workingView.Meta["marker"][0] = "working"
	AsObject(workingView.Type).Attribute("name").Description = "working field"

	require.Equal(t, "application/json", result.ContentType)
	require.Equal(t, "default", result.Views[0].Name)
	require.Equal(t, "input view", result.Views[0].Description)
	require.Equal(t, "input", result.Views[0].Meta["marker"][0])
	require.Equal(t, "input field", AsObject(result.Views[0].Type).Attribute("name").Description)
	require.Equal(t, "application/json", baseline.ContentType)
	require.Equal(t, "default", baselineView.Name)
	require.Equal(t, "input view", baselineView.Description)
	require.Equal(t, "input", baselineView.Meta["marker"][0])
	require.Equal(t, "input field", AsObject(baselineView.Type).Attribute("name").Description)
}

func TestAttributeGraphCopierMapsCopiesToOriginals(t *testing.T) {
	originalChild := &AttributeExpr{Type: String}
	original := &AttributeExpr{
		Type:      &Object{{Name: "child", Attribute: originalChild}},
		finalized: true,
	}
	copier := NewAttributeGraphCopier()

	copied := copier.Copy(original)
	copiedChild := copied.Type.(*Object).Attribute("child")

	require.NotSame(t, original, copied)
	require.Same(t, copied, copier.Copy(copied))
	require.Same(t, original, copier.Original(copied))
	require.Same(t, originalChild, copier.Original(copiedChild))
	require.Same(t, original, copied.AuthoredAttribute())
	require.Same(t, originalChild, copiedChild.AuthoredAttribute())
	require.True(t, copied.finalized)
}

func TestAttributeGraphCopierDetachesMutableValues(t *testing.T) {
	minimum := 1.0
	maximumLength := 3
	original := &AttributeExpr{
		Type: String,
		Docs: &DocsExpr{Description: "original"},
		Validation: &ValidationExpr{
			Values:    []any{map[string]any{"items": []string{"first"}}},
			Minimum:   &minimum,
			MaxLength: &maximumLength,
			Required:  []string{"value"},
		},
		Meta:         MetaExpr{"struct:tag": {"one", "two"}},
		DefaultValue: map[string]any{"items": []string{"first"}},
		UserExamples: []*ExampleExpr{{Value: map[string]any{"items": []string{"first"}}}},
	}

	copied := NewAttributeGraphCopier().Copy(original)
	copied.Docs.Description = "copied"
	copied.Validation.Values[0].(map[string]any)["items"].([]string)[0] = "copied"
	*copied.Validation.Minimum = 2
	*copied.Validation.MaxLength = 4
	copied.Validation.Required[0] = "copied"
	copied.Meta["struct:tag"][0] = "copied"
	copied.DefaultValue.(map[string]any)["items"].([]string)[0] = "copied"
	copied.UserExamples[0].Value.(map[string]any)["items"].([]string)[0] = "copied"

	require.Equal(t, "original", original.Docs.Description)
	require.Equal(t, "first", original.Validation.Values[0].(map[string]any)["items"].([]string)[0])
	require.Equal(t, 1.0, *original.Validation.Minimum)
	require.Equal(t, 3, *original.Validation.MaxLength)
	require.Equal(t, "value", original.Validation.Required[0])
	require.Equal(t, "one", original.Meta["struct:tag"][0])
	require.Equal(t, "first", original.DefaultValue.(map[string]any)["items"].([]string)[0])
	require.Equal(t, "first", original.UserExamples[0].Value.(map[string]any)["items"].([]string)[0])
}

func TestAttributeGraphCopierCopiesTypedValues(t *testing.T) {
	values := []string{"original"}
	object := attributeGraphPublicValue{Values: []string{"original"}}
	original := &AttributeExpr{Type: String, DefaultValue: [2]any{&values, object}}

	copied := NewAttributeGraphCopier().Copy(original)
	values[0] = "changed"
	object.Values[0] = "changed"
	copiedValues := copied.DefaultValue.([2]any)

	require.Equal(t, "original", (*copiedValues[0].(*[]string))[0])
	require.Equal(t, "original", copiedValues[1].(attributeGraphPublicValue).Values[0])
}

func TestAttributeGraphCopierRejectsUnsupportedMutableValues(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		message string
	}{
		{
			name:    "channel",
			value:   make(chan int),
			message: "cannot copy attribute value of type chan int",
		},
		{
			name:    "private mutable field",
			value:   attributeGraphPrivateValue{values: []string{"value"}},
			message: "cannot copy attribute value of type expr.attributeGraphPrivateValue: unexported field values contains mutable data",
		},
		{
			name:    "function",
			value:   func() {},
			message: "cannot copy attribute value of type func()",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.PanicsWithValue(t, test.message, func() {
				NewAttributeGraphCopier().Copy(&AttributeExpr{Type: String, DefaultValue: test.value})
			})
		})
	}
}

func TestAttributeGraphCopierRejectsCyclicMutableValue(t *testing.T) {
	value := make(map[string]any)
	value["self"] = value

	require.PanicsWithValue(
		t,
		"cannot copy cyclic attribute value of type map[string]interface {}",
		func() {
			NewAttributeGraphCopier().Copy(&AttributeExpr{Type: String, DefaultValue: value})
		},
	)
}
