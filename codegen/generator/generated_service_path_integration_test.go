// This file checks that every generated transport and example uses the service
// directory selected by the shared service plan.
package generator

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestGeneratedTransportsUsePlannedServicePaths checks both design orders so
// adding a suffix to one service directory cannot redirect another service's
// HTTP, JSON-RPC, gRPC, command-line, or example imports.
func TestGeneratedTransportsUsePlannedServicePaths(t *testing.T) {
	tests := []struct {
		name    string
		reverse bool
	}{
		{name: "forward"},
		{name: "reverse", reverse: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := servicePathRoot(t, test.reverse)
			plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planExampleData)
			expected := map[string]string{
				"read-value":  "read_value",
				"read_value":  "read_value3",
				"read_value2": "read_value2",
			}
			for _, service := range root.Services {
				require.Equal(t, expected[service.Name], plan.Service(root).Services().Get(service.Name).PathName)
			}

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

			generatedPaths := make(map[string]struct{}, len(files))
			for _, file := range files {
				generatedPaths[filepath.ToSlash(file.Path)] = struct{}{}
			}
			for _, servicePath := range expected {
				for _, filePath := range []string{
					"gen/" + servicePath + "/service.go",
					"gen/http/" + servicePath + "/server/server.go",
					"gen/jsonrpc/" + servicePath + "/server/server.go",
					"gen/grpc/" + servicePath + "/server/server.go",
				} {
					_, ok := generatedPaths[filePath]
					require.True(t, ok, "missing generated file %s", filePath)
				}
			}

			dir := t.TempDir()
			writeGeneratedModule(t, dir, "generated.local")
			for _, file := range files {
				_, err := file.Render(dir)
				require.NoError(t, err)
			}
			runGeneratedTests(t, dir)
		})
	}
}

// servicePathRoot returns one API whose services exercise every generated
// transport and both one-shot and streaming JSON-RPC output.
func servicePathRoot(t *testing.T, reverse bool) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		names := []string{"read-value", "read_value", "read_value2"}
		if reverse {
			names[0], names[2] = names[2], names[0]
		}
		servers := map[string]string{
			"read-value":  "dash",
			"read_value":  "underscore",
			"read_value2": "numbered",
		}
		dsl.API("path api", func() {
			for _, serviceName := range names {
				dsl.Server(servers[serviceName], func() {
					dsl.Services(serviceName)
					dsl.Host("local", func() { dsl.URI("http://localhost") })
				})
			}
		})
		for _, serviceName := range names {
			dsl.Service(serviceName, func() {
				dsl.Method("HTTP call", func() {
					dsl.Payload(dsl.String)
					dsl.Result(dsl.String)
					dsl.HTTP(func() { dsl.POST("/call") })
				})
				dsl.Method("JSON-RPC call", func() {
					dsl.Payload(dsl.String)
					dsl.Result(dsl.String)
					dsl.JSONRPC(func() {})
				})
				dsl.Method("JSON-RPC stream", func() {
					dsl.Payload(dsl.String)
					dsl.StreamingResult(dsl.String)
					dsl.JSONRPC(func() { dsl.ServerSentEvents() })
				})
				dsl.Method("gRPC call", func() {
					dsl.Payload(dsl.String)
					dsl.Result(dsl.String)
					dsl.GRPC(func() {})
				})
			})
		}
	})
}
