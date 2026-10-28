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
// the wrapper message declared for a primitive or collection initial payload or
// stream item, so the .proto declares one request and one stream item message
// per method and the generated client and server compile.
func TestGenerateGRPCStreamEnvelopeWrappedItems(t *testing.T) {
	methods := []struct {
		name          string
		payload       any
		item          any
		bidirectional bool
		requestMsg    string
		streamItemMsg string
	}{
		{
			name:          "UploadBytes",
			item:          d.Bytes,
			requestMsg:    "UploadBytesRequest",
			streamItemMsg: "UploadBytesStreamItem",
		},
		{
			name:          "UploadInt",
			item:          d.Int,
			requestMsg:    "UploadIntRequest",
			streamItemMsg: "UploadIntStreamItem",
		},
		{
			name:          "UploadArray",
			item:          d.ArrayOf(d.String),
			requestMsg:    "UploadArrayRequest",
			streamItemMsg: "UploadArrayStreamItem",
		},
		{
			name:          "BidirectionalBytes",
			item:          d.Bytes,
			bidirectional: true,
			requestMsg:    "BidirectionalBytesRequest",
			streamItemMsg: "BidirectionalBytesStreamItem",
		},
		{
			name:          "UploadPrimitivePayload",
			payload:       d.Int,
			item:          d.Bytes,
			requestMsg:    "UploadPrimitivePayloadRequest",
			streamItemMsg: "UploadPrimitivePayloadStreamItem",
		},
		{
			name:          "UploadArrayPayload",
			payload:       d.ArrayOf(d.String),
			item:          d.Bytes,
			requestMsg:    "UploadArrayPayloadRequest",
			streamItemMsg: "UploadArrayPayloadStreamItem",
		},
	}

	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
	)
	_ = codegen.RunDSL(t, func() {
		d.Service("Upload", func() {
			for _, m := range methods {
				d.Method(m.name, func() {
					if m.payload != nil {
						d.Payload(m.payload)
					} else {
						d.Payload(func() {
							d.Field(1, "name", d.String)
							d.Required("name")
						})
					}
					d.StreamingPayload(m.item)
					if m.bidirectional {
						d.StreamingResult(d.String)
					} else {
						d.Result(d.String)
					}
					d.GRPC(func() {})
				})
			}
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
	proto := string(content)

	for _, m := range methods {
		reqDecls := regexp.MustCompile(`(?m)^message (`+regexp.QuoteMeta(m.requestMsg)+`\w*) \{`).FindAllStringSubmatch(proto, -1)
		require.Len(t, reqDecls, 1, "%s request messages:\n%s", m.name, proto)
		require.Equal(t, m.requestMsg, reqDecls[0][1])
		require.Contains(t, proto, m.requestMsg+" initial_payload = 1;")

		itemDecls := regexp.MustCompile(`(?m)^message (`+regexp.QuoteMeta(m.streamItemMsg)+`\w*) \{`).FindAllStringSubmatch(proto, -1)
		require.Len(t, itemDecls, 1, "%s stream item messages:\n%s", m.name, proto)
		require.Equal(t, m.streamItemMsg, itemDecls[0][1])
		require.Contains(t, proto, m.streamItemMsg+" stream_item = 2;")
	}

	runGeneratedTests(t, directory)
}
