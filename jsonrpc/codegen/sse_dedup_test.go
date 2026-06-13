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
	services := CreateJSONRPCServices(root)

	// Generate JSON-RPC server files (includes the SSE streams file)
	fs := ServerFiles("", services)
	require.NotEmpty(t, fs)

	// Render the SSE streams file (sse.go)
	var code string
	for _, f := range fs {
		if filepath.Base(f.Path) == "sse.go" && filepath.Base(filepath.Dir(f.Path)) == "server" {
			t.Logf("file: %s", f.Path)
			// Render all sections into a single source string
			var b strings.Builder
			for _, s := range f.SectionTemplates {
				require.NoError(t, s.Write(&b))
			}
			code = b.String()
			break
		}
	}
	require.NotEmpty(t, code, "sse.go content not found")

	// The shared machinery must be declared exactly once.
	require.Equal(t, 1, strings.Count(code, "type sseServerStream struct"), "expected a single sseServerStream declaration\n%s", code)
	require.Equal(t, 1, strings.Count(code, "type sseEventWriter struct"), "expected a single sseEventWriter declaration\n%s", code)

	// Each endpoint gets its own stream type even when sharing the event type.
	require.Equal(t, 1, strings.Count(code, "type StreamAServerStream struct"), "expected a single StreamA stream declaration\n%s", code)
	require.Equal(t, 1, strings.Count(code, "type StreamBServerStream struct"), "expected a single StreamB stream declaration\n%s", code)
}
