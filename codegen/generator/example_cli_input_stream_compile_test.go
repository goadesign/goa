// This file compiles generated HTTP and gRPC example clients whose methods
// require streamed input. The example client must reject those commands
// without leaving unused endpoint values in the generated program.
package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

// TestExampleInputStreamClientsCompile generates client-streaming and
// bidirectional commands for HTTP and gRPC, then compiles the generated client.
func TestExampleInputStreamClientsCompile(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.API("gRPC input streams", func() {
			dsl.Server("stream", func() {
				dsl.Services("events")
				dsl.Host("local", func() {
					dsl.URI("http://localhost:8080")
					dsl.URI("grpc://localhost:8080")
				})
			})
		})
		dsl.Service("events", func() {
			dsl.Method("upload", func() {
				dsl.StreamingPayload(dsl.String)
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.POST("/upload")
				})
				dsl.GRPC(func() {})
			})
			dsl.Method("exchange", func() {
				dsl.StreamingPayload(dsl.String)
				dsl.StreamingResult(dsl.String)
				dsl.HTTP(func() {
					dsl.POST("/exchange")
				})
				dsl.GRPC(func() {})
			})
		})
	})
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planExampleData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transportFiles, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transportFiles...)
	exampleFiles, err := assembleExampleFilesForTest(plan)
	require.NoError(t, err)
	for _, file := range exampleFiles {
		if strings.HasPrefix(file.Path, filepath.Join("cmd", "stream-cli")) {
			files = append(files, file)
		}
	}
	files, err = mergeFilesByPath(files)
	require.NoError(t, err)

	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	runGeneratedTests(t, directory)
}
