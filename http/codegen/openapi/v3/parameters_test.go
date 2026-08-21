// This file verifies OpenAPI parameters and headers render one stable example
// in both their schema and their displayed example fields.
package openapiv3

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestParamForAllowEmptyValue(t *testing.T) {
	tests := []struct {
		name     string
		location string
		want     bool
	}{
		{name: "query", location: "query", want: true},
		{name: "path", location: "path", want: false},
		{name: "header", location: "header", want: false},
		{name: "cookie", location: "cookie", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			method := &expr.MethodExpr{Name: "parameter", Service: &expr.ServiceExpr{Name: "test"}}
			generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory(test.name))
			require.NotEmpty(t, generator.At(expr.MethodResultExampleIdentity(method)).String())
			identity := expr.MethodPayloadExampleIdentity(method).Member("value")
			param := paramFor(
				&expr.AttributeExpr{Type: expr.String},
				"value",
				test.location,
				false,
				generator,
				identity,
			)

			require.Equal(t, test.want, param.AllowEmptyValue)
			require.Equal(t, param.Schema.Example, param.Example)
		})
	}
}

func TestHeaderSchemaAndDisplayedExampleShareIdentity(t *testing.T) {
	field := &expr.AttributeExpr{Type: expr.String}
	parent := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "request-id", Attribute: field},
	}}
	headers := expr.NewMappedAttributeExpr(parent)
	method := &expr.MethodExpr{Name: "headers", Service: &expr.ServiceExpr{Name: "test"}}
	generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("headers"))
	require.NotEmpty(t, generator.At(expr.MethodPayloadExampleIdentity(method)).String())

	actual := headersFromAttr(
		headers,
		parent,
		expr.MethodResultExampleIdentity(method),
		generator,
	)["request-id"].Value

	require.Equal(t, actual.Schema.Example, actual.Example)
}
