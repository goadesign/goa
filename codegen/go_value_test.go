// This file verifies that authored defaults become specialized Go values with
// the same names and pointer layout as generated service types.
package codegen

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestRenderGoValueUsesPlannedFieldLayout(t *testing.T) {
	status := &expr.UserTypeExpr{
		TypeName: "Status",
		AttributeExpr: &expr.AttributeExpr{
			Type: expr.String,
		},
	}
	configObject := expr.Object{
		&expr.NamedAttributeExpr{Name: "status", Attribute: &expr.AttributeExpr{Type: status}},
		&expr.NamedAttributeExpr{Name: "data", Attribute: &expr.AttributeExpr{Type: expr.Bytes}},
	}
	config := &expr.UserTypeExpr{
		TypeName:      "Config",
		AttributeExpr: &expr.AttributeExpr{Type: &configObject},
	}
	layout := goValueTestLayout(t, &expr.AttributeExpr{Type: config}, GoLayoutPolicy{
		UseDefault: true,
		SumType:    true,
	}, map[expr.DataType]GoTypeBinding{
		status: goValueTestTypeBinding(t, status),
		config: goValueTestTypeBinding(t, config),
	})

	code, err := RenderGoValue(
		&expr.AttributeExpr{Type: config},
		map[string]any{"status": "ready", "data": "ok"},
		layout,
		true,
		nil,
		"defaultValue",
	)
	require.NoError(t, err)
	require.Equal(t, []string{`var defaultValue1 Status = "ready"`}, code.Declarations)
	require.Equal(t, `&Config{Status: &defaultValue1, Data: []byte("ok")}`, code.Expression)
}

func TestRenderGoValueUsesPlannedUnionConstructor(t *testing.T) {
	choice := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
		},
	}}
	layout := goValueTestLayout(t, choice, GoLayoutPolicy{UseDefault: true, SumType: true}, map[expr.DataType]GoTypeBinding{
		choice.Type: goValueTestUnionBinding(t, choice),
	})

	code, err := RenderGoValue(
		choice,
		map[string]any{"type": "text", "value": "ready"},
		layout,
		false,
		func(attribute *expr.AttributeExpr, branch string) (string, error) {
			require.Same(t, choice, attribute)
			require.Equal(t, "text", branch)
			return "NewChoiceText", nil
		},
		"defaultValue",
	)
	require.NoError(t, err)
	require.Empty(t, code.Declarations)
	require.Equal(t, `NewChoiceText("ready")`, code.Expression)
}

func TestRenderGoValueUsesTypedCustomDefault(t *testing.T) {
	attribute := &expr.AttributeExpr{
		Type: expr.Bytes,
		Meta: expr.MetaExpr{
			"struct:field:type": {"json.RawMessage", "encoding/json", "json"},
		},
	}
	layout := goValueTestLayout(t, attribute, GoLayoutPolicy{UseDefault: true, SumType: true}, nil)

	code, err := RenderGoValue(attribute, json.RawMessage("ok"), layout, false, nil, "defaultValue")
	require.NoError(t, err)
	require.Empty(t, code.Declarations)
	require.Equal(t, `json.RawMessage{0x6f, 0x6b}`, code.Expression)

	code, err = RenderGoValue(attribute, "ok", layout, false, nil, "defaultValue")
	require.NoError(t, err)
	require.Equal(t, `[]byte("ok")`, code.Expression)
}

func TestRenderGoValueConvertsNativeCustomDefault(t *testing.T) {
	attribute := &expr.AttributeExpr{
		Type: expr.Int64,
		Meta: expr.MetaExpr{
			"struct:field:type": {"flag.ErrorHandling", "flag"},
		},
	}
	layout := goValueTestLayout(t, attribute, GoLayoutPolicy{UseDefault: true, SumType: true}, nil)

	code, err := RenderGoValue(attribute, 1, layout, false, nil, "defaultValue")
	require.NoError(t, err)
	require.Equal(t, `1`, code.Expression)
}

func TestRenderGoValueRendersAnyObjectsWithStringKeys(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.Any}
	layout := goValueTestLayout(t, attribute, GoLayoutPolicy{UseDefault: true, SumType: true}, nil)

	code, err := RenderGoValue(attribute, map[string]any{"count": 1}, layout, false, nil, "defaultValue")
	require.NoError(t, err)
	require.Equal(t, `map[string]any{"count":1}`, code.Expression)

	_, err = RenderGoValue(attribute, map[int]any{1: "invalid"}, layout, false, nil, "defaultValue")
	require.EqualError(t, err, "render Go value for any: Any object default key must be a string")
}

