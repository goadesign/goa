// This file verifies that transport client files import generated views only
// when one of their emitted response or stream-receive sections references it.
package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

// TestViewedTransportClientImportsCompile verifies that HTTP and JSON-RPC
// response decoders and gRPC response and stream decoders receive the exact
// generated views import used by their rendered sections. The unary gRPC
// client also proves client.go does not reserve the codec-only import.
func TestViewedTransportClientImportsCompile(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

	codegen.RunDSL(t, func() {
		result := dsl.ResultType("application/vnd.viewed-import", func() {
			dsl.Field(1, "value", dsl.String)
			dsl.Required("value")
			dsl.View("default", func() {
				dsl.Attribute("value")
			})
		})
		dsl.Service("ViewedHTTPJSON", func() {
			dsl.Method("HTTP", func() {
				dsl.Result(result)
				dsl.HTTP(func() {
					dsl.GET("/http")
				})
			})
			dsl.Method("JSONRPC", func() {
				dsl.Result(result)
				dsl.JSONRPC(func() {})
			})
		})
		dsl.Service("ViewedUnary", func() {
			dsl.Method("GRPCUnary", func() {
				dsl.Result(result)
				dsl.GRPC(func() {})
			})
		})
		dsl.Service("ViewedStream", func() {
			dsl.Method("GRPCStream", func() {
				dsl.StreamingResult(result)
				dsl.GRPC(func() {})
			})
		})
		dsl.Service("ViewedHTTPSSE", func() {
			dsl.Method("Events", func() {
				dsl.StreamingResult(result)
				dsl.HTTP(func() {
					dsl.GET("/http-sse")
					dsl.ServerSentEvents("value")
				})
			})
		})
		dsl.Service("ViewedHTTPWebSocket", func() {
			dsl.Method("Events", func() {
				dsl.StreamingResult(result)
				dsl.HTTP(func() {
					dsl.GET("/http-websocket")
				})
			})
		})
		dsl.Service("Ordinary", func() {
			dsl.Method("HTTP", func() {
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.GET("/ordinary")
				})
			})
			dsl.Method("JSONRPC", func() {
				dsl.Result(dsl.String)
				dsl.JSONRPC(func() {})
			})
			dsl.Method("GRPCUnary", func() {
				dsl.Result(dsl.String)
				dsl.GRPC(func() {})
			})
			dsl.Method("GRPCStream", func() {
				dsl.StreamingResult(dsl.String)
				dsl.GRPC(func() {})
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "generated.local/gen")
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	httpJSON := codegen.SnakeCase("ViewedHTTPJSON")
	unary := codegen.SnakeCase("ViewedUnary")
	stream := codegen.SnakeCase("ViewedStream")
	assertFilesImportPath(t, genDir, "/"+httpJSON+"/views\"", []string{
		filepath.Join("http", httpJSON, "client", "encode_decode.go"),
		filepath.Join("http", httpJSON, "client", "types.go"),
		filepath.Join("http", httpJSON, "server", "encode_decode.go"),
		filepath.Join("http", httpJSON, "server", "types.go"),
		filepath.Join("jsonrpc", httpJSON, "client", "encode_decode.go"),
		filepath.Join("jsonrpc", httpJSON, "client", "types.go"),
		filepath.Join("jsonrpc", httpJSON, "server", "server.go"),
		filepath.Join("jsonrpc", httpJSON, "server", "types.go"),
	})
	assertNoImportPath(t, filepath.Join(genDir, "grpc", unary, "client", "client.go"), "/"+unary+"/views\"")
	assertFilesImportPath(t, genDir, "/"+unary+"/views\"", []string{
		filepath.Join("grpc", unary, "client", "encode_decode.go"),
		filepath.Join("grpc", unary, "client", "types.go"),
		filepath.Join("grpc", unary, "server", "encode_decode.go"),
		filepath.Join("grpc", unary, "server", "types.go"),
	})
	assertFilesImportPath(t, genDir, "/"+stream+"/views\"", []string{
		filepath.Join("grpc", stream, "client", "client.go"),
		filepath.Join("grpc", stream, "client", "types.go"),
		filepath.Join("grpc", stream, "server", "encode_decode.go"),
		filepath.Join("grpc", stream, "server", "types.go"),
	})
	assertNoImportPath(t, filepath.Join(genDir, "grpc", stream, "client", "encode_decode.go"), "/"+stream+"/views\"")
	assertViewedStreamingTransportFiles(t, genDir)
	ordinary := codegen.SnakeCase("Ordinary")
	assertFilesOmitImportPath(t, genDir, "/"+ordinary+"/views\"", []string{
		filepath.Join("http", ordinary, "client", "encode_decode.go"),
		filepath.Join("http", ordinary, "client", "types.go"),
		filepath.Join("http", ordinary, "server", "encode_decode.go"),
		filepath.Join("http", ordinary, "server", "types.go"),
		filepath.Join("jsonrpc", ordinary, "client", "encode_decode.go"),
		filepath.Join("jsonrpc", ordinary, "client", "types.go"),
		filepath.Join("jsonrpc", ordinary, "server", "server.go"),
		filepath.Join("jsonrpc", ordinary, "server", "types.go"),
		filepath.Join("grpc", ordinary, "client", "client.go"),
		filepath.Join("grpc", ordinary, "client", "encode_decode.go"),
		filepath.Join("grpc", ordinary, "client", "types.go"),
		filepath.Join("grpc", ordinary, "server", "encode_decode.go"),
		filepath.Join("grpc", ordinary, "server", "types.go"),
	})
	runGeneratedTests(t, genDir)
}

// assertViewedStreamingTransportFiles verifies that HTTP SSE and WebSocket
// services render and that only the WebSocket receive file imports views
// directly. Server send files call the service constructor and therefore do
// not import the views package themselves.
func assertViewedStreamingTransportFiles(t *testing.T, genDir string) {
	t.Helper()
	httpSSE := codegen.SnakeCase("ViewedHTTPSSE")
	httpWebSocket := codegen.SnakeCase("ViewedHTTPWebSocket")
	for _, path := range []string{
		filepath.Join("http", httpSSE, "server", "sse.go"),
		filepath.Join("http", httpSSE, "client", "sse.go"),
		filepath.Join("http", httpWebSocket, "server", "websocket.go"),
	} {
		require.FileExists(t, filepath.Join(genDir, path))
	}
	assertImportPath(
		t,
		filepath.Join(genDir, "http", httpWebSocket, "client", "websocket.go"),
		"/"+httpWebSocket+"/views\"",
	)
	assertFilesOmitImportPath(t, genDir, "/"+httpSSE+"/views\"", []string{
		filepath.Join("http", httpSSE, "server", "sse.go"),
		filepath.Join("http", httpSSE, "client", "sse.go"),
	})
	assertNoImportPath(
		t,
		filepath.Join(genDir, "http", httpWebSocket, "server", "websocket.go"),
		"/"+httpWebSocket+"/views\"",
	)
}

// assertFilesImportPath verifies that each generated file imports a package
// whose full path ends with suffix.
func assertFilesImportPath(t *testing.T, root, suffix string, paths []string) {
	t.Helper()
	for _, path := range paths {
		assertImportPath(t, filepath.Join(root, path), suffix)
	}
}

// assertFilesOmitImportPath verifies that none of the generated files reserve
// the package whose full path ends with suffix.
func assertFilesOmitImportPath(t *testing.T, root, suffix string, paths []string) {
	t.Helper()
	for _, path := range paths {
		assertNoImportPath(t, filepath.Join(root, path), suffix)
	}
}

// assertImportPath verifies that a generated file imports a package whose full
// path ends with suffix.
func assertImportPath(t *testing.T, path, suffix string) {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(content), suffix)
}

// assertNoImportPath verifies that a generated file does not reserve an import
// for a package unused by its rendered sections.
func assertNoImportPath(t *testing.T, path, suffix string) {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.False(t, strings.Contains(string(content), suffix), string(content))
}
