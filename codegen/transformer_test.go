// This file verifies shared transform rules that are independent of a
// transport generator.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"goa.design/goa/v3/expr"
)

func TestAppendHelpersUsesDeclarationIdentity(t *testing.T) {
	firstPlan := &TransformPlan{}
	secondPlan := &TransformPlan{}
	shared := NewExactName(NameFunction, "sharedHelper")
	separate := NewExactName(NameFunction, "separateHelper")
	old := []*TransformFunctionData{
		{ID: TransformHelperID{plan: firstPlan}, Declaration: shared, Name: "sharedHelper"},
		{ID: TransformHelperID{plan: firstPlan, index: 1}, Declaration: separate, Name: "separateHelper"},
		{Name: "legacyOne"},
	}
	added := []*TransformFunctionData{
		{ID: TransformHelperID{plan: secondPlan}, Declaration: shared, Name: "sharedHelper"},
		{Name: "sharedHelper"},
		{Name: "legacyOne"},
		{Name: "legacyTwo"},
	}

	got := AppendHelpers(old, added)

	if assert.Len(t, got, 5) {
		assert.Same(t, shared, got[0].Declaration)
		assert.Same(t, separate, got[1].Declaration)
		assert.Nil(t, got[2].Declaration)
		assert.Nil(t, got[3].Declaration)
		assert.Nil(t, got[4].Declaration)
		assert.Equal(t, "sharedHelper", got[3].Name)
		assert.Equal(t, "legacyTwo", got[4].Name)
	}
}

func TestAppendHelpersRejectsConflictingLegacyDefinitions(t *testing.T) {
	tests := map[string]*TransformFunctionData{
		"parameter type": {
			Name:          "transformValue",
			ParamTypeRef:  "OtherSource",
			ResultTypeRef: "Target",
			Code:          "return value",
		},
		"result type": {
			Name:          "transformValue",
			ParamTypeRef:  "Source",
			ResultTypeRef: "OtherTarget",
			Code:          "return value",
		},
		"body": {
			Name:          "transformValue",
			ParamTypeRef:  "Source",
			ResultTypeRef: "Target",
			Code:          "return other",
		},
	}
	for name, added := range tests {
		t.Run(name, func(t *testing.T) {
			old := []*TransformFunctionData{{
				Name:          "transformValue",
				ParamTypeRef:  "Source",
				ResultTypeRef: "Target",
				Code:          "return value",
			}}
			assert.PanicsWithValue(t, "transform helper \"transformValue\" has different definitions", func() {
				AppendHelpers(old, []*TransformFunctionData{added})
			})
		})
	}
}

func TestIsPrimitivePointer(t *testing.T) {
	primitiveAlias := func(name string, primitive expr.Primitive) expr.DataType {
		return &expr.UserTypeExpr{
			TypeName: name,
			UID:      name,
			AttributeExpr: &expr.AttributeExpr{
				Type: primitive,
			},
		}
	}
	newObj := func(fieldName string, fieldType expr.DataType, req bool) *expr.AttributeExpr {
		attr := &expr.AttributeExpr{
			Type: &expr.Object{
				&expr.NamedAttributeExpr{Name: fieldName, Attribute: &expr.AttributeExpr{Type: fieldType}},
			},
		}
		if req {
			attr.Validation = &expr.ValidationExpr{Required: []string{fieldName}}
		}
		return attr
	}
	tc := []struct {
		Test     string
		Context  *AttributeContext
		Attr     *expr.AttributeExpr
		Name     string
		Expected bool
	}{
		{
			Test:     "pointer attribute",
			Context:  &AttributeContext{},
			Attr:     newObj("foo", expr.String, false),
			Name:     "foo",
			Expected: true,
		},
		{
			Test:     "non pointer attribute",
			Context:  &AttributeContext{},
			Attr:     newObj("foo", expr.String, true),
			Name:     "foo",
			Expected: false,
		},
		{
			Test:     "pointer context with pointer attribute",
			Context:  &AttributeContext{Pointer: true},
			Attr:     newObj("foo", expr.String, false),
			Name:     "foo",
			Expected: true,
		},
		{
			Test:     "pointer context with non pointer attribute",
			Context:  &AttributeContext{Pointer: true},
			Attr:     newObj("foo", expr.String, true),
			Name:     "foo",
			Expected: true,
		},
		{
			Test:     "pointer context with bytes alias",
			Context:  &AttributeContext{Pointer: true},
			Attr:     newObj("foo", primitiveAlias("BytesAlias", expr.Bytes), true),
			Name:     "foo",
			Expected: false,
		},
		{
			Test:     "pointer context with any alias",
			Context:  &AttributeContext{Pointer: true},
			Attr:     newObj("foo", primitiveAlias("AnyAlias", expr.Any), true),
			Name:     "foo",
			Expected: false,
		},
		{
			Test:     "ignore required context with pointer attribute",
			Context:  &AttributeContext{IgnoreRequired: true},
			Attr:     newObj("foo", expr.String, false),
			Name:     "foo",
			Expected: true,
		},
		{
			Test:     "ignore required context with non pointer attribute",
			Context:  &AttributeContext{IgnoreRequired: true},
			Attr:     newObj("foo", expr.String, true),
			Name:     "foo",
			Expected: false,
		},
		{
			Test:     "missing attribute",
			Context:  &AttributeContext{},
			Attr:     newObj("foo", expr.String, false),
			Name:     "bar",
			Expected: false,
		},
	}
	for _, c := range tc {
		t.Run(c.Test, func(t *testing.T) {
			got := c.Context.IsPrimitivePointer(c.Name, c.Attr)
			if got != c.Expected {
				t.Errorf("expected %v, got %v", c.Expected, got)
			}
		})
	}
}
