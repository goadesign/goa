// This file generates HTTP clients and servers from shared named declarations.
// Only the mapped service changes its JSON property names; complete values and
// the other service's contracts must survive generation and both HTTP directions.
package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/http/codegen/openapi"
)

func TestGenerateHTTPBodyFieldNames(t *testing.T) {
	registry := testRegistry("gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
		testGenerator(planOpenAPIData, testOpenAPIFiles),
	)
	codegen.RunDSL(t, func() {
		dsl.API("fieldnames", func() {
			dsl.Meta("openapi:versions", "2.0", "3.0", "3.2")
		})
		item := dsl.Type("Item", func() {
			dsl.Field(1, "display_name", dsl.String)
			dsl.Field(2, "is_enabled", dsl.Boolean)
			dsl.Field(3, "item_count", dsl.Int, func() { dsl.Default(0) })
			dsl.Field(4, "opaque_text", dsl.String)
			dsl.Field(8, "label", dsl.String, func() {
				dsl.Meta("struct:tag:json", "Label", "omitempty")
			})
			dsl.Field(7, "note_text", dsl.String, func() {
				dsl.Meta("struct:tag:json", "LegacyNote,omitempty")
			})
			dsl.Required("display_name", "is_enabled", "item_count", "opaque_text")
			dsl.Example(dsl.Val{"display_name": "sample", "is_enabled": false, "item_count": 0, "opaque_text": "unchanged"})
		})
		derived := dsl.Type("Derived", func() { dsl.Extend(item) })
		node := dsl.Type("Node", func() {
			dsl.Field(1, "node_name", dsl.String)
			dsl.Field(2, "children", dsl.ArrayOf("Node"))
			dsl.Required("node_name", "children")
		})
		document := dsl.Type("Document", func() {
			dsl.Field(1, "items", dsl.ArrayOf(item))
			dsl.Field(2, "by_key", dsl.MapOf(dsl.String, item))
			dsl.Field(3, "derived", derived)
			dsl.Field(4, "tree", node)
			dsl.OneOf("selection", func() {
				dsl.Field(5, "first", item)
				dsl.Field(6, "second", derived)
			})
			dsl.Required("items", "by_key", "derived", "tree", "selection")
		})
		unrelated := dsl.Type("Unrelated", func() {
			dsl.Field(1, "stable_name", dsl.String)
			dsl.Required("stable_name")
		})
		problem := dsl.Type("Problem", func() {
			dsl.Field(1, "error_code", dsl.String)
			dsl.Required("error_code")
		})
		for _, name := range []string{"mapped", "plain"} {
			dsl.Service(name, func() {
				if name == "mapped" {
					dsl.HTTP(func() {
						dsl.BodyFieldName(item, "display_name:displayName")
						dsl.BodyFieldName(item, "is_enabled:isEnabled")
						dsl.BodyFieldName(item, "item_count:itemCount")
						dsl.BodyFieldName(item, "opaque_text:opaqueText")
						dsl.BodyFieldName(item, "note_text:noteText")
						dsl.BodyFieldName(item, "label:label")
						dsl.BodyFieldName(derived, "display_name:derivedName")
						dsl.BodyFieldName(document, "by_key:byKey")
						dsl.BodyFieldName(node, "node_name:nodeName")
						dsl.BodyFieldName(problem, "error_code:errorCode")
					})
				}
				dsl.Method("unrelated", func() {
					dsl.Payload(unrelated)
					dsl.Result(unrelated)
					dsl.HTTP(func() { dsl.POST("/" + name + "/unrelated") })
				})
				dsl.Method("exchange", func() {
					dsl.Payload(document)
					dsl.Result(document)
					dsl.Error("bad", problem)
					dsl.HTTP(func() {
						dsl.POST("/" + name)
						dsl.Response(200)
						dsl.Response("bad", 400)
					})
					dsl.GRPC(func() {})
				})
			})
		}
	})
	directory := filepath.Join(t.TempDir(), codegen.Gendir)
	writeGeneratedModule(t, directory, "generated.local/gen")
	_, err := generate(filepath.Dir(directory), "gen", false, registry)
	require.NoError(t, err)
	source, err := os.ReadFile(filepath.Join("testdata", "http_body_field_names", "names_test.go"))
	require.NoError(t, err)
	writeGeneratedContractTest(t, directory, "mapped", string(source))
	checkHTTPBodyFieldNameSchemas(t, directory)
	for _, service := range []string{"mapped", "plain"} {
		files, err := filepath.Glob(filepath.Join(directory, "grpc", service, "pb", "*.proto"))
		require.NoError(t, err)
		require.NotEmpty(t, files)
		var definitions []byte
		for _, file := range files {
			content, err := os.ReadFile(file)
			require.NoError(t, err)
			definitions = append(definitions, content...)
		}
		require.Contains(t, string(definitions), "display_name = 1")
		require.NotContains(t, string(definitions), "displayName =")
		require.NotContains(t, string(definitions), "derivedName =")
	}
	runGeneratedTests(t, directory)
}

func checkHTTPBodyFieldNameSchemas(t *testing.T, directory string) {
	t.Helper()
	for _, filename := range []string{"openapi.json", "openapi3.json", "openapi3.2.json"} {
		content, err := os.ReadFile(filepath.Join(directory, "http", filename))
		require.NoError(t, err)
		var document struct {
			Definitions map[string]*openapi.Schema
			Components  struct{ Schemas map[string]*openapi.Schema }
		}
		require.NoError(t, json.Unmarshal(content, &document))
		schemas := document.Components.Schemas
		if filename == "openapi.json" {
			schemas = document.Definitions
		}
		var mapped, plain bool
		for _, schema := range schemas {
			if _, ok := schema.Properties["displayName"]; ok {
				mapped = true
				require.Contains(t, schema.Required, "displayName")
				require.Contains(t, schema.Properties, "noteText")
				require.Equal(t, float64(0), schema.Properties["itemCount"].DefaultValue)
				require.Contains(t, schema.Properties, "label")
				require.NotContains(t, schema.Properties, "Label")
				require.NotContains(t, schema.Properties, "display_name")
				if schema.Example != nil {
					example, ok := schema.Example.(map[string]any)
					require.True(t, ok)
					require.Contains(t, example, "displayName")
					require.NotContains(t, example, "display_name")
				}
			}
			if _, ok := schema.Properties["display_name"]; ok {
				plain = true
			}
		}
		require.True(t, mapped, filename)
		require.True(t, plain, filename)
	}
}
