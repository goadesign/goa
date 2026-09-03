// This file verifies that OpenAPI text and examples can be replaced for one
// build without changing the evaluated Goa design.
package openapi

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestValues(t *testing.T) {
	api := expr.NewAPIExpr("calc", nil)
	attribute := &expr.AttributeExpr{Type: expr.String}
	examples := []*expr.ExampleExpr{{Summary: "default", Value: "translated"}}

	var empty Values
	require.Equal(t, "original title", empty.Title(api, "original title"))
	require.Equal(t, "original description", empty.Description(api, "original description"))
	require.Equal(t, []*expr.ExampleExpr(nil), empty.Examples(attribute, nil))

	localized := empty.
		WithTitle(api, "localized title").
		WithDescription(api, "localized description").
		WithExamples(attribute, examples)

	require.Equal(t, "localized title", localized.Title(api, "original title"))
	require.Equal(t, "localized description", localized.Description(api, "original description"))
	require.Equal(t, examples, localized.Examples(attribute, nil))
	require.Equal(t, "original title", empty.Title(api, "original title"))

	// Values owns its example list so neither caller nor reader can change it.
	examples[0] = &expr.ExampleExpr{Summary: "changed", Value: "changed"}
	firstRead := localized.Examples(attribute, nil)
	require.Equal(t, "translated", firstRead[0].Value)
	firstRead[0] = &expr.ExampleExpr{Summary: "changed again", Value: "changed again"}
	require.Equal(t, "translated", localized.Examples(attribute, nil)[0].Value)
}

func TestValuesUseAuthoredAttributeForCopies(t *testing.T) {
	authored := &expr.AttributeExpr{Type: expr.String}
	copy := expr.DupAtt(expr.DupAtt(authored))
	localized := (Values{}).WithExamples(authored, []*expr.ExampleExpr{{Value: "translated"}})

	require.Equal(t, "translated", localized.Examples(copy, nil)[0].Value)
}

func TestValuesOwnCompleteExampleLists(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.String}
	examples := []*expr.ExampleExpr{
		{Summary: "first", Value: map[string]any{"items": []any{"one"}}},
		{Summary: "second", Value: "two"},
	}
	values := (Values{}).WithExamples(attribute, examples)

	examples[0].Value.(map[string]any)["items"].([]any)[0] = "changed"
	firstRead := values.Examples(attribute, nil)
	require.Len(t, firstRead, 2)
	require.Equal(t, "one", firstRead[0].Value.(map[string]any)["items"].([]any)[0])
	require.Equal(t, "two", firstRead[1].Value)

	firstRead[0].Value.(map[string]any)["items"].([]any)[0] = "changed again"
	require.Equal(t, "one", values.Examples(attribute, nil)[0].Value.(map[string]any)["items"].([]any)[0])
}

func TestValuesApplyExampleDescriptionsRegardlessOfCallOrder(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.String}
	example := &expr.ExampleExpr{Summary: "default", Description: "original", Value: "value"}

	values := (Values{}).
		WithExamples(attribute, []*expr.ExampleExpr{example}).
		WithDescription(example, "translated")

	require.Equal(t, "translated", values.Examples(attribute, nil)[0].Description)
	require.Equal(t, "original", example.Description)
}

func TestDocsFromExprWithValues(t *testing.T) {
	docs := &expr.DocsExpr{Description: "original", URL: "https://goa.design"}
	values := (Values{}).WithDescription(docs, "localized")

	localized := DocsFromExprWithValues(docs, nil, values)
	require.Equal(t, "localized", localized.Description)
	require.Equal(t, "original", DocsFromExpr(docs, nil).Description)
}
