package openapi_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	v2 "goa.design/goa/v3/http/codegen/openapi/v2"
	v3 "goa.design/goa/v3/http/codegen/openapi/v3"
)

func TestInlineUnionBranchLocalizedValues(t *testing.T) {
	var envelope expr.UserType
	root := codegen.RunDSL(t, func() {
		envelope = d.Type("Envelope", func() {
			d.OneOf("choice", func() {
				d.Attribute("detail", func() {
					d.Attribute("label", d.String)
					d.Required("label")
				})
			})
			d.Required("choice")
		})
		d.Service("details", func() {
			d.Method("read", func() {
				d.Result(envelope)
				d.HTTP(func() { d.GET("/details"); d.Response(d.StatusOK) })
			})
		})
	})
	branch := expr.AsUnion(envelope.Attribute().Find("choice").Type).Values[0].Attribute
	const description = "Localized detail"
	example := map[string]any{"label": "translated"}
	values := (openapi.Values{}).
		WithDescription(branch, description).
		WithExamples(branch, []*expr.ExampleExpr{{Value: example}})

	spec2, err := v2.NewV2WithValues(root, root.API.Servers[0].Hosts[0], expr.NewExampleGenerator(root.API.RandomizerFactory), values)
	require.NoError(t, err)
	found := false
	for _, schema := range spec2.Definitions {
		if schema.Properties["label"] != nil {
			found = true
			require.Equal(t, description, schema.Description)
			require.Equal(t, example, schema.Example)
		}
	}
	require.True(t, found, "missing inline object definition in OpenAPI v2")

	for _, version := range []openapi.Version{openapi.Version30, openapi.Version32} {
		spec3 := v3.NewWithValues(root, version, expr.NewExampleGenerator(root.API.RandomizerFactory), values)
		found = false
		for _, schema := range spec3.Components.Schemas {
			if schema.Properties["label"] != nil {
				found = true
				require.Equal(t, description, schema.Description)
				require.Equal(t, example, schema.Example)
			}
		}
		require.True(t, found, "missing inline object component in OpenAPI %s", version)
	}
}
