// These tests distinguish a union's authored identity from its JSON mapping
// and verify that a flattened default calls the existing typed constructor.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestFlattenedUnionIdentityAndDefault(t *testing.T) {
	object := &expr.UserTypeExpr{TypeName: "Complete", AttributeExpr: &expr.AttributeExpr{
		Type:       &expr.Object{{Name: "reference", Attribute: &expr.AttributeExpr{Type: expr.String}}},
		Validation: &expr.ValidationExpr{Required: []string{"reference"}},
	}}
	tagged := &expr.Union{TypeName: "Choice", TypeKey: "resultType", Values: []*expr.NamedAttributeExpr{{Name: "complete", Attribute: &expr.AttributeExpr{Type: object}}}}
	flat := expr.DupAtt(&expr.AttributeExpr{Type: tagged}).Type.(*expr.Union)
	flat.Flatten = true
	require.Equal(t, tagged.Hash(), flat.Hash())
	require.NotEqual(t, NewUnionTypeID(tagged), NewUnionTypeID(flat))
	attribute := &expr.AttributeExpr{Type: flat}
	complete := flat.Values[0].Attribute.Type.(expr.UserType)
	layout := goValueTestLayout(t, attribute, GoLayoutPolicy{UseDefault: true, SumType: true}, map[expr.DataType]GoTypeBinding{
		flat:     goValueTestUnionBinding(t, attribute),
		complete: goValueTestTypeBinding(t, complete),
	})
	value, err := RenderGoValue(attribute, map[any]any{"resultType": "complete", "reference": "done"}, layout, false,
		func(_ *expr.AttributeExpr, branch string) (string, error) {
			require.Equal(t, "complete", branch)
			return "NewChoiceComplete", nil
		}, "defaultValue")
	require.NoError(t, err)
	require.Empty(t, value.Declarations)
	require.Equal(t, `NewChoiceComplete(&Complete{Reference: "done"})`, value.Expression)
}
