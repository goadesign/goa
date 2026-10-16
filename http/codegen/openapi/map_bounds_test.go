// This file checks generated map schemas as clients receive them in JSON and
// YAML. Map entry counts, field presence, and value lengths remain separate.
package openapi_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv2 "goa.design/goa/v3/http/codegen/openapi/v2"
	openapiv3 "goa.design/goa/v3/http/codegen/openapi/v3"
)

func TestGeneratedMapBounds(t *testing.T) {
	root := codegen.RunDSL(t, mapBoundsDSL)
	v2, err := openapiv2.NewV2(root, root.API.Servers[0].Hosts[0])
	require.NoError(t, err)
	generator := expr.NewExampleGenerator(root.API.RandomizerFactory)
	standalone := openapiv2.BuildAttributeSchema(
		root.API, &expr.AttributeExpr{Type: root.UserType("LabelRequest")}, generator,
	)
	documents := []struct {
		name     string
		value    any
		location string
	}{
		{"2.0", v2, "#/definitions/LabelRequest"},
		{"3.0", openapiv3.New(root, openapi.Version30), "#/components/schemas/LabelRequest"},
		{"3.2", openapiv3.New(root, openapi.Version32), "#/components/schemas/LabelRequest"},
		{"standalone", standalone, "#/$defs/LabelRequest"},
	}
	formats := []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"json", json.Marshal, json.Unmarshal},
		{"yaml", yaml.Marshal, yaml.Unmarshal},
	}
	for _, document := range documents {
		for _, format := range formats {
			t.Run(document.name+"/"+format.name, func(t *testing.T) {
				data, err := format.marshal(document.value)
				require.NoError(t, err)
				var decoded map[string]any
				require.NoError(t, format.unmarshal(data, &decoded))
				compiler := jsonschema.NewCompiler()
				compiler.DefaultDraft(jsonschema.Draft2020)
				require.NoError(t, compiler.AddResource("https://example.test/schema", decoded))
				request := mapSchemaAt(t, decoded, document.location)
				properties := request["properties"].(map[string]any)

				for _, field := range []string{"inline", "named", "referenced"} {
					t.Run(field, func(t *testing.T) {
						property := properties[field].(map[string]any)
						if ref, ok := property["$ref"].(string); ok {
							property = mapSchemaAt(t, decoded, ref)
						}
						require.NotContains(t, property, "minLength")
						require.NotContains(t, property, "maxLength")
						require.Contains(t, property, "minProperties")
						require.Contains(t, property, "maxProperties")
						schema, err := compiler.Compile("https://example.test/schema" + document.location + "/properties/" + field)
						require.NoError(t, err)
						for count := range 4 {
							value := make(map[string]any, count)
							for index := range count {
								value[fmt.Sprintf("key%d", index)] = "ok"
							}
							err := schema.Validate(value)
							if count == 1 || count == 2 {
								require.NoError(t, err, "map with %d entries", count)
							} else {
								require.Error(t, err, "map with %d entries", count)
							}
						}
						require.Error(t, schema.Validate(map[string]any{"key": "x"}))
						require.Error(t, schema.Validate(map[string]any{"key": "large"}))
					})
				}

				cases := []struct {
					name  string
					field string
					value any
					valid bool
				}{
					{
						name: "zero accepts empty", field: "zero",
						value: map[string]any{}, valid: true,
					},
					{
						name: "zero rejects one", field: "zero",
						value: map[string]any{"a": "ok"},
					},
					{
						name: "unbounded accepts empty", field: "unbounded",
						value: map[string]any{}, valid: true,
					},
					{
						name: "unbounded accepts three", field: "unbounded",
						value: map[string]any{"a": "ok", "b": "ok", "c": "ok"}, valid: true,
					},
					{
						name: "nested counts each map", field: "nested",
						value: map[string]any{
							"a": map[string]any{"a": "ok", "b": "ok"},
							"b": map[string]any{"a": "ok", "b": "ok"},
						},
						valid: true,
					},
					{
						name: "nested rejects empty child", field: "nested",
						value: map[string]any{"a": map[string]any{}},
					},
					{
						name: "nested rejects large child", field: "nested",
						value: map[string]any{"a": map[string]any{"a": "ok", "b": "ok", "c": "ok"}},
					},
					{
						name: "string accepts length two", field: "text",
						value: "ok", valid: true,
					},
					{
						name: "string rejects length one", field: "text", value: "x",
					},
					{
						name: "string rejects length five", field: "text", value: "large",
					},
					{
						name: "array accepts two", field: "list",
						value: []any{"a", "b"}, valid: true,
					},
					{
						name: "array rejects empty", field: "list", value: []any{},
					},
					{
						name: "array rejects three", field: "list", value: []any{"a", "b", "c"},
					},
				}
				for _, test := range cases {
					t.Run(test.name, func(t *testing.T) {
						schema, err := compiler.Compile("https://example.test/schema" + document.location + "/properties/" + test.field)
						require.NoError(t, err)
						err = schema.Validate(test.value)
						if test.valid {
							require.NoError(t, err)
						} else {
							require.Error(t, err)
						}
					})
				}
				require.NotContains(t, properties["unbounded"], "minProperties")
				require.NotContains(t, properties["unbounded"], "maxProperties")
				require.NotContains(t, properties["text"], "minProperties")
				require.NotContains(t, properties["list"], "minProperties")
				schema, err := compiler.Compile("https://example.test/schema" + document.location)
				require.NoError(t, err)
				require.Error(t, schema.Validate(map[string]any{}), "required field must be present")
				require.NoError(t, schema.Validate(map[string]any{"required_empty": map[string]any{}}), "required map may be empty and optional bounded maps may be absent")
			})
		}
	}
}

