// These checks generate distinct raw JSON branches and compile their service,
// HTTP, and protobuf callers. Custom raw JSON keeps its bytes; ordinary bytes do not
// change their base64 representation in JSON.
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

func TestGenerateRawJSONByteUnions(t *testing.T) {
	root := codegen.RunDSL(t, rawJSONByteUnionDSL)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planTransportData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transports, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transports...)
	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	source, err := os.ReadFile(filepath.Join("testdata", "raw_json_unions", "roundtrip_test.go"))
	require.NoError(t, err)
	writeGeneratedContractTest(t, filepath.Join(dir, "gen"), "sample", string(source))
	runGeneratedTests(t, dir)
}

func rawJSONByteUnionDSL() {
	d.API("raw-json", func() {})
	raw := d.Type("RawValue", d.Bytes, func() {
		d.Meta("struct:field:type", "stdjson.RawMessage", "encoding/json", "stdjson")
	})
	operation := d.Type("Operation", func() {
		d.OneOf("value", func() {
			d.TypeName("OperationValue")
			d.Field(1, "json", d.Bytes, "Raw JSON value", func() {
				d.Meta("struct:field:type", "json.RawMessage", "encoding/json")
			})
		})
		d.Required("value")
	})
	two := d.Type("Two", func() {
		d.OneOf("value", func() {
			d.TypeName("TwoValue")
			d.Field(1, "first", d.Bytes, "First raw JSON value", func() {
				d.Meta("struct:field:type", "json.RawMessage", "encoding/json")
			})
			d.Field(2, "second", d.Bytes, "Second raw JSON value", func() {
				d.Meta("struct:field:type", "wirejson.RawMessage", "encoding/json", "wirejson")
			})
			d.Field(3, "named", raw, "Named raw JSON value")
			d.Field(4, "plain", d.Bytes, "Ordinary bytes")
		})
		d.Required("value")
	})
	nested := d.Type("Nested", func() {
		d.Field(1, "one", operation, "Single raw JSON choice")
		d.Field(2, "many", d.ArrayOf(two), "Ordered raw JSON choices")
		d.Field(3, "named", d.MapOf(d.String, two), "Raw JSON choices by name")
		d.Required("one", "many", "named")
	})
	d.Service("sample", func() {
		d.Method("echo", func() {
			d.Payload(operation)
			d.Result(operation)
			d.GRPC(func() {})
			d.HTTP(func() {
				d.POST("/echo")
				d.Response(d.StatusOK)
			})
		})
		d.Method("two", func() {
			d.Payload(two)
			d.Result(two)
			d.GRPC(func() {})
			d.HTTP(func() {
				d.POST("/two")
				d.Response(d.StatusOK)
			})
		})
		d.Method("nested", func() {
			d.Payload(nested)
			d.Result(nested)
			d.GRPC(func() {})
			d.HTTP(func() {
				d.POST("/nested")
				d.Response(d.StatusOK)
			})
		})
	})
}
