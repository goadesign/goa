package codegen

import (
    "path/filepath"
    "regexp"
    "strings"
    "testing"

    "github.com/stretchr/testify/require"

    "goa.design/goa/v3/jsonrpc/codegen/testdata"
)

// TestJSONRPCSSE_DedupEventTypes verifies the SSE server stream switch contains only
// one case for a shared event type used by multiple streaming endpoints.
func TestJSONRPCSSE_DedupEventTypes(t *testing.T) {
    root := RunJSONRPCDSL(t, testdata.JSONRPCSSEDuplicateEventDSL)
    services := CreateJSONRPCServices(root)

    // Generate JSON-RPC server files (includes service-level SSE impl when present)
    fs := ServerFiles("", services)
    require.NotEmpty(t, fs)

    // Render the service-level SSE stream implementation file (sse.go)
    var code string
    for _, f := range fs {
        if filepath.Base(f.Path) == "sse.go" && filepath.Base(filepath.Dir(f.Path)) == "server" {
            t.Logf("file: %s", f.Path)
            // Render all sections into a single source string
            var b strings.Builder
            for _, s := range f.SectionTemplates {
                _ = s.Write(&b)
            }
            code = b.String()
            break
        }
    }
    require.NotEmpty(t, code, "sse.go content not found")

    // Count case occurrences for the shared event type.
    // The generated code uses pattern: case *<pkg>.SharedSSEEvent:
    re := regexp.MustCompile(`(?m)^\s*case\s+\*[^.]+\.SharedSSEEvent\s*:`)
    matches := re.FindAllStringIndex(code, -1)
    require.Equal(t, 1, len(matches), "expected a single case for SharedSSEEvent, got %d\n%s", len(matches), code)
}
