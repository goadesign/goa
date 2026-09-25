// These tests render complete HTTP clients and servers from named response
// types. They check imports before file finalization and compile the generated
// packages, so unused-import removal cannot hide an incorrect import plan.
package codegen

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestHTTPResponseValidationImportsUsePreparedBodies checks each response's
// planned imports and compiles its generated service, client and server.
func TestHTTPResponseValidationImportsUsePreparedBodies(t *testing.T) {
	tests := []struct {
		name          string
		stream        bool
		inline        bool
		typeValidator bool
	}{
		{name: "ordinary_named_string", inline: true},
		{name: "selected_named_collection", inline: true},
		{name: "error_named_string", inline: true},
		{name: "sse_named_string", stream: true, inline: true},
		{name: "mixed_sse_named_string", stream: true, inline: true},
		{name: "nested_named_collections", inline: true},
		{name: "named_object", typeValidator: true},
		{name: "sse_named_object", stream: true, typeValidator: true},
		{name: "array_of_named_objects", typeValidator: true},
		{name: "empty_mixed_sse", stream: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := namedResponseBodyRoot(t, test.name)
			plan := linkedHTTPPlanForRoot(t, root)
			suffix := "/client/encode_decode.go"
			if test.stream {
				suffix = "/client/sse.go"
			}
			t.Run("imports", func(t *testing.T) {
				imports := plannedHTTPFileImports(t, plan.ClientFiles(), suffix)
				require.Equal(t, test.inline, slices.Contains(imports, "unicode/utf8"))
				if test.inline || test.name == "array_of_named_objects" {
					require.Contains(t, imports, codegen.GoaImport("").Path)
				}
				if test.typeValidator {
					// The decoded object's validator owns its field checks,
					// even when an enclosing array is checked by the codec.
					imports := plannedHTTPFileImports(t, plan.ClientTypeFiles(), "/client/types.go")
					require.Contains(t, imports, "unicode/utf8")
				}
			})

			// Compile each design independently so another endpoint cannot
			// accidentally supply an import missing from this response.
			serviceFiles, err := service.Files(plan.servicePlan)
			require.NoError(t, err)
			files := slices.Clone(serviceFiles)
			files = append(files, plan.ClientFiles()...)
			files = append(files, plan.ClientTypeFiles()...)
			files = append(files, plan.ServerFiles()...)
			files = append(files, plan.ServerTypeFiles()...)
			files = append(files, plan.PathFiles()...)
			runGeneratedMixedSSECompile(t, files, "./gen/...")
		})
	}
}

// namedResponseBodyRoot isolates one response shape with a Unicode length rule.
// Ordinary, error and streaming responses must all use the HTTP body's shape
// when deciding which generated file performs the check.
func namedResponseBodyRoot(t *testing.T, kind string) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		label := dsl.Type("Label", dsl.String, func() {
			dsl.MinLength(2)
		})
		labels := dsl.Type("Labels", dsl.ArrayOf(label))
		groups := dsl.Type("Groups", dsl.MapOf(dsl.String, labels))
		item := dsl.Type("Item", func() {
			dsl.Attribute("label", dsl.String, func() {
				dsl.MinLength(2)
			})
			dsl.Required("label")
		})
		dsl.Service("Response Shapes", func() {
			dsl.Method("Read", func() {
				switch kind {
				case "ordinary_named_string":
					dsl.Result(label)
				case "selected_named_collection":
					dsl.Result(func() {
						dsl.Attribute("labels", labels)
						dsl.Required("labels")
					})
				case "error_named_string":
					dsl.Error("bad", label)
				case "sse_named_string":
					dsl.StreamingResult(label)
				case "mixed_sse_named_string":
					dsl.Result(dsl.String)
					dsl.StreamingResult(label)
				case "nested_named_collections":
					dsl.Result(groups)
				case "named_object":
					dsl.Result(item)
				case "sse_named_object":
					dsl.StreamingResult(item)
				case "array_of_named_objects":
					dsl.Result(dsl.ArrayOf(item))
				case "empty_mixed_sse":
					dsl.Result(dsl.String, func() {
						dsl.MinLength(2)
					})
					dsl.StreamingResult(dsl.Empty)
				default:
					t.Fatalf("unknown response fixture %q", kind)
				}
				dsl.HTTP(func() {
					dsl.GET("/read")
					switch kind {
					case "selected_named_collection":
						dsl.Response(dsl.StatusOK, func() {
							dsl.Body("labels")
						})
					case "error_named_string":
						dsl.Response(dsl.StatusNoContent)
						dsl.Response("bad", dsl.StatusBadRequest)
					case "sse_named_string", "mixed_sse_named_string", "sse_named_object", "empty_mixed_sse":
						dsl.ServerSentEvents()
					}
				})
			})
		})
	})
}