// mapBoundsDSL gives document and standalone builders the same map constraints.
// Each nested map counts its own entries, regardless of the total in the request.
func mapBoundsDSL() {
	dsl.API("labels", func() {
		dsl.Title("Labels API")
	})
	labels := dsl.Type("Labels", dsl.MapOf(dsl.String, dsl.String), func() {
		dsl.MinLength(1)
		dsl.MaxLength(2)
		dsl.Elem(func() {
			dsl.MinLength(2)
			dsl.MaxLength(4)
		})
	})
	reference := dsl.Type("LabelFields", func() {
		dsl.Attribute("referenced", labels)
	})
	request := dsl.Type("LabelRequest", func() {
		dsl.Reference(reference)
		dsl.Attribute("referenced")
		dsl.Attribute("named", labels)
		dsl.Attribute("inline", dsl.MapOf(dsl.String, dsl.String), func() {
			dsl.MinLength(1)
			dsl.MaxLength(2)
			dsl.Elem(func() {
				dsl.MinLength(2)
				dsl.MaxLength(4)
			})
		})
		dsl.Attribute("nested", dsl.MapOf(dsl.String, labels), func() {
			dsl.MinLength(1)
			dsl.MaxLength(2)
		})
		dsl.Attribute("zero", dsl.MapOf(dsl.String, dsl.String), func() {
			dsl.MinLength(0)
			dsl.MaxLength(0)
		})
		dsl.Attribute("unbounded", dsl.MapOf(dsl.String, dsl.String))
		dsl.Attribute("required_empty", dsl.MapOf(dsl.String, dsl.String))
		dsl.Attribute("text", dsl.String, func() {
			dsl.MinLength(2)
			dsl.MaxLength(4)
		})
		dsl.Attribute("list", dsl.ArrayOf(dsl.String), func() {
			dsl.MinLength(1)
			dsl.MaxLength(2)
		})
		dsl.Required("required_empty")
	})
	dsl.Service("labels", func() {
		dsl.Method("set", func() {
			dsl.Payload(request)
			dsl.HTTP(func() {
				dsl.POST("/labels")
			})
		})
	})
}

// mapSchemaAt reads a local schema reference from a decoded document so tests
// can check the exact keywords emitted for inline and named map definitions.
func mapSchemaAt(t *testing.T, document map[string]any, reference string) map[string]any {
	t.Helper()
	var value any = document
	for _, name := range strings.Split(strings.TrimPrefix(reference, "#/"), "/") {
		object, ok := value.(map[string]any)
		require.True(t, ok, "object at %s", name)
		value, ok = object[name]
		require.True(t, ok, "schema path %s", reference)
	}
	object, ok := value.(map[string]any)
	require.True(t, ok, "schema at %s", reference)
	return object
}
