package codegen_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/generator"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/jsonrpc/codegen/testdata"
)

// TestJSONRPCKitchenSink pins every file the transport and example generators
// produce for a design covering the full JSON-RPC surface (plain methods,
// required/optional IDs, custom errors, WebSocket, SSE, mixed HTTP+JSON-RPC
// transports). Each rendered file is compared against a golden copy under
// testdata/golden/kitchen_sink and the set of generated paths is compared
// against a manifest so files that appear or disappear fail the test.
func TestJSONRPCKitchenSink(t *testing.T) {
	root := expr.RunDSL(t, testdata.JSONRPCKitchenSinkDSL)
	roots := []eval.Root{root}

	tfiles, err := generator.Transport("kitchensink", roots)
	require.NoError(t, err)
	efiles, err := generator.Example("kitchensink", roots)
	require.NoError(t, err)

	tmp := t.TempDir()
	for _, f := range append(tfiles, efiles...) {
		_, err := f.Render(tmp)
		require.NoError(t, err)
	}

	var paths []string
	err = filepath.WalkDir(tmp, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(tmp, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	require.NoError(t, err)
	sort.Strings(paths)

	goldenDir := filepath.Join("testdata", "golden", "kitchen_sink")
	manifest := strings.Join(paths, "\n") + "\n"
	testutil.CompareOrUpdateGolden(t, manifest, filepath.Join(goldenDir, "manifest.golden"))

	for _, rel := range paths {
		content, err := os.ReadFile(filepath.Join(tmp, rel))
		require.NoError(t, err)
		testutil.CompareOrUpdateGolden(t, string(content), filepath.Join(goldenDir, rel+".golden"))
	}
}
