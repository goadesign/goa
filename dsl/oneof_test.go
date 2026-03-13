package dsl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

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

func TestOneOfTypeConstructor(t *testing.T) {
	eval.Context = &eval.DSLContext{}
	expr.Root = &expr.RootExpr{}

	circle := Type("Circle", func() {
		Attribute("radius", Int)
	})
	square := Type("Square", func() {
		Attribute("side", Int)
	})

	dt := OneOf(circle, square)
	if eval.Context.Errors != nil {
		t.Errorf("unexpected DSL errors: %v", eval.Context.Errors)
	}

	union, ok := dt.(*expr.Union)
	if !ok {
		t.Errorf("expected *expr.Union, got %T", dt)
		return
	}
	if union.TypeName != "CircleOrSquare" {
		t.Errorf("expected union type name %q, got %q", "CircleOrSquare", union.TypeName)
	}
	if len(union.Values) != 2 {
		t.Errorf("expected 2 union branches, got %d", len(union.Values))
		return
	}
	if union.Values[0].Name != "Circle" {
		t.Errorf("expected first branch name %q, got %q", "Circle", union.Values[0].Name)
	}
	if union.Values[1].Name != "Square" {
		t.Errorf("expected second branch name %q, got %q", "Square", union.Values[1].Name)
	}
}

func TestOneOfTypeConstructorDuplicateNames(t *testing.T) {
	eval.Context = &eval.DSLContext{}
	expr.Root = &expr.RootExpr{}

	dt := OneOf(String, String)
	if eval.Context.Errors != nil {
		t.Errorf("unexpected DSL errors: %v", eval.Context.Errors)
	}

	union, ok := dt.(*expr.Union)
	if !ok {
		t.Errorf("expected *expr.Union, got %T", dt)
		return
	}
	if len(union.Values) != 2 {
		t.Errorf("expected 2 union branches, got %d", len(union.Values))
		return
	}
	if union.Values[0].Name != "String" {
		t.Errorf("expected first branch name %q, got %q", "String", union.Values[0].Name)
	}
	if union.Values[1].Name != "String2" {
		t.Errorf("expected second branch name %q, got %q", "String2", union.Values[1].Name)
	}
}

func TestOneOfTypeConstructorReservesExplicitNames(t *testing.T) {
	eval.Context = &eval.DSLContext{}
	expr.Root = &expr.RootExpr{}

	typeA := Type("TypeA", func() {
		Attribute("value", String)
	})
	typeA2 := Type("TypeA2", func() {
		Attribute("value", Int)
	})

	dt := OneOf(typeA, typeA, typeA2)
	if eval.Context.Errors != nil {
		t.Errorf("unexpected DSL errors: %v", eval.Context.Errors)
	}

	union, ok := dt.(*expr.Union)
	if !ok {
		t.Errorf("expected *expr.Union, got %T", dt)
		return
	}
	if len(union.Values) != 3 {
		t.Errorf("expected 3 union branches, got %d", len(union.Values))
		return
	}
	if union.Values[0].Name != "TypeA" {
		t.Errorf("expected first branch name %q, got %q", "TypeA", union.Values[0].Name)
	}
	if union.Values[1].Name != "TypeA3" {
		t.Errorf("expected second branch name %q, got %q", "TypeA3", union.Values[1].Name)
	}
	if union.Values[2].Name != "TypeA2" {
		t.Errorf("expected third branch name %q, got %q", "TypeA2", union.Values[2].Name)
	}
}

func TestOneOfTypeConstructorRejectsInvalidVariant(t *testing.T) {
	eval.Context = &eval.DSLContext{}
	expr.Root = &expr.RootExpr{}

	dt := OneOf(String, 42)
	union, ok := dt.(*expr.Union)
	if !ok {
		t.Errorf("expected invalid union placeholder, got %T", dt)
	}
	if eval.Context.Errors == nil {
		t.Errorf("expected DSL error for invalid OneOf variant")
	}
	if union != nil && union.TypeName != "InvalidOneOf" {
		t.Errorf("expected invalid union type name %q, got %q", "InvalidOneOf", union.TypeName)
	}
}

