// This file verifies that reusing a generated function removes only the
// bodies it owns. Other callers of a shared child retain their conversion.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestTransformPlanUsesExistingFunction(t *testing.T) {
	for _, shared := range []bool{false, true} {
		t.Run(fmt.Sprintf("shared_child=%t", shared), func(t *testing.T) {
			leaf := &expr.UserTypeExpr{TypeName: "Leaf", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
				{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
			}}}
			middle := &expr.UserTypeExpr{TypeName: "Middle", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
				{Name: "leaf", Attribute: &expr.AttributeExpr{Type: leaf}},
			}}}
			fields := &expr.Object{{Name: "middle", Attribute: &expr.AttributeExpr{Type: middle}}}
			if shared {
				fields.Set("direct", &expr.AttributeExpr{Type: leaf})
			}
			root := &expr.UserTypeExpr{TypeName: "Root", AttributeExpr: &expr.AttributeExpr{Type: fields}}
			attribute := &expr.AttributeExpr{Type: root}
			plan, err := NewTransformPlan(attribute, attribute, "", nil)
			require.NoError(t, err)
			owner := newGeneratedPackage("test", "example.com/test", "gen")
			existing := NewExactName(NameFunction, "copyMiddle")
			require.NoError(t, owner.DeclareName(existing))
			require.NoError(t, owner.freeze())
			found := false
			var existingID TransformHelperDefinitionID
			for _, definition := range plan.HelperDefinitions() {
				if definition.Source.Type.Name() == "Middle" {
					require.NoError(t, plan.UseExistingHelperDefinition(definition.ID, existing))
					found = true
					existingID = definition.ID
				}
			}
			require.True(t, found)
			want := 0
			if shared {
				want = 1
			}
			require.Len(t, plan.HelperDefinitions(), want)
			require.Len(t, plan.Helpers(), want)
			code, helpers := renderTransformPlan(t, plan)
			require.Contains(t, code, "copyMiddle(source.Middle)")
			require.Len(t, helpers, want)
			source := `package test
 type Leaf struct { Value *string }
 type Middle struct { Leaf *Leaf }
 type Root struct { Middle *Middle; Direct *Leaf }
 func copyMiddle(v *Middle) *Middle { return &Middle{Leaf: v.Leaf} }
 func copyRoot(source *Root) *Root {
 ` + code + "\nreturn target\n}\n"
			for _, helper := range helpers {
				require.NotEqual(t, "copyMiddle", helper.Name)
				source += "func " + helper.Name + "(v " + helper.ParamTypeRef + ") " + helper.ResultTypeRef + " {\n" + helper.Code + "\nreturn res\n}\n"
			}
			compileTransformSource(t, source)
			_, _, err = plan.Render("source", "target", true)
			require.NoError(t, err)
			require.EqualError(t, plan.UseExistingHelperDefinition(existingID, existing), "existing transform functions must be selected before contexts are bound")
		})
	}
}
