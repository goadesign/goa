package codegen

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestIdempotentHTTPEndpointCodegen(t *testing.T) {
	root := expr.RunDSL(t, func() {
		Service("Retry", func() {
			Method("read", func() {
				Idempotent()
				Error("busy", func() {
					Temporary()
				})
				HTTP(func() {
					GET("/read")
					Response("busy", StatusServiceUnavailable)
				})
			})
			Method("write", func() {
				HTTP(func() {
					POST("/write")
				})
			})
		})
	})
	services := CreateHTTPServices(root)
	clientFiles := ClientFiles("", services)
	require.NotEmpty(t, clientFiles)

	clientCode := codegen.SectionsCode(t, clientFiles[0].Section("client-endpoint-init"))

	assert.Contains(t, clientCode, `goa.RetryEndpoint(endpoint, "busy")`)
	assert.Equal(t, 1, strings.Count(clientCode, "goa.RetryEndpoint("))
}

// TestFileGenerationIdempotent builds the HTTP services data once and renders
// the complete generated file set twice, asserting that both renders produce
// byte-identical outputs. This guards against file generators mutating shared
// analysis state (e.g. the ServerTypeNames/ClientTypeNames dedup sets or the
// PathInit argument data) in ways that change subsequent renders.
func TestFileGenerationIdempotent(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"websocket-mixed-endpoints", testdata.MixedEndpointsDSL},
		{"payload-body-user-inner", testdata.PayloadBodyUserInnerDSL},
		{"result-body-multiple-views", testdata.ResultBodyMultipleViewsDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			services := CreateHTTPServices(root)

			render := func(dir string) {
				files := PathFiles(services)
				files = append(files, ServerFiles("gen", services)...)
				files = append(files, ClientFiles("gen", services)...)
				files = append(files, ServerTypeFiles("gen", services)...)
				files = append(files, ClientTypeFiles("gen", services)...)
				require.NotEmpty(t, files)
				for _, f := range files {
					_, err := f.Render(dir)
					require.NoError(t, err)
				}
			}

			first, second := t.TempDir(), t.TempDir()
			render(first)
			render(second)

			got := readTree(t, second)
			want := readTree(t, first)
			require.Equal(t, keys(want), keys(got), "rendered file sets differ between runs")
			for path, content := range want {
				if string(got[path]) != string(content) {
					t.Errorf("%s: content differs between first and second render", path)
				}
			}
		})
	}
}

// readTree returns the contents of all regular files under dir keyed by path
// relative to dir.
func readTree(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	tree := make(map[string][]byte)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path) // #nosec G304 -- test reads from t.TempDir
		if err != nil {
			return err
		}
		tree[rel] = content
		return nil
	})
	require.NoError(t, err)
	return tree
}

// keys returns the sorted keys of m.
func keys(m map[string][]byte) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	slices.Sort(ks)
	return ks
}
