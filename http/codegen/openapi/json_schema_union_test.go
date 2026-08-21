// This file verifies shared JSON Schema rendering for unions, including that a
// typed owner keeps each discriminator paired with its generated member value.
package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestAttributeTypeSchemaCorrelatesUnionDiscriminatorAndValue(t *testing.T) {
	method := &expr.MethodExpr{Name: "union", Service: &expr.ServiceExpr{Name: "test"}}
	generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
		expr.MethodPayloadExampleIdentity(method),
	)
	schema := AttributeTypeSchema(&expr.APIExpr{}, unionAttribute(), generator)

	require.Len(t, schema.AnyOf, 2)
	assertUnionSchemaBranch(t, schema.AnyOf[0], "text", Type(String))
	assertUnionSchemaBranch(t, schema.AnyOf[1], "count", Type(Integer))
	assert.Empty(t, schema.Properties)
}

func assertUnionSchemaBranch(t *testing.T, branch *Schema, tag string, valueType Type) {
	t.Helper()
	assert.Equal(t, Type(Object), branch.Type)
	assert.Equal(t, []string{"type", "value"}, branch.Required)
	require.Contains(t, branch.Properties, "type")
	assert.Equal(t, []any{tag}, branch.Properties["type"].Enum)
	require.Contains(t, branch.Properties, "value")
	assert.Equal(t, valueType, branch.Properties["value"].Type)
}

func unionAttribute() *expr.AttributeExpr {
	return &expr.AttributeExpr{
		Type: &expr.Union{
			TypeName: "outcome",
			Values: []*expr.NamedAttributeExpr{
				{Name: "text", Attribute: &expr.AttributeExpr{Type: expr.String}},
				{Name: "count", Attribute: &expr.AttributeExpr{Type: expr.Int}},
			},
		},
	}
}
