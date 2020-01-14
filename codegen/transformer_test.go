package codegen

import (
	"testing"

	"goa.design/goa/v3/expr"
)

func TestIsPrimitivePointer(t *testing.T) {
	newObj := func(fieldName string, fieldType expr.DataType, req bool) *expr.AttributeExpr {
		attr := &expr.AttributeExpr{
			Type: &expr.Object{
				&expr.NamedAttributeExpr{fieldName, &expr.AttributeExpr{Type: fieldType}},
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
			Test:     "ignore required context with pointer attribute",
			Context:  &AttributeContext{IgnoreRequired: true},
			Attr:     newObj("foo", expr.String, false),
			Name:     "foo",
			Expected: false,
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
