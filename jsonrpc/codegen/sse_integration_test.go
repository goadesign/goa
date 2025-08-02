package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/jsonrpc/codegen/testdata"
)

func TestJSONRPCSSEIntegration(t *testing.T) {
	// Skip if not in CI or explicitly requested
	if os.Getenv("GOA_INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test. Set GOA_INTEGRATION_TEST=1 to run.")
	}

	// Run the DSL
	root := RunJSONRPCDSL(t, testdata.JSONRPCSSEObjectDSL)
	services := CreateJSONRPCServices(root)
	
	// Generate all files
	serverFiles := ServerFiles("", services)
	clientFiles := ClientFiles("", services)
	sseFiles := SSEServerFiles("", services)
	
	// Combine all files
	allFiles := append(serverFiles, clientFiles...)
	allFiles = append(allFiles, sseFiles...)
	
	// Create temp directory
	tmpDir := t.TempDir()
	
	// Write all files
	for _, f := range allFiles {
		t.Logf("Rendering file: %s", f.Path)
		fullPath, err := f.Render(tmpDir)
		require.NoError(t, err)
		t.Logf("  -> Full path: %s", fullPath)
	}
	
	// Try to compile the generated code
	// This would require setting up go.mod, etc. so we'll just verify
	// that files were generated with expected content
	
	// Check key files exist
	serverStreamPath := filepath.Join(tmpDir, "gen/jsonrpc/jsonrpcsse_object_service/server/stream.go")
	require.FileExists(t, serverStreamPath)
	
	clientStreamPath := filepath.Join(tmpDir, "gen/jsonrpc/jsonrpcsse_object_service/client/stream.go")
	require.FileExists(t, clientStreamPath)
	
	// Read and verify server stream has JSON-RPC notification code
	serverContent, err := os.ReadFile(serverStreamPath)
	require.NoError(t, err)
	require.Contains(t, string(serverContent), "JSON-RPC notification")
	require.Contains(t, string(serverContent), `"jsonrpc": "2.0"`)
	require.Contains(t, string(serverContent), "text/event-stream")
	
	// Read and verify client stream has JSON-RPC decoding
	clientContent, err := os.ReadFile(clientStreamPath)
	require.NoError(t, err)
	require.Contains(t, string(clientContent), "decodeResult")
	require.Contains(t, string(clientContent), "JSON-RPC notification")
}