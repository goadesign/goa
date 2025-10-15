package codegen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/jsonrpc/codegen/testdata"
)

func TestJSONRPCSSE(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"string", testdata.JSONRPCSSEStringDSL},
		{"object", testdata.JSONRPCSSEObjectDSL},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunJSONRPCDSL(t, c.DSL)
			services := CreateJSONRPCServices(root)
			
			// Generate SSE files
			fs := SSEServerFiles("", services)
			require.NotEmpty(t, fs, "expected SSE files to be generated")
			
			// Debug: print all generated files
			for _, f := range fs {
				t.Logf("Generated file: %s", f.Path)
			}
			
			// Find the server stream file
			var serverStreamFile *codegen.File
			for _, f := range fs {
				if filepath.Base(f.Path) == "stream.go" && filepath.Base(filepath.Dir(f.Path)) == "server" {
					serverStreamFile = f
					break
				}
			}
			require.NotNil(t, serverStreamFile, "server stream file not found")
			
			// Find the jsonrpc-sse-server-stream section
			var streamSection *codegen.SectionTemplate
			for _, s := range serverStreamFile.SectionTemplates {
				if s.Name == "jsonrpc-sse-server-stream" {
					streamSection = s
					break
				}
			}
			require.NotNil(t, streamSection, "jsonrpc-sse-server-stream section not found")
			
			// Compare with golden file
			code := codegen.SectionCode(t, streamSection)
			golden := filepath.Join("testdata", "golden", "jsonrpc-sse-"+c.Name+".golden")
			testutil.CompareOrUpdateGolden(t, code, golden)
		})
	}
}