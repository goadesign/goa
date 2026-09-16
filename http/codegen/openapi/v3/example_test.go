// This file checks how OpenAPI 3 objects receive authored and replacement
// examples without changing the evaluated attribute.
package openapiv3

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

func TestInitExamplesUsesReplacementDescription(t *testing.T) {
	first := &expr.ExampleExpr{Summary: "first", Description: "original", Value: "one"}
	second := &expr.ExampleExpr{Summary: "second", Description: "unchanged", Value: "two"}
	attribute := &expr.AttributeExpr{
		Type:         expr.String,
		UserExamples: []*expr.ExampleExpr{first, second},
	}
	values := (openapi.Values{}).WithDescription(first, "translated")
	media := new(MediaType)
	userType := &expr.UserTypeExpr{AttributeExpr: attribute, TypeName: "Message"}
	generator := expr.NewExampleGenerator(expr.NewDeterministicRandomizerFactory()).At(
		expr.UserTypeExampleIdentity(userType),
	)

	initExamples(media, attribute, generator, values)

	require.Equal(t, "translated", media.Examples["first"].Value.Description)
	require.Equal(t, "unchanged", media.Examples["second"].Value.Description)
	require.Equal(t, "original", first.Description)
}
