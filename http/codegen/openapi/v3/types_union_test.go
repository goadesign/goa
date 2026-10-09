// This file verifies OpenAPI 3 schema conversion for unions, including that a
// typed owner keeps each discriminator paired with its generated member value.
package openapiv3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

func TestSchemafyCorrelatesUnionDiscriminatorAndValue(t *testing.T) {
	method := &expr.MethodExpr{Name: "union", Service: &expr.ServiceExpr{Name: "test"}}
	generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
		expr.MethodPayloadExampleIdentity(method),
	)
	schema := (&schemafier{rand: generator}).schemafy(unionAttribute())

	require.Len(t, schema.AnyOf, 2)
	assertUnionSchemaBranch(t, schema.AnyOf[0], "text", openapi.Type(openapi.String))
	assertUnionSchemaBranch(t, schema.AnyOf[1], "count", openapi.Type(openapi.Integer))
	assert.Empty(t, schema.Properties)
}

func TestUntaggedUnionSchemaRetainsBranchValidation(t *testing.T) {
	attribute := unionAttribute()
	union := expr.AsUnion(attribute.Type)
	union.Untagged = true
	union.Values[0].Attribute.Validation = &expr.ValidationExpr{Values: []any{"dynamic"}}
	method := &expr.MethodExpr{Name: "union", Service: &expr.ServiceExpr{Name: "test"}}
	generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(expr.MethodPayloadExampleIdentity(method))
	schema := (&schemafier{rand: generator}).schemafy(attribute)
	require.Len(t, schema.AnyOf, 2)
	assert.Empty(t, schema.Type)
	assert.Empty(t, schema.Properties)
	assert.Equal(t, openapi.Type(openapi.String), schema.AnyOf[0].Type)
	assert.Equal(t, []any{"dynamic"}, schema.AnyOf[0].Enum)
	assert.Equal(t, openapi.Type(openapi.Integer), schema.AnyOf[1].Type)
}

func assertUnionSchemaBranch(t *testing.T, branch *openapi.Schema, tag string, valueType openapi.Type) {
	t.Helper()
	assert.Equal(t, openapi.Type(openapi.Object), branch.Type)
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

// TestFlattenedUnionSchemaRetainsObjectValidation checks the advertised object
// branch directly, including the required field inherited from its named type.
func TestFlattenedUnionSchemaRetainsObjectValidation(t *testing.T) {
	method := &expr.MethodExpr{Name: "finish", Service: &expr.ServiceExpr{Name: "operations"}}
	generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("flat")).At(expr.MethodResultExampleIdentity(method))
	complete := &expr.UserTypeExpr{TypeName: "Complete", AttributeExpr: &expr.AttributeExpr{
		Type:       &expr.Object{{Name: "reference", Attribute: &expr.AttributeExpr{Type: expr.String}}},
		Validation: &expr.ValidationExpr{Required: []string{"reference"}},
	}}
	attribute := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "Outcome", TypeKey: "resultType", Flatten: true,
		Values: []*expr.NamedAttributeExpr{{Name: "complete", Attribute: &expr.AttributeExpr{Type: complete}}},
	}}
	schema := (&schemafier{rand: generator}).schemafy(attribute)
	require.Len(t, schema.AnyOf, 1)
	branch := schema.AnyOf[0]
	require.Empty(t, branch.Ref)
	require.Equal(t, []string{"reference", "resultType"}, branch.Required)
	require.Equal(t, []any{"complete"}, branch.Properties["resultType"].Enum)
	require.Equal(t, openapi.Type(openapi.String), branch.Properties["reference"].Type)
	require.NotContains(t, branch.Properties, "value")
	require.Nil(t, complete.Attribute().Find("resultType"))
	require.Equal(t, []string{"reference"}, complete.Attribute().AllRequired())
}
