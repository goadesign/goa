// Named object conversions use the fields of the defining Go struct. Additional
// constraints on a derived type do not change inherited pointer representations.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestGoTransformNamedObjectDefinitionFields(t *testing.T) {
	for _, extraRequired := range []bool{false, true} {
		for _, encode := range []bool{false, true} {
			t.Run(fmt.Sprintf("extra-required=%t/encode=%t", extraRequired, encode), func(t *testing.T) {
				attributes := make([]*expr.AttributeExpr, 2)
				for index, name := range []string{"Source", "Target"} {
					base := &expr.UserTypeExpr{
						TypeName: name + "Base",
						AttributeExpr: &expr.AttributeExpr{
							Type: &expr.Object{
								{Name: "Label", Attribute: &expr.AttributeExpr{Type: expr.String}},
								{Name: "Note", Attribute: &expr.AttributeExpr{Type: expr.String}},
								{Name: "Count", Attribute: &expr.AttributeExpr{Type: expr.Int, DefaultValue: 3}},
							},
							Validation: &expr.ValidationExpr{Required: []string{"Label"}},
						},
					}
					derived := &expr.UserTypeExpr{
						TypeName: name,
						AttributeExpr: &expr.AttributeExpr{
							Type: base,
						},
					}
					if extraRequired {
						derived.Validation = &expr.ValidationExpr{Required: []string{"Note"}}
					}
					attributes[index] = &expr.AttributeExpr{Type: derived}
				}
				scope := NewNameScope()
				sourceContext := NewAttributeContext(!encode, false, true, "", scope)
				targetContext := NewAttributeContext(encode, false, true, "", scope)
				code, helpers, err := GoTransform(attributes[0], attributes[1], "in", "out",
					sourceContext, targetContext, "", false)
				require.NoError(t, err)
				require.Empty(t, helpers)
				require.Contains(t, code, "out = &Target{")
				if encode {
					require.Contains(t, code, "Label: &in.Label,")
					require.Contains(t, code, "Count: &in.Count,")
				} else {
					require.Contains(t, code, "Label: *in.Label,")
					require.Contains(t, code, "out.Count = *in.Count")
				}
				require.Contains(t, code, "Note: in.Note,")
				require.NotContains(t, code, "Note: *in.Note,")
				require.NotContains(t, code, "Note: &in.Note,")
			})
		}
	}
}
