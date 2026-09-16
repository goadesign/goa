// These tests compile complete services with custom path parameter types.
// They verify that both HTTP clients and servers import the packages used
// in their generated path function signatures.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

func TestGenerateCustomPathType(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		d.Service("products", func() {
			d.Method("lookup", func() {
				d.Payload(func() {
					d.Attribute("id", d.String, func() {
						d.Format(d.FormatUUID)
						d.Meta("struct:field:type", "uuid.UUID", "github.com/google/uuid")
					})
					d.Required("id")
				})
				d.Result(d.String)
				d.HTTP(func() {
					d.GET("/products/{id}")
				})
			})
		})
	})
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planTransportData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transport, err := testTransportFiles(plan)
	require.NoError(t, err)
	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range append(files, transport...) {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	runGeneratedTests(t, dir)
}