func TestRenderGoValueUsesContainerElementLayouts(t *testing.T) {
	choice := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
		},
	}}
	tests := []struct {
		name      string
		attribute *expr.AttributeExpr
		value     any
		want      string
	}{
		{
			name: "array of OneOf values",
			attribute: &expr.AttributeExpr{Type: &expr.Array{
				ElemType: choice,
			}},
			value: []any{map[string]any{"type": "text", "value": "ready"}},
			want:  `[]Choice{NewChoiceText("ready")}`,
		},
		{
			name: "map of OneOf values",
			attribute: &expr.AttributeExpr{Type: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.String},
				ElemType: choice,
			}},
			value: map[string]any{"mode": map[string]any{"type": "text", "value": "ready"}},
			want:  `map[string]Choice{"mode":NewChoiceText("ready")}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			layout := goValueTestLayout(t, test.attribute, GoLayoutPolicy{UseDefault: true, SumType: true}, map[expr.DataType]GoTypeBinding{
				choice.Type: goValueTestUnionBinding(t, choice),
			})
			code, err := RenderGoValue(
				test.attribute,
				test.value,
				layout,
				false,
				func(_ *expr.AttributeExpr, _ string) (string, error) {
					return "NewChoiceText", nil
				},
				"defaultValue",
			)
			require.NoError(t, err)
			require.Equal(t, test.want, code.Expression)
		})
	}
}

func TestRenderGoValueUsesPointerBackedArrayElements(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: &expr.Array{
		ElemType:         &expr.AttributeExpr{Type: expr.String},
		NonNullableElems: true,
	}}
	layout := goValueTestLayout(t, attribute, GoLayoutPolicy{
		UseDefault:          true,
		SumType:             true,
		ArrayElementPointer: true,
	}, nil)

	code, err := RenderGoValue(attribute, []string{"ready"}, layout, false, nil, "defaultValue")
	require.NoError(t, err)
	require.Equal(t, []string{`var defaultValue1 string = "ready"`}, code.Declarations)
	require.Equal(t, `[]*string{&defaultValue1}`, code.Expression)
}

func TestRenderGoValueAllocatesMapLocalsInKeyOrder(t *testing.T) {
	valueObject := &expr.Object{{Name: "label", Attribute: &expr.AttributeExpr{Type: expr.String}}}
	attribute := &expr.AttributeExpr{Type: &expr.Map{
		KeyType:  &expr.AttributeExpr{Type: expr.String},
		ElemType: &expr.AttributeExpr{Type: valueObject},
	}}
	layout := goValueTestLayout(t, attribute, GoLayoutPolicy{
		Pointer:    true,
		UseDefault: true,
		SumType:    true,
	}, nil)
	value := map[string]any{
		"z": map[string]any{"label": "last"},
		"a": map[string]any{"label": "first"},
	}

	for range 20 {
		code, err := RenderGoValue(attribute, value, layout, false, nil, "defaultValue")
		require.NoError(t, err)
		require.Equal(t, []string{
			`var defaultValue1 string = "first"`,
			`var defaultValue2 string = "last"`,
		}, code.Declarations)
		require.Equal(t, `map[string]*struct {
	Label *string
}{"a":&struct {
	Label *string
}{Label: &defaultValue1}, "z":&struct {
	Label *string
}{Label: &defaultValue2}}`, code.Expression)
	}
}

func goValueTestLayout(t *testing.T, attribute *expr.AttributeExpr, policy GoLayoutPolicy, bindings map[expr.DataType]GoTypeBinding) LinkedGoType {
	t.Helper()
	plan, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:            "generated.local/gen/service",
		Policy:           policy,
		Bind:             goTypeTestBinder(bindings),
		RetainNamedValue: true,
	})
	require.NoError(t, err)
	return plan.Link("generated.local/gen/service", func(importPath string) string {
		if importPath == "encoding/json" {
			return "json"
		}
		return goTypeTestQualifier(importPath)
	})
}

func goValueTestTypeBinding(t *testing.T, userType expr.UserType) GoTypeBinding {
	t.Helper()
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	binding := GoTypeBinding{
		Owner: "generated.local/gen/service",
		Type:  declareGoTypeTestUserType(t, generation, "generated.local/gen/service", userType),
	}
	require.NoError(t, generation.Freeze())
	return binding
}

func goValueTestUnionBinding(t *testing.T, attribute *expr.AttributeExpr) GoTypeBinding {
	t.Helper()
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	binding := GoTypeBinding{
		Owner: "generated.local/gen/service",
		Union: declareGoTypeTestUnion(t, generation, "generated.local/gen/service", attribute.Type.(*expr.Union)),
	}
	require.NoError(t, generation.Freeze())
	return binding
}
