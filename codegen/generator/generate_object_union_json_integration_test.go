// This file generates object unions through the native service, HTTP,
// JSON-RPC, gRPC and OpenAPI plans. The compiled callers verify the new JSON
// mapping and the existing tagged mapping in the same generated application.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
)

func TestGenerateObjectUnionJSONMapping(t *testing.T) {
	registry := testRegistry("gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
		testGenerator(planOpenAPIData, testOpenAPIFiles),
	)
	codegen.RunDSL(t, objectUnionJSONDSL)
	directory := filepath.Join(t.TempDir(), codegen.Gendir)
	writeGeneratedModule(t, directory, "generated.local/gen")
	_, err := generate(filepath.Dir(directory), "gen", false, registry)
	require.NoError(t, err)
	source, err := os.ReadFile(filepath.Join("testdata", "object_union_json", "union_test.go"))
	require.NoError(t, err)
	writeGeneratedContractTest(t, directory, "unions", string(source))
	runGeneratedTests(t, directory)
}

// objectUnionJSONDSL uses the same typed result in an ordinary HTTP body,
// a selected HTTP body, a JSON-RPC result and protobuf messages.
func objectUnionJSONDSL() {
	d.API("object-unions", func() { d.Meta("openapi:versions", "2.0", "3.0", "3.2") })
	complete := d.Type("Complete", func() {
		d.Field(1, "reference", d.String, "Completed operation reference", func() {
			d.MinLength(1)
			d.Meta("struct:tag:json:name", "stored_reference")
		})
		d.Required("reference")
	})
	pending := d.Type("Pending", func() {
		d.Field(1, "message", d.String, "Input requested before completion", func() { d.MinLength(1) })
		d.Field(2, "state", d.String, "Operation state returned to the caller", func() { d.MinLength(1) })
		d.Required("message", "state")
	})
	outcome := d.Type("Outcome", func() {
		d.OneOf("outcome", "Exactly one operation result", func() {
			d.TypeName("ResultChoice")
			d.Meta("oneof:json:flatten")
			d.Meta("oneof:type:field", "resultType")
			d.Field(1, "complete", complete, "Completed operation")
			d.Field(2, "input_required", pending, "Input is required")
		})
		d.Required("outcome")
	})
	collection := d.Type("Collection", func() {
		d.Field(1, "items", d.ArrayOf(outcome), "Operation results in order")
		d.Field(2, "named", d.MapOf(d.String, outcome), "Results by operation name")
		d.Required("items", "named")
	})
	tagged := d.Type("Tagged", func() {
		d.OneOf("choice", func() {
			d.TypeName("TaggedChoice")
			d.Meta("oneof:type:field", "kind")
			d.Meta("oneof:value:field", "data")
			d.Field(1, "text", d.String, "Text result")
			d.Field(2, "complete", complete, "Object result")
		})
		d.Required("choice")
	})
	d.Service("unions", func() {
		for _, method := range []struct {
			name     string
			value    any
			selected bool
		}{
			{"echo", outcome, false}, {"selected", outcome, true}, {"nested", collection, false}, {"tagged", tagged, false}, {"tagged_selected", tagged, true},
		} {
			d.Method(method.name, func() {
				d.Payload(method.value)
				d.Result(method.value)
				d.HTTP(func() {
					d.POST("/" + method.name)
					if method.selected {
						field := "outcome"
						if method.name == "tagged_selected" {
							field = "choice"
						}
						d.Body(field)
					}
					d.Response(d.StatusOK, func() {
						if method.selected {
							field := "outcome"
							if method.name == "tagged_selected" {
								field = "choice"
							}
							d.Body(field)
						}
					})
				})
				d.GRPC(func() {})
			})
		}
	})
	d.Service("rpcunions", func() {
		d.JSONRPC(func() { d.POST("/rpc") })
		d.Method("echo", func() {
			d.Payload(outcome)
			d.Result(outcome)
			d.JSONRPC(func() {
				d.Body("outcome")
				d.Response(d.StatusOK, func() { d.Body("outcome") })
			})
		})
	})
}
