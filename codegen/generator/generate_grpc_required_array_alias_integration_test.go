// This file verifies that generated gRPC codecs compile for required arrays
// whose service elements are primitive aliases.
package generator

import (
	"path/filepath"
	"testing"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
)

// TestGenerateGRPCRequiredPrimitiveAliasArray proves the service-to-protobuf
// and protobuf-to-service conversions preserve required string alias elements.
func TestGenerateGRPCRequiredPrimitiveAliasArray(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
	)

	_ = codegen.RunDSL(t, func() {
		alias := d.Type("Alias", d.String)
		payload := d.Type("Payload", func() {
			d.Field(1, "values", d.ArrayOfRequired(alias))
			d.Required("values")
		})
		d.Service("Aliases", func() {
			d.Method("Store", func() {
				d.Payload(payload)
				d.GRPC(func() {})
			})
		})
	})

	directory := filepath.Join(t.TempDir(), codegen.Gendir)
	writeGeneratedModule(t, directory, "generated.local/gen")
	if _, err := generate(filepath.Dir(directory), "gen", false, registry); err != nil {
		t.Fatalf("generate required primitive alias array: %v", err)
	}
	runGeneratedTests(t, directory)
}
