// This file verifies that HTTP clients keep Goa's built-in service error type
// after the transport generator copies method error expressions.
package generator

import (
	"path/filepath"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

func TestGenerateHTTPErrorResultAndCustomError(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
	)
	codegen.RunDSL(t, func() {
		custom := dsl.Type("CustomError", func() {
			dsl.ErrorName("name", dsl.String)
			dsl.Attribute("message", dsl.String)
			dsl.Required("name", "message")
		})
		dsl.Service("Records", func() {
			dsl.Method("Read", func() {
				dsl.Error("not_found")
				dsl.Error("rejected", custom)
				dsl.HTTP(func() {
					dsl.GET("/records")
					dsl.Response("not_found", dsl.StatusNotFound)
					dsl.Response("rejected", dsl.StatusBadRequest)
				})
			})
		})
	})

	directory := filepath.Join(t.TempDir(), codegen.Gendir)
	writeGeneratedModule(t, directory, "generated.local/gen")
	_, err := generate(filepath.Dir(directory), "gen", false, registry)
	if err != nil {
		t.Fatalf("generate HTTP errors: %v", err)
	}
	runGeneratedTests(t, directory)
}
