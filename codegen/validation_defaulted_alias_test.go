// These tests keep validation's pointer operations consistent with generated
// fields when a primitive type supplies the field's default.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestValidationPlanDefaultedAliasField(t *testing.T) {
	for _, test := range []struct {
		name         string
		typeDefault  any
		fieldDefault any
		required     bool
		pointer      bool
		useDefault   bool
		wantPointer  bool
	}{
		{name: "inherited default", typeDefault: "brief", useDefault: true},
		{name: "field default", fieldDefault: "brief", useDefault: true},
		{name: "required alias", typeDefault: "brief", required: true, useDefault: true},
		{name: "optional alias", useDefault: true, wantPointer: true},
		{name: "defaults disabled", typeDefault: "brief", wantPointer: true},
		{name: "pointer representation", typeDefault: "brief", useDefault: true, pointer: true, wantPointer: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			const owner = "generated.local/gen/service"
			label := &expr.UserTypeExpr{
				TypeName: "Label",
				AttributeExpr: &expr.AttributeExpr{
					Type:         expr.String,
					DefaultValue: test.typeDefault,
					Validation:   &expr.ValidationExpr{Values: []any{"brief", "detailed"}},
				},
			}
			attribute := &expr.AttributeExpr{Type: &expr.Object{
				{Name: "label", Attribute: &expr.AttributeExpr{Type: label, DefaultValue: test.fieldDefault}},
			}}
			if test.required {
				attribute.Validation = &expr.ValidationExpr{Required: []string{"label"}}
			}
			generation, err := NewGeneration("generated.local/gen", nil)
			require.NoError(t, err)
			declaration := declareGoTypeTestUserType(t, generation, owner, label)
			layout, err := PlanGoType(attribute, GoTypePlanOptions{
				Owner:  owner,
				Policy: GoLayoutPolicy{Pointer: test.pointer, UseDefault: test.useDefault, SumType: true},
				Bind: goTypeTestBinder(map[expr.DataType]GoTypeBinding{
					label: {Owner: owner, Type: declaration},
				}),
			})
			require.NoError(t, err)
			require.Equal(t, test.wantPointer, layout.Fields()[0].IsPointer())
			plan, err := NewValidationPlan(attribute, layout, ValidationPlanOptions{Required: true})
			require.NoError(t, err)
			require.NoError(t, generation.Freeze())
			linked, err := plan.Link(layout.Link(owner, validationPlanTestQualifier))
			require.NoError(t, err)
			body := linked.Render("value", "body")
			if test.wantPointer {
				require.Contains(t, body, "if value.Label != nil")
				require.Contains(t, body, "string(*value.Label)")
			} else {
				require.NotContains(t, body, "value.Label != nil")
				require.NotContains(t, body, "*value.Label")
				require.Contains(t, body, "string(value.Label)")
			}
			require.Contains(t, body, "goa.InvalidEnumValueError")
		})
	}
}
