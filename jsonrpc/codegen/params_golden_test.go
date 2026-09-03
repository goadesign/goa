// This file checks the generated JSON-RPC parameter shapes and how optional
// direct values keep absence separate from authored empty values.
package codegen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
)

// TestJSONRPCParamsGeneratedSource compares the generated branches for
// positional values, direct containers, and optional unions.
func TestJSONRPCParamsGeneratedSource(t *testing.T) {
	_, plan := linkedJSONRPCPlan(t, jsonRPCParamsPlanDSL)
	tests := []struct {
		name     string
		files    []*codegen.File
		file     string
		section  string
		contains string
		golden   string
	}{
		{
			name:     "primitive alias request encoder",
			files:    plan.ClientFiles(),
			file:     "encode_decode.go",
			section:  "jsonrpc-request-encoder",
			contains: "func EncodeAliasRequest",
			golden:   "testdata/golden/params_alias_request_encoder.go.golden",
		},
		{
			name:     "structured array request encoder",
			files:    plan.ClientFiles(),
			file:     "encode_decode.go",
			section:  "jsonrpc-request-encoder",
			contains: "func EncodeArrayRequest",
			golden:   "testdata/golden/params_array_request_encoder.go.golden",
		},
		{
			name:     "optional array request encoder",
			files:    plan.ClientFiles(),
			file:     "encode_decode.go",
			section:  "jsonrpc-request-encoder",
			contains: "func EncodeOptionalArrayRequest",
			golden:   "testdata/golden/params_optional_array_request_encoder.go.golden",
		},
		{
			name:     "required primitive request decoder",
			files:    plan.ServerFiles(),
			file:     "encode_decode.go",
			section:  "jsonrpc-request-decoder",
			contains: "func DecodeTextRequest",
			golden:   "testdata/golden/params_text_request_decoder.go.golden",
		},
		{
			name:     "optional primitive request decoder",
			files:    plan.ServerFiles(),
			file:     "encode_decode.go",
			section:  "jsonrpc-request-decoder",
			contains: "func DecodeOptionalTextRequest",
			golden:   "testdata/golden/params_optional_text_request_decoder.go.golden",
		},
		{
			name:     "optional union request decoder",
			files:    plan.ServerFiles(),
			file:     "encode_decode.go",
			section:  "jsonrpc-request-decoder",
			contains: "func DecodeOptionalUnionRequest",
			golden:   "testdata/golden/params_optional_union_request_decoder.go.golden",
		},
		{
			name:     "primitive stream writer",
			files:    plan.ServerFiles(),
			file:     "sse.go",
			section:  "jsonrpc-sse-server-stream",
			contains: "type StreamTextServerStream",
			golden:   "testdata/golden/params_text_stream_server.go.golden",
		},
		{
			name:     "optional primitive stream reader",
			files:    plan.ClientFiles(),
			file:     "stream.go",
			section:  "jsonrpc-sse-client-stream",
			contains: "StreamOptionalTextStreamImpl struct",
			golden:   "testdata/golden/params_optional_text_stream_client.go.golden",
		},
		{
			name:     "optional primitive stream writer",
			files:    plan.ServerFiles(),
			file:     "sse.go",
			section:  "jsonrpc-sse-server-stream",
			contains: "type StreamOptionalTextServerStream",
			golden:   "testdata/golden/params_optional_text_stream_server.go.golden",
		},
		{
			name:     "defaulted primitive stream reader",
			files:    plan.ClientFiles(),
			file:     "stream.go",
			section:  "jsonrpc-sse-client-stream",
			contains: "StreamDefaultTextStreamImpl struct",
			golden:   "testdata/golden/params_default_text_stream_client.go.golden",
		},
		{
			name:     "structured array stream writer",
			files:    plan.ServerFiles(),
			file:     "sse.go",
			section:  "jsonrpc-sse-server-stream",
			contains: "type StreamArrayServerStream",
			golden:   "testdata/golden/params_array_stream_server.go.golden",
		},
		{
			name:     "optional union stream writer",
			files:    plan.ServerFiles(),
			file:     "sse.go",
			section:  "jsonrpc-sse-server-stream",
			contains: "type StreamUnionDataServerStream",
			golden:   "testdata/golden/params_optional_union_stream_server.go.golden",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			section := paramsGoldenSection(t, test.files, test.file, test.section, test.contains)
			testutil.AssertGo(t, test.golden, codegen.SectionCode(t, section))
		})
	}
}

