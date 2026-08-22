// This file verifies that symbolic validation planning preserves service and
// view validation output without reading expressions after package freeze.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

// TestValidationPlanPreservesRulesPathsAndRequiredness compares retained rule
// rendering with the existing service/view validation generator.
func TestValidationPlanPreservesRulesPathsAndRequiredness(t *testing.T) {
	minimum := 2.0
	minLength := 3
	attribute := &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "name", Attribute: &expr.AttributeExpr{
				Type: expr.String,
				Validation: &expr.ValidationExpr{
					Pattern:   "^[a-z]+$",
					MinLength: &minLength,
				},
			}},
			{Name: "count", Attribute: &expr.AttributeExpr{
				Type:       expr.Int,
				Validation: &expr.ValidationExpr{Minimum: &minimum},
			}},
			{Name: "nested", Attribute: &expr.AttributeExpr{Type: &expr.Object{}}},
		},
		Validation: &expr.ValidationExpr{Required: []string{"name", "nested"}},
	}
	policy := GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true}
	legacyContext := NewAttributeContext(
		policy.Pointer,
		policy.IgnoreRequired,
		policy.UseDefault,
		"",
		NewNameScope(),
	)
	legacyContext.UnionPointer = policy.UnionPointer
	want := ValidationCode(attribute, nil, legacyContext, true, false, true, "target")

	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: policy,
	})
	require.NoError(t, err)
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
	require.NoError(t, err)

	attribute.Validation = nil
	attribute.Type = expr.String
	minimum = 99
	minLength = 99

	linked, err := plan.Link(layout.Link("generated.local/gen/service", validationPlanTestQualifier))
	require.NoError(t, err)
	require.Equal(t, want, linked.Render("target", "target"))
}

// TestValidationPlanCopiesEnumValues verifies that accepted mutable enum
// values cannot change a validation program after planning.
func TestValidationPlanCopiesEnumValues(t *testing.T) {
	bytesValue := []byte{1, 2}
	arrayValue := []any{
		bytesValue,
		map[string]any{"nested": []any{"kept"}},
	}
	mapValue := map[string]any{"array": arrayValue}
	attribute := &expr.AttributeExpr{
		Type: expr.Any,
		Validation: &expr.ValidationExpr{Values: []any{
			bytesValue,
			arrayValue,
			mapValue,
		}},
	}
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: GoLayoutPolicy{UseDefault: true, SumType: true},
	})
	require.NoError(t, err)
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
	require.NoError(t, err)

	bytesValue[0] = 9
	arrayValue[1].(map[string]any)["nested"].([]any)[0] = "changed"
	mapValue["added"] = true
	attribute.Validation.Values[0] = "replaced"

	require.Equal(t, []any{
		[]byte{1, 2},
		[]any{
			[]byte{1, 2},
			map[string]any{"nested": []any{"kept"}},
		},
		map[string]any{"array": []any{
			[]byte{1, 2},
			map[string]any{"nested": []any{"kept"}},
		}},
	}, plan.root.rules.values)
}

// TestValidationPlanPreservesContainersUnionsAndValidatorBindings verifies all
// recursive service/view shapes retain exact nested validator declarations.
func TestValidationPlanPreservesContainersUnionsAndValidatorBindings(t *testing.T) {
	const owner = "generated.local/gen/service"
	minLength := 1
	minimum := 4.0
	child := goTypeTestUserType("Child", &expr.Object{
		{Name: "name", Attribute: &expr.AttributeExpr{
			Type:       expr.String,
			Validation: &expr.ValidationExpr{MinLength: &minLength},
		}},
	})
	union := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "label", Attribute: &expr.AttributeExpr{
				Type:       expr.String,
				Validation: &expr.ValidationExpr{Pattern: ".+"},
			}},
			{Name: "child", Attribute: &expr.AttributeExpr{Type: child}},
		},
	}
	mapKey := &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{MinLength: &minLength},
	}
	mapValue := &expr.AttributeExpr{
		Type:       expr.Int,
		Validation: &expr.ValidationExpr{Minimum: &minimum},
	}
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "children", Attribute: &expr.AttributeExpr{Type: &expr.Array{
			ElemType: &expr.AttributeExpr{Type: child},
		}}},
		{Name: "labels", Attribute: &expr.AttributeExpr{Type: &expr.Map{
			KeyType:  mapKey,
			ElemType: mapValue,
		}}},
		{Name: "choice", Attribute: &expr.AttributeExpr{Type: union}},
	}}
	policy := GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true}
	legacyContext := NewAttributeContext(
		policy.Pointer,
		policy.IgnoreRequired,
		policy.UseDefault,
		"",
		NewNameScope(),
	)
	legacyContext.UnionPointer = policy.UnionPointer
	want := ValidationCode(attribute, nil, legacyContext, true, false, true, "target")

	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	childDeclaration := declareGoTypeTestUserType(t, generation, owner, child)
	unionDeclaration := declareGoTypeTestUnion(t, generation, owner, union)
	generatedPackage := generation.Package(owner)
	validator := NewExactName(NameFunction, "ValidateChild")
	require.NoError(t, generatedPackage.DeclareName(validator))

	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  owner,
		Policy: policy,
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			child: {Owner: owner, Type: childDeclaration},
			union: {Owner: owner, Union: unionDeclaration},
		}),
	})
	require.NoError(t, err)
	plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{
		Required: true,
		Bind: func(request ValidatorBindingRequest) (*NameDeclaration, error) {
			require.Same(t, child, request.Attribute.Type)
			require.Equal(t, owner, request.Layout.Owner())
			require.Empty(t, request.View)
			return validator, nil
		},
	})
	require.NoError(t, err)
	require.Equal(t, []*NameDeclaration{validator, validator}, plan.ValidatorDeclarations())

	mapKey.Validation = nil
	mapValue.Validation = nil
	union.Values = nil
	child.SetAttribute(&expr.AttributeExpr{Type: expr.String})
	attribute.Type = expr.String

	require.NoError(t, generation.Freeze())
	linked, err := plan.Link(layout.Link(owner, validationPlanTestQualifier))
	require.NoError(t, err)
	require.Equal(t, want, linked.Render("target", "target"))
	require.Empty(t, linked.Imports())
}

// TestValidationPlanRejectsUnboundNestedValidator verifies planning never
// falls back to reconstructing a validator name from a user type.
func TestValidationPlanRejectsUnboundNestedValidator(t *testing.T) {
	minLength := 1
	child := goTypeTestUserType("Child", &expr.Object{
		{Name: "name", Attribute: &expr.AttributeExpr{
			Type:       expr.String,
			Validation: &expr.ValidationExpr{MinLength: &minLength},
		}},
	})
	attribute := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "child", Attribute: &expr.AttributeExpr{Type: child}},
	}}
	generation, err := NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	declaration := declareGoTypeTestUserType(t, generation, "generated.local/gen/service", child)
	layout, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:  "generated.local/gen/service",
		Policy: GoLayoutPolicy{Pointer: true, UseDefault: true, SumType: true},
		Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
			child: {Owner: "generated.local/gen/service", Type: declaration},
		}),
	})
	require.NoError(t, err)

	_, err = NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
	require.EqualError(t, err, "plan validation for field \"child\": validator binder must not be nil")
}

// validationPlanTestQualifier resolves the focused generated package aliases.
func validationPlanTestQualifier(importPath string) string {
	switch importPath {
	case "generated.local/gen/service":
		return "service"
	default:
		panic(fmt.Sprintf("unexpected validation import %q", importPath))
	}
}
