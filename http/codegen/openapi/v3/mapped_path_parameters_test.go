// This file verifies that mapped HTTP names select parameter locations while
// payload fields keep their declared names, validation rules, and examples.
package openapiv3

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	dsl "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/http/codegen/openapi"
)

// TestMappedHTTPParameterLocations checks documents from payload fields with
// different HTTP names, including unchanged query and path controls.
func TestMappedHTTPParameterLocations(t *testing.T) {
	cases := []struct {
		name            string
		httpName        string
		path            string
		pathMapping     string
		location        string
		payloadRequired bool
	}{
		{name: "mapped path", httpName: "widgetName", path: "/widgets/{widgetName}", location: "path"},
		{name: "unmapped path", httpName: "name", path: "/widgets/{name}", location: "path"},
		{name: "mapped optional query", httpName: "widgetName", path: "/widgets", location: "query"},
		{name: "mapped required query", httpName: "widgetName", path: "/widgets", location: "query", payloadRequired: true},
		{name: "unmapped query", httpName: "name", path: "/widgets", location: "query"},
		{name: "query payload name matches another path token", httpName: "search", path: "/widgets/{name}", pathMapping: "key:name", location: "query"},
	}
	versions := []struct {
		name    string
		version openapi.Version
	}{
		{name: "3.0", version: openapi.Version30},
		{name: "3.2", version: openapi.Version32},
	}
	for _, version := range versions {
		for _, tc := range cases {
			t.Run(version.name+"/"+tc.name, func(t *testing.T) {
				root := codegen.RunDSL(t, func() {
					dsl.API("widgets", func() {
						dsl.Title("Widget API")
					})
					identity := dsl.Type("WidgetIdentity", func() {
						dsl.Attribute("name", dsl.String, "Widget name", func() {
							dsl.MinLength(1)
							dsl.Example("blue-widget")
						})
					})
					dsl.Service("widgets", func() {
						dsl.Method("show", func() {
							dsl.Payload(func() {
								dsl.Reference(identity)
								dsl.Attribute("name")
								if tc.pathMapping != "" {
									dsl.Attribute("key", dsl.String)
								}
								if tc.payloadRequired {
									dsl.Required("name")
								}
							})
							dsl.HTTP(func() {
								dsl.GET(tc.path)
								dsl.Param("name:" + tc.httpName)
								if tc.pathMapping != "" {
									dsl.Param(tc.pathMapping)
								}
							})
						})
					})
				})
				spec := New(root, version.version)
				operation := spec.Paths[tc.path].Get

				var selected *Parameter
				for _, parameter := range operation.Parameters {
					if parameter.Value.Name == tc.httpName {
						selected = parameter.Value
						break
					}
				}
				require.NotNil(t, selected)
				require.Equal(t, tc.location, selected.In)
				require.Equal(t, tc.location == "path" || tc.payloadRequired, selected.Required)
				require.NotNil(t, root.Service("widgets").Method("show").Payload.Find("name"))
				require.Equal(t, openapi.Type(openapi.String), selected.Schema.Type)
				require.NotNil(t, selected.Schema.MinLength)
				require.Equal(t, 1, *selected.Schema.MinLength)
				require.Equal(t, "blue-widget", selected.Example)
				require.Equal(t, selected.Example, selected.Schema.Example)
				require.Equal(t, tc.location == "query", selected.AllowEmptyValue)
			})
		}
	}
}
