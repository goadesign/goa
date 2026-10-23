// These tests check union selection before branch field validation. Optional
// value unions may be absent; pointer unions use nil alone for absence.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/expr"
)

func TestValidationChecksUnconstrainedUnionSelection(t *testing.T) {
	choice := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "Selection",
		Values: []*expr.NamedAttributeExpr{
			{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "empty", Attribute: &expr.AttributeExpr{Type: &expr.Object{}}},
		},
	}}
	for _, test := range []struct {
		name              string
		required, pointer bool
	}{
		{name: "required value", required: true},
		{name: "optional value"},
		{name: "required pointer", required: true, pointer: true},
		{name: "optional pointer", pointer: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			attribute := &expr.AttributeExpr{Type: &expr.Object{{Name: "choice", Attribute: choice}}}
			if test.required {
				attribute.Validation = &expr.ValidationExpr{Required: []string{"choice"}}
			}
			policy := GoLayoutPolicy{UseDefault: true, SumType: true, UnionPointer: test.pointer, Pointer: test.pointer}
			require.True(t, NeedsValidation(attribute, policy))
			const owner = "generated.local/gen/selection"
			generation, err := NewGeneration("generated.local/gen", nil)
			require.NoError(t, err)
			declaration := declareGoTypeTestUnion(t, generation, owner, choice.Type.(*expr.Union))
			layout, err := PlanGoType(attribute, GoTypePlanOptions{
				Owner: owner, Policy: policy,
				Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{choice.Type: {Owner: owner, Union: declaration}}),
			})
			require.NoError(t, err)
			plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: test.required})
			require.NoError(t, err)
			require.NoError(t, generation.Freeze())
			linked, err := plan.Link(layout.Link(layout.Owner(), validationPlanTestQualifier))
			require.NoError(t, err)
			context := NewAttributeContext(test.pointer, false, true, "", NewNameScope())
			context.UnionPointer = test.pointer
			for _, source := range []string{
				linked.Render("target", "body"),
				ValidationCode(attribute, nil, context, test.required, false, false, "target"),
			} {
				require.Contains(t, source, "target.Choice.Validate()")
				if !test.required && !test.pointer {
					require.Contains(t, source, `if target.Choice.Kind() != "" {`)
				} else {
					require.NotContains(t, source, `if target.Choice.Kind() != "" {`)
				}
			}
		})
	}
	require.False(t, NeedsValidation(choice, GoLayoutPolicy{UseDefault: true}))
}