// TestJSONRPCDefaultedSelectedParamsGeneratedSource checks that generated
// clients preserve explicit empty values and generated servers apply defaults
// only when the selected request value is absent.
func TestJSONRPCDefaultedSelectedParamsGeneratedSource(t *testing.T) {
	_, plan := linkedJSONRPCPlan(t, jsonRPCDefaultedParamsGoldenDSL)
	tests := []struct {
		name     string
		files    []*codegen.File
		file     string
		section  string
		contains string
		golden   string
	}{
		{
			name:     "primitive alias client encoder",
			files:    plan.ClientFiles(),
			file:     "encode_decode.go",
			section:  "jsonrpc-request-encoder",
			contains: "func EncodeDefaultTextRequest",
			golden:   "testdata/golden/params_default_text_request_encoder.go.golden",
		},
		{
			name:     "collection client encoder",
			files:    plan.ClientFiles(),
			file:     "encode_decode.go",
			section:  "jsonrpc-request-encoder",
			contains: "func EncodeDefaultArrayRequest",
			golden:   "testdata/golden/params_default_array_request_encoder.go.golden",
		},
		{
			name:     "primitive alias server decoder",
			files:    plan.ServerFiles(),
			file:     "encode_decode.go",
			section:  "jsonrpc-request-decoder",
			contains: "func DecodeDefaultTextRequest",
			golden:   "testdata/golden/params_default_text_request_decoder.go.golden",
		},
		{
			name:     "collection server decoder",
			files:    plan.ServerFiles(),
			file:     "encode_decode.go",
			section:  "jsonrpc-request-decoder",
			contains: "func DecodeDefaultArrayRequest",
			golden:   "testdata/golden/params_default_array_request_decoder.go.golden",
		},
		{
			name:     "primitive alias server constructor",
			files:    plan.ServerTypeFiles(),
			file:     "types.go",
			section:  "server-payload-init",
			contains: "func NewDefaultTextPayload",
			golden:   "testdata/golden/params_default_text_payload_constructor.go.golden",
		},
		{
			name:     "collection server constructor",
			files:    plan.ServerTypeFiles(),
			file:     "types.go",
			section:  "server-payload-init",
			contains: "func NewDefaultArrayPayload",
			golden:   "testdata/golden/params_default_array_payload_constructor.go.golden",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			section := paramsGoldenSection(t, test.files, test.file, test.section, test.contains)
			testutil.AssertGo(t, test.golden, codegen.SectionCode(t, section))
		})
	}
}

// paramsGoldenSection finds one generated section by its function or type
// declaration instead of relying on a method's position in the design.
func paramsGoldenSection(t *testing.T, files []*codegen.File, fileName, sectionName, contains string) *codegen.SectionTemplate {
	t.Helper()
	for _, file := range files {
		if filepath.Base(file.Path) != fileName {
			continue
		}
		for _, section := range file.Section(sectionName) {
			if source := codegen.SectionCode(t, section); strings.Contains(source, contains) {
				return section
			}
		}
	}
	require.Fail(t, "generated section not found", "%s in %s", contains, fileName)
	return nil
}

// jsonRPCDefaultedParamsGoldenDSL declares one scalar alias and one collection
// selected as request bodies so generated source keeps their defaults honest.
func jsonRPCDefaultedParamsGoldenDSL() {
	mode := dsl.Type("Mode", dsl.String, func() {
		dsl.Pattern("^[a-z]*$")
		dsl.Default("safe")
	})
	values := dsl.Type("Values", dsl.ArrayOf(dsl.String))
	dsl.Service("defaults", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("default_text", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", mode)
			})
			dsl.JSONRPC(func() {
				dsl.Body("value")
			})
		})
		dsl.Method("default_array", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", values, func() {
					dsl.Default([]string{"safe"})
				})
			})
			dsl.JSONRPC(func() {
				dsl.Body("value")
			})
		})
	})
}
