// This file generates typed form and JSON clients, servers, and OpenAPI together.
// The generated module sends real HTTP requests and checks body validation before
// endpoint invocation, including mapped names and optional body construction.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

func TestGenerateHTTPForms(t *testing.T) {
	registry := testRegistry("gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
		testGenerator(planOpenAPIData, testOpenAPIFiles),
	)
	codegen.RunDSL(t, formRequestDSL)
	directory := filepath.Join(t.TempDir(), codegen.Gendir)
	writeGeneratedModule(t, directory, "generated.local/gen")
	_, err := generate(filepath.Dir(directory), "gen", false, registry)
	require.NoError(t, err)
	source, err := os.ReadFile(filepath.Join("testdata", "http_forms", "forms_test.go"))
	require.NoError(t, err)
	writeGeneratedContractTest(t, directory, "formtests", string(source))
	runGeneratedTests(t, directory)
}

// formRequestDSL combines flat form bodies, selected optional bodies, renamed
// transport fields, ordinary JSON and a sibling JSON-RPC service in one design.
func formRequestDSL() {
	d.API("forms", func() {
		d.Meta("openapi:versions", "2.0", "3.0", "3.2")
	})
	identifier := d.Type("Identifier", d.String, func() {
		d.MinLength(1)
		d.Meta("struct:pkg:path", "shared/types")
	})
	numbers := d.Type("Numbers", d.ArrayOfRequired(d.Int), func() {
		d.MinLength(1)
	})
	details := d.Type("Details", func() {
		d.Attribute("message", d.String)
		d.Attribute("attempts", d.Int, func() {
			d.Default(3)
			d.Minimum(0)
		})
		d.Required("message")
	})
	d.Service("forms", func() {
		d.Method("Exchange", func() {
			d.Payload(func() {
				d.Attribute("name", identifier)
				d.Attribute("empty", d.String)
				d.Attribute("flag", d.Boolean)
				d.Attribute("count", d.Int, func() {
					d.Default(3)
					d.Minimum(0)
				})
				d.Attribute("small", d.Int32)
				d.Attribute("large", d.Int64)
				d.Attribute("unsigned", d.UInt)
				d.Attribute("unsigned32", d.UInt32)
				d.Attribute("unsigned64", d.UInt64)
				d.Attribute("fraction32", d.Float32)
				d.Attribute("fraction64", d.Float64)
				d.Attribute("data", d.Bytes)
				d.Attribute("numbers", numbers)
				d.Attribute("labels", d.ArrayOf(d.String))
				d.Attribute("blobs", d.ArrayOf(d.Bytes))
				d.Attribute("site", d.String)
				d.Attribute("page", d.Int)
				d.Attribute("token", d.String)
				d.Attribute("cookie", d.String)
				d.Required("name", "empty", "numbers", "site", "page")
			})
			d.Result(d.String)
			d.HTTP(func() {
				d.POST("/forms/{site}")
				d.Param("page:page_number")
				d.Header("token:X-Token")
				d.Cookie("cookie:session")
				d.FormRequest()
				d.Response(200)
			})
		})
		d.Method("Optional", func() {
			d.Payload(func() {
				d.Attribute("body", details)
				d.Attribute("token", d.String)
			})
			d.Result(d.String)
			d.HTTP(func() {
				d.POST("/optional")
				d.Header("token:X-Token")
				d.Body("body")
				d.FormRequest()
				d.Response(200)
			})
		})
		d.Method("Mapped", func() {
			d.Payload(func() {
				d.Attribute("clientId", identifier, func() {
					d.Meta("struct:field:name", "ClientID")
				})
				d.Required("clientId")
			})
			d.Result(d.String)
			d.HTTP(func() {
				d.POST("/mapped")
				d.Body(func() {
					d.Attribute("clientId:client_id")
					d.Required("clientId")
				})
				d.FormRequest()
				d.Response(200)
			})
		})
		for _, transport := range []struct {
			name string
			form bool
		}{
			{"MappedFlat", true}, {"MappedFlatJSON", false},
		} {
			d.Method(transport.name, func() {
				d.Payload(func() {
					d.Attribute("label", d.String)
					d.Attribute("tags", d.ArrayOf(d.String))
					d.Required("label", "tags")
				})
				d.HTTP(func() {
					d.POST("/" + transport.name)
					d.Body(func() {
						d.Attribute("label:display_name", d.String)
						d.Attribute("tags:tag", d.ArrayOf(d.String))
						d.Required("label", "tags")
					})
					if transport.form {
						d.FormRequest()
					}
					d.Response(204)
				})
			})
		}
		d.Method("OptionalJSON", func() {
			d.Payload(func() {
				d.Attribute("body", details)
			})
			d.Result(d.String)
			d.HTTP(func() {
				d.POST("/optional-json")
				d.Body("body")
				d.Response(200)
			})
		})
		d.Method("JSON", func() {
			d.Payload(details)
			d.Result(d.String)
			d.HTTP(func() {
				d.POST("/json")
				d.Response(200)
			})
		})
	})
	d.Service("rpcforms", func() {
		d.JSONRPC(func() {
			d.POST("/rpc")
		})
		d.Method("read", func() {
			d.Payload(details)
			d.Result(d.String)
			d.JSONRPC(func() {})
		})
	})
}

// TestGenerateHTTPFormExamples compiles generated application startup and CLI
// code so selecting forms never requires a handwritten encoder or decoder.
func TestGenerateHTTPFormExamples(t *testing.T) {
	root := codegen.RunDSL(t, formRequestDSL)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planExampleData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transport, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transport...)
	examples, err := assembleExampleFilesForTest(plan)
	require.NoError(t, err)
	files = append(files, examples...)
	files, err = mergeFilesByPath(files)
	require.NoError(t, err)
	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	runGeneratedTests(t, directory)
}
