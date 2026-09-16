// This file assembles complete Files examples and verifies that the generator
// writes servers for static files without inventing clients that have no
// method to call.
package generator

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

func TestFilesExamplesGenerateOnlyCallableClients(t *testing.T) {
	tests := []struct {
		name            string
		design          func()
		present         []string
		absent          []string
		absentFragments []string
	}{
		{
			name:   "files-only",
			design: filesOnlyExampleDSL,
			present: []string{
				filepath.Join("cmd", "public", "http.go"),
				filepath.Join("gen", "http", "assets", "server", "server.go"),
			},
			absentFragments: []string{
				"/cli/",
				"-cli/",
			},
		},
		{
			name:   "files-with-grpc",
			design: filesWithGRPCExampleDSL,
			present: []string{
				filepath.Join("cmd", "public-cli", "main.go"),
				filepath.Join("cmd", "public-cli", "grpc.go"),
				filepath.Join("gen", "grpc", "cli", "public", "cli.go"),
			},
			absent: []string{
				"cmd/public-cli/http.go",
				"gen/http/cli/public/cli.go",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory, paths := renderCompleteExample(t, test.design)
			testutil.CompareOrUpdateGolden(
				t,
				strings.Join(paths, "\n"),
				filepath.Join("testdata", "golden", "files_client", test.name, "paths.golden"),
			)

			for _, path := range test.present {
				require.Contains(t, paths, filepath.ToSlash(path))
				content, err := os.ReadFile(filepath.Join(directory, path))
				require.NoError(t, err)
				testutil.AssertGo(
					t,
					filepath.Join("testdata", "golden", "files_client", test.name, path+".golden"),
					string(content),
				)
			}
			for _, path := range test.absent {
				require.NotContains(t, paths, path)
			}
			for _, path := range paths {
				for _, fragment := range test.absentFragments {
					require.NotContains(t, path, fragment)
				}
			}
			runGeneratedTests(t, directory)
		})
	}
}

// renderCompleteExample assembles service, transport, and example files,
// writes them into one module, and returns every generated path.
func renderCompleteExample(t *testing.T, design func()) (string, []string) {
	t.Helper()
	root := codegen.RunDSL(t, design)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planExampleData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transportFiles, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transportFiles...)
	exampleFiles, err := assembleExampleFilesForTest(plan)
	require.NoError(t, err)
	files = append(files, exampleFiles...)
	files, err = mergeFilesByPath(files)
	require.NoError(t, err)

	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	paths := make([]string, len(files))
	for index, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
		paths[index] = filepath.ToSlash(file.Path)
	}
	sort.Strings(paths)
	return directory, paths
}

// filesOnlyExampleDSL defines an HTTP server whose only routes serve files.
func filesOnlyExampleDSL() {
	dsl.API("files", func() {
		dsl.Server("public", func() {
			dsl.Services("assets")
			dsl.Host("local", func() {
				dsl.URI("http://localhost:8080")
			})
		})
	})
	dsl.Service("assets", func() {
		dsl.Files("/index.html", "index.html")
	})
}

// filesWithGRPCExampleDSL adds one callable gRPC method to the same service
// that serves static files over HTTP.
func filesWithGRPCExampleDSL() {
	dsl.API("files", func() {
		dsl.Server("public", func() {
			dsl.Services("assets")
			dsl.Host("local", func() {
				dsl.URI("http://localhost:8080")
				dsl.URI("grpc://localhost:8080")
			})
		})
	})
	dsl.Service("assets", func() {
		dsl.Files("/index.html", "index.html")
		dsl.Method("read_status", func() {
			dsl.Result(dsl.String)
			dsl.GRPC(func() {})
		})
	})
}
