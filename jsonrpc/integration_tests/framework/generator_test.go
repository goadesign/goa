// This file checks the design source rendered by the JSON-RPC integration
// test generator before Goa turns that design into transport code.
package framework

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRenderDesignUsesIDForMappedRequestID ensures the integration fixture
// declares the envelope-owned request field with Goa's JSON-RPC ID function.
func TestRenderDesignUsesIDForMappedRequestID(t *testing.T) {
	dir := t.TempDir()
	generator := NewGenerator(dir, map[string]MethodInfo{
		"echo_string_idmap": {
			Action:    ActionEcho,
			Type:      TypeString,
			Modifier:  ModifierIDMap,
			Transport: TransportHTTP,
		},
	})

	require.NoError(t, generator.renderDesign(generator.buildDesignData()))
	design, err := os.ReadFile(filepath.Join(dir, "design", "design.go"))
	require.NoError(t, err)
	require.Contains(t, string(design), `ID("request_id", String)`)
	require.False(t, strings.Contains(string(design), `Field(2, "request_id", String)`))
}