func TestOneOfTypeConstructorResolvesNamedUserTypes(t *testing.T) {
	eval.Context = &eval.DSLContext{}
	expr.Root = &expr.RootExpr{}

	Type("CustomType", func() {
		Attribute("value", String)
	})

	dt := OneOf("CustomType", Int)
	if eval.Context.Errors != nil {
		t.Errorf("unexpected DSL errors: %v", eval.Context.Errors)
	}

	union, ok := dt.(*expr.Union)
	if !ok {
		t.Errorf("expected *expr.Union, got %T", dt)
		return
	}
	if len(union.Values) != 2 {
		t.Errorf("expected 2 union branches, got %d", len(union.Values))
		return
	}
	if union.Values[0].Name != "CustomType" {
		t.Errorf("expected first branch name %q, got %q", "CustomType", union.Values[0].Name)
	}
	if union.Values[1].Name != "Int" {
		t.Errorf("expected second branch name %q, got %q", "Int", union.Values[1].Name)
	}
}

func TestOneOfTypeConstructorInsideAttributeWithNamedUserType(t *testing.T) {
	root := expr.RunDSL(t, func() {
		Type("CustomType", func() {
			Attribute("value", String)
		})
		Type("Parent", func() {
			Attribute("choice", OneOf("CustomType", Int))
		})
	})
	parent := root.UserType("Parent")
	if parent == nil {
		t.Errorf("expected Parent type")
		return
	}

	choice := parent.Attribute().Find("choice")
	if choice == nil {
		t.Errorf("expected choice attribute")
		return
	}
	union := expr.AsUnion(choice.Type)
	if union == nil {
		t.Errorf("expected union attribute type")
		return
	}
	if len(union.Values) != 2 {
		t.Errorf("expected 2 union branches, got %d", len(union.Values))
	}
}

func TestOneOfTypeConstructorRejectsUnnamedComplexVariants(t *testing.T) {
	eval.Context = &eval.DSLContext{}
	expr.Root = &expr.RootExpr{}

	dt := OneOf(ArrayOf(String), ArrayOf(Int))
	union, ok := dt.(*expr.Union)
	if !ok {
		t.Errorf("expected invalid union placeholder, got %T", dt)
		return
	}
	if union.TypeName != "InvalidOneOf" {
		t.Errorf("expected invalid union type name %q, got %q", "InvalidOneOf", union.TypeName)
	}
	if eval.Context.Errors == nil {
		t.Errorf("expected DSL error for unnamed complex OneOf variants")
	}
}

func TestOneOfTypeConstructorMissingTypeDoesNotFallbackToString(t *testing.T) {
	err := expr.RunInvalidDSL(t, func() {
		Type("Parent", func() {
			Attribute("choice", OneOf("MissingType", Int))
		})
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), `unknown type reference "MissingType"`)
}

func TestOneOfTypeConstructorPayloadMetaOverridesKeys(t *testing.T) {
	root := expr.RunDSL(t, func() {
		circle := Type("Circle", func() {
			Attribute("radius", Int)
		})
		square := Type("Square", func() {
			Attribute("side", Int)
		})

		Service("Shapes", func() {
			Method("draw", func() {
				Payload(OneOf(circle, square), func() {
					Meta("oneof:type:field", "kind")
					Meta("oneof:value:field", "data")
				})
			})
		})
	})

	payload := expr.AsUnion(root.Services[0].Methods[0].Payload.Type)
	if payload == nil {
		t.Errorf("expected payload union type")
		return
	}
	if payload.GetTypeKey() != "kind" {
		t.Errorf("expected payload type key %q, got %q", "kind", payload.GetTypeKey())
	}
	if payload.GetValueKey() != "data" {
		t.Errorf("expected payload value key %q, got %q", "data", payload.GetValueKey())
	}
}

