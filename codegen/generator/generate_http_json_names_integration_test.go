// This file generates clients, servers, and OpenAPI from types with internal
// JSON names. The generated tests exchange HTTP bodies and separately marshal
// service values so both uses must keep their own field names.
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

func TestGenerateHTTPNonTransportJSONNames(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
		testGenerator(planOpenAPIData, testOpenAPIFiles),
	)
	codegen.RunDSL(t, func() {
		dsl.API("names", func() {
			dsl.Meta("openapi:versions", "2.0", "3.0", "3.2")
		})
		detail := dsl.Type("Detail", func() {
			dsl.Attribute("unitName", dsl.String, func() {
				dsl.Meta("struct:tag:json:name", "unit_name")
			})
			dsl.Required("unitName")
		})
		entry := dsl.Type("Entry", func() {
			dsl.Attribute("labelText", dsl.String, func() {
				dsl.Meta("struct:tag:json:name", "label_text")
			})
			dsl.Attribute("detail", detail, func() {
				dsl.Meta("struct:tag:json:name", "stored_detail")
			})
			dsl.Required("labelText", "detail")
		})
		document := dsl.Type("Document", func() {
			dsl.Attribute("displayName", dsl.String, func() {
				dsl.Meta("struct:tag:json:name", "display_name")
			})
			dsl.Attribute("noteText", dsl.String, func() {
				dsl.Meta("struct:tag:json:name", "note_text")
			})
			dsl.Attribute("entries", dsl.ArrayOf(entry), func() {
				dsl.Meta("struct:tag:json:name", "stored_entries")
			})
			dsl.Required("displayName", "entries")
		})
		externalName := dsl.Type("ExternalName", func() {
			dsl.Attribute("public_name", dsl.String, func() {
				dsl.Meta("struct:field:name", "DisplayName")
				dsl.Meta("struct:tag:json:name", "stored_name")
			})
			dsl.Required("public_name")
		})
		dsl.Service("names", func() {
			dsl.Method("exchange", func() {
				dsl.Payload(document)
				dsl.Result(document)
				dsl.HTTP(func() {
					dsl.POST("/documents")
					dsl.Response(200)
				})
			})
			dsl.Method("mapped", func() {
				dsl.Payload(externalName)
				dsl.Result(externalName)
				dsl.HTTP(func() {
					dsl.POST("/mapped")
					dsl.Response(200)
				})
			})
		})
		dsl.Service("rpcnames", func() {
			dsl.JSONRPC(func() {
				dsl.POST("/rpc")
			})
			dsl.Method("exchange", func() {
				dsl.Payload(document)
				dsl.Result(document)
				dsl.JSONRPC(func() {})
			})
		})
	})

	directory := filepath.Join(t.TempDir(), codegen.Gendir)
	writeGeneratedModule(t, directory, "generated.local/gen")
	_, err := generate(filepath.Dir(directory), "gen", false, registry)
	require.NoError(t, err)
	source, err := os.ReadFile(filepath.Join("testdata", "http_json_names", "names_test.go"))
	require.NoError(t, err)
	writeGeneratedContractTest(t, directory, "names", string(source))
	checkJSONNameSchemas(t, directory)
	runGeneratedTests(t, directory)
}

// checkJSONNameSchemas reads each generated specification and checks the
// nested types used by the request and response tests against the HTTP names.
func checkJSONNameSchemas(t *testing.T, directory string) {
	t.Helper()
	for _, filename := range []string{"openapi.json", "openapi3.json", "openapi3.2.json"} {
		t.Run(filename, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(directory, "http", filename))
			require.NoError(t, err)
			var document struct {
				Definitions map[string]*openapi.Schema
				Components  struct {
					Schemas map[string]*openapi.Schema
				}
			}
			require.NoError(t, json.Unmarshal(content, &document))
			schemas := document.Components.Schemas
			if filename == "openapi.json" {
				schemas = document.Definitions
			}
			for _, test := range []struct {
				name     string
				property string
				internal string
			}{
				{"Document", "displayName", "display_name"},
				{"Entry", "labelText", "label_text"},
				{"Detail", "unitName", "unit_name"},
			} {
				require.Contains(t, schemas, test.name)
				require.Contains(t, schemas[test.name].Properties, test.property)
				require.Contains(t, schemas[test.name].Required, test.property)
				require.NotContains(t, schemas[test.name].Properties, test.internal)
			}
		})
	}
}
