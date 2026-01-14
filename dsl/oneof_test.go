package dsl_test

import (
	"testing"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestOneOfCustomKeys(t *testing.T) {
	cases := []struct {
		name           string
		typeKey        string
		valueKey       string
		expectedType   string
		expectedValue  string
		expectError    bool
		errorSubstring string
	}{
		{
			name:          "custom type and value keys",
			typeKey:       "kind",
			valueKey:      "data",
			expectedType:  "kind",
			expectedValue: "data",
		},
		{
			name:          "default keys when not specified",
			typeKey:       "",
			valueKey:      "",
			expectedType:  "type",
			expectedValue: "value",
		},
		{
			name:          "custom type key only",
			typeKey:       "discriminator",
			valueKey:      "",
			expectedType:  "discriminator",
			expectedValue: "value",
		},
		{
			name:          "custom value key only",
			typeKey:       "",
			valueKey:      "payload",
			expectedType:  "type",
			expectedValue: "payload",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}

			ut := &expr.UserTypeExpr{
				AttributeExpr: &expr.AttributeExpr{
					Type: &expr.Object{},
				},
				TypeName: "TestType",
			}

			eval.Execute(func() {
				OneOf("Shape", func() {
					if tc.typeKey != "" {
						Meta("oneof:type:field", tc.typeKey)
					}
					if tc.valueKey != "" {
						Meta("oneof:value:field", tc.valueKey)
					}
					Attribute("Circle", Int)
					Attribute("Square", String)
				})
			}, ut)

			if tc.expectError {
				if eval.Context.Errors == nil {
					t.Fatal("expected DSL error, got none")
				}
				return
			}

			if eval.Context.Errors != nil {
				t.Fatalf("unexpected DSL errors: %v", eval.Context.Errors)
			}

			obj := ut.Attribute().Type.(*expr.Object)
			shapeAttr := obj.Attribute("Shape")
			if shapeAttr == nil {
				t.Fatal("Shape attribute not found")
			}

			union, ok := shapeAttr.Type.(*expr.Union)
			if !ok {
				t.Fatalf("expected Union type, got %T", shapeAttr.Type)
			}

			if union.GetTypeKey() != tc.expectedType {
				t.Errorf("expected GetTypeKey() %q, got %q", tc.expectedType, union.GetTypeKey())
			}
			if union.GetValueKey() != tc.expectedValue {
				t.Errorf("expected GetValueKey() %q, got %q", tc.expectedValue, union.GetValueKey())
			}
		})
	}
}

func TestOneOfCustomKeysSameKeyError(t *testing.T) {
	eval.Context = &eval.DSLContext{}

	ut := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{},
		},
		TypeName: "TestType",
	}

	eval.Execute(func() {
		OneOf("Shape", func() {
			Meta("oneof:type:field", "same")
			Meta("oneof:value:field", "same")
			Attribute("Circle", Int)
		})
	}, ut)

	if eval.Context.Errors == nil {
		t.Fatal("expected DSL error for same type and value keys, got none")
	}
}
