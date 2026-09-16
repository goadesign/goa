// This file verifies that shared OpenAPI v3 components use the named Goa type
// description while each response keeps its own error description.
package openapiv3

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestSharedErrorComponentDescription(t *testing.T) {
	versions := []struct {
		name    string
		version openapi.Version
	}{
		{"3.0", openapi.Version30},
		{"3.2", openapi.Version32},
	}
	designs := []struct {
		name        string
		dsl         func()
		description string
	}{
		{"method order", testdata.SharedErrorDescriptionDSL, "Shared error value"},
		{"reversed method order", testdata.ReversedSharedErrorDescriptionDSL, "Shared error value"},
		{"undescribed type", testdata.UndescribedSharedErrorDSL, ""},
	}
	for _, version := range versions {
		t.Run(version.name, func(t *testing.T) {
			for _, design := range designs {
				t.Run(design.name, func(t *testing.T) {
					root := codegen.RunDSL(t, design.dsl)
					spec := New(
						root,
						version.version,
					)
					require.Equal(t, design.description, spec.Components.Schemas["SharedError"].Description)
				})
			}
		})
	}
}

func TestSharedErrorResponseDescriptions(t *testing.T) {
	for _, version := range []openapi.Version{openapi.Version30, openapi.Version32} {
		root := codegen.RunDSL(t, testdata.SharedErrorDescriptionDSL)
		spec := New(root, version)

		first := spec.Paths["/first"].Get.Responses["400"].Value.Description
		second := spec.Paths["/second"].Get.Responses["400"].Value.Description
		require.NotNil(t, first)
		require.NotNil(t, second)
		require.Equal(t, "first_error: First failure", *first)
		require.Equal(t, "second_error: Second failure", *second)
	}
}

func TestSharedErrorComponentLocalizedDescription(t *testing.T) {
	for _, version := range []openapi.Version{openapi.Version30, openapi.Version32} {
		root := codegen.RunDSL(t, testdata.SharedErrorDescriptionDSL)
		sharedError := root.UserType("SharedError")
		values := (openapi.Values{}).WithDescription(sharedError.Attribute(), "Localized shared error")
		spec := NewWithValues(
			root,
			version,
			expr.NewExampleGenerator(root.API.RandomizerFactory),
			values,
		)

		require.Equal(t, "Localized shared error", spec.Components.Schemas["SharedError"].Description)
	}
}

func TestUndescribedSharedErrorIgnoresLocalizedResponseDescription(t *testing.T) {
	for _, version := range []openapi.Version{openapi.Version30, openapi.Version32} {
		root := codegen.RunDSL(t, testdata.UndescribedSharedErrorDSL)
		firstError := root.Service("errors").Method("first").Error("first_error")
		values := (openapi.Values{}).
			WithDescription(firstError, "Localized first failure").
			WithDescription(firstError.AttributeExpr, "Localized first failure")
		spec := NewWithValues(
			root,
			version,
			expr.NewExampleGenerator(root.API.RandomizerFactory),
			values,
		)

		require.Empty(t, spec.Components.Schemas["SharedError"].Description)
		response := spec.Paths["/first"].Get.Responses["400"].Value.Description
		require.NotNil(t, response)
		require.Equal(t, "first_error: Localized first failure", *response)
	}
}