func TestOneOfTypeConstructorForMethodPayloadAndResult(t *testing.T) {
	root := expr.RunDSL(t, func() {
		circle := Type("Circle", func() {
			Attribute("radius", Int)
			Required("radius")
		})
		square := Type("Square", func() {
			Attribute("side", Int)
			Required("side")
		})

		Service("Shapes", func() {
			Method("draw", func() {
				Payload(OneOf(circle, square))
				Result(OneOf(circle, square))
			})
		})
	})

	method := root.Services[0].Methods[0]
	payload := expr.AsUnion(method.Payload.Type)
	if payload == nil {
		t.Errorf("expected payload union type")
		return
	}
	if payload.TypeName != "CircleOrSquare" {
		t.Errorf("expected payload union type name %q, got %q", "CircleOrSquare", payload.TypeName)
	}

	result := expr.AsUnion(method.Result.Type)
	if result == nil {
		t.Errorf("expected result union type")
		return
	}
	if result.TypeName != "CircleOrSquare" {
		t.Errorf("expected result union type name %q, got %q", "CircleOrSquare", result.TypeName)
	}
}

func TestOneOfTypeConstructorSupportsForwardDeclaredTypeInAttribute(t *testing.T) {
	root := expr.RunDSL(t, func() {
		Type("A", func() {
			Attribute("choice", OneOf("B", Int))
		})
		Type("B", func() {
			Attribute("value", String)
			Required("value")
		})
	})

	atype := root.UserType("A")
	require.NotNil(t, atype)

	choice := atype.Attribute().Find("choice")
	require.NotNil(t, choice)

	union := expr.AsUnion(choice.Type)
	require.NotNil(t, union)
	require.Len(t, union.Values, 2)

	first, ok := union.Values[0].Attribute.Type.(expr.UserType)
	require.True(t, ok, "expected first union branch to resolve to a user type")
	if first.Name() != "B" {
		t.Errorf("expected first union branch type %q, got %q", "B", first.Name())
	}
}

func TestOneOfTypeConstructorSupportsForwardDeclaredTypeInPayloadAndResult(t *testing.T) {
	root := expr.RunDSL(t, func() {
		Service("calc", func() {
			Method("show", func() {
				Payload(OneOf("Later", Int))
				Result(OneOf("Later", Int))
			})
		})
		Type("Later", func() {
			Attribute("message", String)
			Required("message")
		})
	})

	method := root.Services[0].Methods[0]
	payload := expr.AsUnion(method.Payload.Type)
	require.NotNil(t, payload)
	require.Len(t, payload.Values, 2)

	payloadType, ok := payload.Values[0].Attribute.Type.(expr.UserType)
	require.True(t, ok, "expected payload branch to resolve to a user type")
	if payloadType.Name() != "Later" {
		t.Errorf("expected payload branch type %q, got %q", "Later", payloadType.Name())
	}

	result := expr.AsUnion(method.Result.Type)
	require.NotNil(t, result)
	require.Len(t, result.Values, 2)

	resultType, ok := result.Values[0].Attribute.Type.(expr.UserType)
	require.True(t, ok, "expected result branch to resolve to a user type")
	if resultType.Name() != "Later" {
		t.Errorf("expected result branch type %q, got %q", "Later", resultType.Name())
	}
}

func TestOneOfTypeConstructorSupportsRecursiveNamedVariants(t *testing.T) {
	root := expr.RunDSL(t, func() {
		Type("Node", func() {
			Attribute("next", OneOf("Node", "Leaf"))
		})
		Type("Leaf", func() {
			Attribute("value", String)
		})
	})

	node := root.UserType("Node")
	require.NotNil(t, node)

	next := node.Attribute().Find("next")
	require.NotNil(t, next)

	union := expr.AsUnion(next.Type)
	require.NotNil(t, union)
	require.Len(t, union.Values, 2)

	first, ok := union.Values[0].Attribute.Type.(expr.UserType)
	require.True(t, ok, "expected recursive branch to resolve to a user type")
	if first.Name() != "Node" {
		t.Errorf("expected recursive branch type %q, got %q", "Node", first.Name())
	}
}

func TestOneOfTypeConstructorReportsUnresolvedForwardType(t *testing.T) {
	err := expr.RunInvalidDSL(t, func() {
		Type("A", func() {
			Attribute("choice", OneOf("Missing", Int))
		})
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), `unknown type reference "Missing"`)
}
