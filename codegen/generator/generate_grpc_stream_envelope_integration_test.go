// This file verifies that gRPC methods carrying both a one-shot payload and
// streamed payload items generate compiling envelope code.
package generator

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
)

// TestGenerateGRPCStreamEnvelopeWrappedItems proves the stream envelope reuses
// the wrapper message declared for a primitive or collection stream item, so
// the .proto declares one item message and the generated client and server
// compile.
func TestGenerateGRPCStreamEnvelopeWrappedItems(t *testing.T) {
	tests := []struct {
		name          string
		item          any
		bidirectional bool
	}{
		{name: "client bytes", item: d.Bytes},
		{name: "client int", item: d.Int},
		{name: "client array", item: d.ArrayOf(d.String)},
		{name: "bidirectional bytes", item: d.Bytes, bidirectional: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := testRegistry(
				"gen",
				testGenerator(planServiceData, testServiceFiles),
				testGenerator(planTransportData, testTransportFiles),
			)
			_ = codegen.RunDSL(t, func() {
				d.Service("Upload", func() {
					d.Method("Upload", func() {
						d.Payload(func() {
							d.Field(1, "name", d.String)
							d.Required("name")
						})
						d.StreamingPayload(tt.item)
						if tt.bidirectional {
							d.StreamingResult(d.String)
						} else {
							d.Result(d.String)
						}
						d.GRPC(func() {})
					})
				})
			})

			directory := filepath.Join(t.TempDir(), codegen.Gendir)
			writeGeneratedModule(t, directory, "generated.local/gen")
			_, err := generate(filepath.Dir(directory), "gen", false, registry)
			require.NoError(t, err)

			protoFiles, err := filepath.Glob(filepath.Join(directory, "grpc", "upload", "pb", "*.proto"))
			require.NoError(t, err)
			require.Len(t, protoFiles, 1)
			content, err := os.ReadFile(protoFiles[0])
			require.NoError(t, err)
			messages := regexp.MustCompile(`(?m)^message (\w*StreamItem\w*) \{`).FindAllStringSubmatch(string(content), -1)
			require.Len(t, messages, 1, "stream item messages:\n%s", content)
			require.Contains(t, string(content), messages[0][1]+" stream_item = 2;")
			runGeneratedTests(t, directory)
		})
	}
}
