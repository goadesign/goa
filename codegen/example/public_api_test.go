// This file protects small released helpers that plugins and generator tools
// can use without rebuilding example-generation data.
package example

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRootPathReturnsProjectImportPath checks the released project path helper.
func TestRootPathReturnsProjectImportPath(t *testing.T) {
	require.Equal(t, "goa.design/calc", RootPath("goa.design/calc/gen"))
	require.Equal(t, ".", RootPath("gen"))
}
