// This file checks that starter files are preserved relative to the requested
// output directory, even when generation starts in a different directory.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
)

func TestExampleGenerationUsesOutputDirectoryForPreservation(t *testing.T) {
	paths := []string{
		"status.go",
		filepath.Join("interceptors", "status_server.go"),
		filepath.Join("interceptors", "status_client.go"),
		"multipart.go",
		filepath.Join("cmd", "preserve", "main.go"),
		filepath.Join("cmd", "preserve", "http.go"),
		filepath.Join("cmd", "preserve", "grpc.go"),
		filepath.Join("cmd", "preserve-cli", "main.go"),
		filepath.Join("cmd", "preserve-cli", "http.go"),
		filepath.Join("cmd", "preserve-cli", "grpc.go"),
	}
	preserved := map[string][]byte{
		"status.go": []byte("existing service\n"),
		filepath.Join("interceptors", "status_server.go"): []byte("existing interceptor\n"),
		"multipart.go": []byte("existing multipart helpers\n"),
		filepath.Join("cmd", "preserve", "grpc.go"):     []byte("existing server\n"),
		filepath.Join("cmd", "preserve-cli", "http.go"): []byte("existing client\n"),
	}
	tests := []struct {
		name     string
		existing map[string][]byte
	}{
		{name: "output is empty"},
		{name: "output has starter files", existing: preserved},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			codegen.RunDSL(t, exampleOutputPreservationDSL)
			launchDir := t.TempDir()
			outputDir := t.TempDir()
			writeGeneratedModule(t, filepath.Join(outputDir, codegen.Gendir), "generated.local/gen")
			writeExampleFiles(t, launchDir, paths, []byte("file from launch directory\n"))
			for path, content := range test.existing {
				writeExampleFile(t, outputDir, path, content)
			}

			t.Chdir(launchDir)
			_, err := generate(outputDir, "example", false, newDefaultRegistry())
			require.NoError(t, err)

			for _, path := range paths {
				content, err := os.ReadFile(filepath.Join(outputDir, path))
				require.NoError(t, err, path)
				if existing, ok := test.existing[path]; ok {
					require.Equal(t, existing, content, path)
				} else {
					require.NotEqual(t, []byte("file from launch directory\n"), content, path)
				}
			}
		})
	}
}

// exampleOutputPreservationDSL exercises every starter file producer that
// preserves user-written files.
func exampleOutputPreservationDSL() {
	trace := d.Interceptor("trace")
	d.API("preserve", func() {})
	d.Service("status", func() {
		d.ServerInterceptor(trace)
		d.ClientInterceptor(trace)
		d.Method("upload", func() {
			d.Payload(func() {
				d.Field(1, "message", d.String)
			})
			d.Result(d.String)
			d.HTTP(func() {
				d.POST("/upload")
				d.MultipartRequest()
			})
			d.GRPC(func() {})
		})
	})
}

// writeExampleFiles writes the same misleading content at every relative path
// in the directory where generation starts.
func writeExampleFiles(t *testing.T, dir string, paths []string, content []byte) {
	t.Helper()
	for _, path := range paths {
		writeExampleFile(t, dir, path, content)
	}
}

// writeExampleFile writes one fixture file and creates its parent directory.
func writeExampleFile(t *testing.T, dir, path string, content []byte) {
	t.Helper()
	fullPath := filepath.Join(dir, path)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o750))
	require.NoError(t, os.WriteFile(fullPath, content, 0o600))
}
