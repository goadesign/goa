// These tests verify that validation follows the concrete union representation:
// service unions are values, while transport unions may use pointers for presence.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	dsl "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestUnionValidationPreservesValueContextForRequiredOnlyObjectBranches(t *testing.T) {
	root := RunDSL(t, requiredObjectUnionDSL)
	scope := NewNameScope()
	unionT := root.UserType("UnionUserValidate")
	att := &expr.AttributeExpr{Type: unionT}

	valueCode := ValidationCode(att, nil, NewAttributeContext(false, false, false, "", scope), true, false, false, "target")
	require.NotContains(t, valueCode, "ValidateSomeType(actual)")
	require.NotContains(t, valueCode, "ValidateSomeOtherType(actual)")
	require.NotContains(t, valueCode, "if actual != nil")

	pointerCode := ValidationCode(att, nil, NewAttributeContext(true, false, false, "", scope), true, false, false, "target")
	require.Contains(t, pointerCode, "ValidateSomeType(actual)")
	require.Contains(t, pointerCode, "ValidateSomeOtherType(actual)")
}

func TestUnionValidationUsesGeneratedFieldRepresentation(t *testing.T) {
	union := &expr.Union{
		TypeName: "Scope",
		Values: []*expr.NamedAttributeExpr{
			{
				Name: "description",
				Attribute: &expr.AttributeExpr{
					Type:       expr.String,
					Validation: &expr.ValidationExpr{Pattern: ".+"},
				},
			},
		},
	}
	attribute := &expr.AttributeExpr{
		Type: &expr.Object{
			{Name: "optional", Attribute: &expr.AttributeExpr{Type: union}},
			{Name: "required", Attribute: &expr.AttributeExpr{Type: union}},
		},
		Validation: &expr.ValidationExpr{Required: []string{"required"}},
	}
	scope := NewNameScope()

	serviceCode := ValidationCode(
		attribute, nil, NewAttributeContext(false, false, true, "", scope), true, false, false, "target",
	)
	require.Contains(t, serviceCode, `if target.Required.Kind() == "" {`)
	require.NotContains(t, serviceCode, "if target.Optional != nil {")

	transportCtx := NewAttributeContext(true, false, false, "", scope)
	transportCtx.UnionPointer = true
	transportCode := ValidationCode(attribute, nil, transportCtx, true, false, false, "target")
	require.Contains(t, transportCode, "if target.Required == nil {")
	require.Contains(t, transportCode, "if target.Required != nil {")
	require.Contains(t, transportCode, "if target.Optional != nil {")
	require.NotContains(t, transportCode, `if target.Required.Kind() == "" {`)

	marshalCtx := NewAttributeContext(false, false, true, "", scope)
	marshalCtx.UnionPointer = true
	marshalCode := ValidationCode(attribute, nil, marshalCtx, true, false, false, "target")
	require.Contains(t, marshalCode, `if target.Required.Kind() == "" {`)
	require.NotContains(t, marshalCode, "if target.Required != nil {")
	require.Contains(t, marshalCode, "if target.Optional != nil {")
}

// requiredObjectUnionDSL defines a OneOf with required-only object branches so
// validation generation can distinguish pointer and value contexts.
func requiredObjectUnionDSL() {
	var someType = dsl.Type("SomeType", func() {
		dsl.Attribute("a", dsl.String)
		dsl.Required("a")
	})
	var someOtherType = dsl.Type("SomeOtherType", func() {
		dsl.Attribute("b", dsl.String)
		dsl.Required("b")
	})

	_ = dsl.Type("UnionUserValidate", func() {
		dsl.OneOf("values", func() {
			dsl.Attribute("SomeType", someType)
			dsl.Attribute("SomeOtherType", someOtherType)
		})
	})
}
