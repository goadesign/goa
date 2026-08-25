package codegen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/jsonrpc/codegen/testdata"
)

// TestJSONRPCSSE_DedupEventTypes verifies the generated SSE server file
// declares the shared SSE stream machinery exactly once when multiple
// streaming endpoints share the same event type, with one stream type per
// endpoint.
func TestJSONRPCSSE_DedupEventTypes(t *testing.T) {
	root := expr.RunDSL(t, testdata.JSONRPCSSEDuplicateEventDSL)
	plan := CreateJSONRPCPlan(root)

	// Generate JSON-RPC server files (includes the SSE streams file)
	fs := plan.ServerFiles()
	require.NotEmpty(t, fs)

	// Render the shared writer and the endpoint streams.
	var serverCode, streamCode string
	for _, f := range fs {
		name := filepath.Base(f.Path)
		if (name == "server.go" || name == "sse.go") && filepath.Base(filepath.Dir(f.Path)) == "server" {
			t.Logf("file: %s", f.Path)
			var b strings.Builder
			for _, s := range f.SectionTemplates {
				require.NoError(t, s.Write(&b))
			}
			if name == "server.go" {
				serverCode = b.String()
			} else {
				streamCode = b.String()
			}
		}
	}
	require.NotEmpty(t, serverCode, "server.go content not found")
	require.NotEmpty(t, streamCode, "sse.go content not found")

	// The shared machinery must be declared exactly once.
	require.Equal(t, 1, strings.Count(serverCode, "sseServerStream struct"), "expected a single sseServerStream declaration\n%s", serverCode)
	require.Equal(t, 1, strings.Count(serverCode, "sseEventBuffer struct"), "expected a single sseEventBuffer declaration\n%s", serverCode)
	require.NotContains(t, streamCode, "sseServerStream struct")
	require.NotContains(t, streamCode, "sseEventBuffer struct")

	// Each endpoint gets its own stream type even when sharing the event type.
	require.Equal(t, 1, strings.Count(streamCode, "type StreamAServerStream struct"), "expected a single StreamA stream declaration\n%s", streamCode)
	require.Equal(t, 1, strings.Count(streamCode, "type StreamBServerStream struct"), "expected a single StreamB stream declaration\n%s", streamCode)
}
