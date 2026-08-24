// This file checks that each Swagger union choice pairs its name with the
// schema for the matching value.
package openapiv2

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

func TestAttributeTypeSchemaCorrelatesUnionDiscriminatorAndValue(t *testing.T) {
	method := &expr.MethodExpr{Name: "union", Service: &expr.ServiceExpr{Name: "test"}}
	generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test")).At(
		expr.MethodPayloadExampleIdentity(method),
	)
	schema := newSchemaBuilder(openapi.Values{}).attributeTypeSchemaWithPrefix(
		&expr.APIExpr{},
		unionAttribute(),
		"",
		generator,
	)

	require.Len(t, schema.AnyOf, 2)
	assertUnionSchemaBranch(t, schema.AnyOf[0], "text", openapi.String)
	assertUnionSchemaBranch(t, schema.AnyOf[1], "count", openapi.Integer)
	assert.Empty(t, schema.Properties)
}

func TestBuildAttributeSchemaKeepsDefinitionsSeparate(t *testing.T) {
	type result struct {
		field  string
		schema *openapi.Schema
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	build := func(field string) {
		ready.Done()
		<-start
		generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory(field))
		results <- result{
			field: field,
			schema: BuildAttributeSchema(
				&expr.APIExpr{},
				namedObjectAttribute(field),
				generator,
			),
		}
	}
	go build("first")
	go build("second")
	ready.Wait()
	close(start)

	for range 2 {
		built := <-results
		require.Equal(t, "#/$defs/Shared", built.schema.Ref)
		require.Contains(t, built.schema.Defs, "Shared")
		require.Contains(t, built.schema.Defs["Shared"].Properties, built.field)
		other := "first"
		if built.field == "first" {
			other = "second"
		}
		require.NotContains(t, built.schema.Defs["Shared"].Properties, other)
		_, err := json.Marshal(built.schema)
		require.NoError(t, err)
	}
}

// assertUnionSchemaBranch checks one generated union branch's tag, required
// fields, and value type.
func assertUnionSchemaBranch(t *testing.T, branch *openapi.Schema, tag string, valueType openapi.Type) {
	t.Helper()
	assert.Equal(t, openapi.Type(openapi.Object), branch.Type)
	assert.Equal(t, []string{"type", "value"}, branch.Required)
	require.Contains(t, branch.Properties, "type")
	assert.Equal(t, []any{tag}, branch.Properties["type"].Enum)
	require.Contains(t, branch.Properties, "value")
	assert.Equal(t, valueType, branch.Properties["value"].Type)
}

// unionAttribute returns the string-or-integer union used by these schema tests.
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

// namedObjectAttribute returns a named object with one field.
func namedObjectAttribute(field string) *expr.AttributeExpr {
	object := expr.Object{
		&expr.NamedAttributeExpr{
			Name:      field,
			Attribute: &expr.AttributeExpr{Type: expr.String},
		},
	}
	return &expr.AttributeExpr{
		Type: &expr.UserTypeExpr{
			AttributeExpr: &expr.AttributeExpr{Type: &object},
			TypeName:      "Shared",
		},
	}
}
