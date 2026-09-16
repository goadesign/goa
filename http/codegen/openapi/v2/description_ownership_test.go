// This file verifies that shared OpenAPI v2 definitions use the named Goa
// type description instead of text from one response that uses the type.
package openapiv2

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestSharedErrorDefinitionDescription(t *testing.T) {
	cases := []struct {
		name        string
		dsl         func()
		description string
	}{
		{"method order", testdata.SharedErrorDescriptionDSL, "Shared error value"},
		{"reversed method order", testdata.ReversedSharedErrorDescriptionDSL, "Shared error value"},
		{"undescribed type", testdata.UndescribedSharedErrorDSL, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := codegen.RunDSL(t, tc.dsl)
			spec, err := NewV2(
				root,
				root.API.Servers[0].Hosts[0],
			)
			require.NoError(t, err)
			require.Equal(t, tc.description, spec.Definitions["SharedError"].Description)
		})
	}
}

func TestSharedErrorDefinitionLocalizedDescription(t *testing.T) {
	root := codegen.RunDSL(t, testdata.SharedErrorDescriptionDSL)
	sharedError := root.UserType("SharedError")
	values := (openapi.Values{}).WithDescription(sharedError.Attribute(), "Localized shared error")

	spec, err := NewV2WithValues(
		root,
		root.API.Servers[0].Hosts[0],
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		values,
	)
	require.NoError(t, err)
	require.Equal(t, "Localized shared error", spec.Definitions["SharedError"].Description)
}
