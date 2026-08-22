// This file checks that generated files use the exact package names assigned
// before the files are written.
package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	httpHelperCollisionOrder string
	jsonRPCSharedPackageMode uint8
)

const (
	jsonRPCUnary jsonRPCSharedPackageMode = iota
	jsonRPCSSE
	jsonRPCWebSocket
)

// ComparePackageName orders names added by the collision test.
func (o httpHelperCollisionOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	return strings.Compare(string(o), string(other.(httpHelperCollisionOrder)))
}

// TestGeneratedTransportPackagesCompileWithServiceAliasCollisions proves that
// client, server, protobuf, command-line, and service imports still name the
// service they belong to when service names produce the same Go import name.
func TestGeneratedTransportPackagesCompileWithServiceAliasCollisions(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		interceptor := dsl.Interceptor("Trace", func() {})
		for _, name := range []string{"Foo", "Fooc", "Foosvr", "Foojssvr"} {
			dsl.Service(name, func() {
				if name == "Foo" {
					dsl.ClientInterceptor(interceptor)
				}
				dsl.Method("Read", func() {
					dsl.Payload(dsl.String)
					dsl.Result(dsl.String)
					dsl.HTTP(func() { dsl.POST("/" + strings.ToLower(name)) })
					dsl.GRPC(func() {})
				})
				dsl.Method("ReadJSON", func() {
					dsl.Payload(dsl.String)
					dsl.Result(dsl.String)
					dsl.JSONRPC(func() {})
				})
			})
		}
	})

	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planTransportData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transport, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transport...)
	exampleFiles, err := assembleExampleFilesForTest(plan)
	require.NoError(t, err)
	files = append(files, exampleFiles...)

	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	runGeneratedTests(t, dir)
}

// TestGeneratedHTTPHelpersCompileWithPackageNameCollisions checks that file and
// mixed-result stream helpers use their chosen names in definitions and calls.
func TestGeneratedHTTPHelpersCompileWithPackageNameCollisions(t *testing.T) {
	root := httpHelperCollisionRoot(t, "Foo Bar", "First")
	reserve := func(plan *Plan) error {
		pkg, err := plan.Generation().ClaimPackage("generated.local/gen/http/foo_bar/server")
		if err != nil {
			return err
		}
		declarations := []*codegen.NameDeclaration{
			codegen.NewPreferredName(codegen.NameType, "appendFS", codegen.UnexportedName, httpHelperCollisionOrder("append-fs")),
			codegen.NewPreferredName(codegen.NameFunction, "appendPrefix", codegen.UnexportedName, httpHelperCollisionOrder("append-prefix")),
			codegen.NewPreferredName(codegen.NameType, "discardCreateServerStream", codegen.UnexportedName, httpHelperCollisionOrder("discard-stream")),
		}
		for _, declaration := range declarations {
			if err := pkg.DeclareName(declaration); err != nil {
				return err
			}
		}
		return nil
	}
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planServiceData, reserve, planTransportData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	protocolFiles, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, protocolFiles...)
	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	runGeneratedTests(t, dir)
}

// TestGeneratedJSONRPCPackagesCompileAcrossSharedPackageRoots checks that each
// generated client and server uses the names chosen for its own declarations
// when two designs write files into the same Go packages.
func TestGeneratedJSONRPCPackagesCompileAcrossSharedPackageRoots(t *testing.T) {
	tests := []struct {
		name string
		mode jsonRPCSharedPackageMode
	}{
		{name: "ordinary", mode: jsonRPCUnary},
		{name: "server sent events", mode: jsonRPCSSE},
		{name: "web socket", mode: jsonRPCWebSocket},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first := jsonRPCSharedPackageRoot(t, "Foo Bar", "First", test.mode)
			second := jsonRPCSharedPackageRoot(t, "Foo-Bar", "Second", test.mode)
			plan := mustTestPlan(t, "generated.local/gen", []eval.Root{first, second}, planTransportData)
			files, err := testServiceFiles(plan)
			require.NoError(t, err)
			for _, root := range []*expr.RootExpr{first, second} {
				jsonPlan := plan.jsonrpc[root]
				files = append(files, jsonPlan.ServerFiles()...)
				files = append(files, jsonPlan.ClientFiles()...)
				files = append(files, jsonPlan.ServerTypeFiles()...)
				files = append(files, jsonPlan.ClientTypeFiles()...)
				files = append(files, jsonPlan.PathFiles()...)
			}
			files, err = mergeFilesByPath(files)
			require.NoError(t, err)
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

// jsonRPCSharedPackageRoot returns one design whose service name controls the
// generated directory. typePrefix keeps its service types separate from the
// other design that writes into the same directory. Every design uses the Call
// method, so their stream types and constructors request the same Go names.
func jsonRPCSharedPackageRoot(t *testing.T, serviceName, typePrefix string, mode jsonRPCSharedPackageMode) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		payload := dsl.Type(typePrefix+"Payload", func() {
			dsl.Attribute("value", dsl.String)
		})
		result := dsl.Type(typePrefix+"Result", func() {
			dsl.Attribute("value", dsl.String)
		})
		dsl.Service(serviceName, func() {
			dsl.Method("Call", func() {
				switch mode {
				case jsonRPCUnary:
					dsl.Payload(payload)
					dsl.Result(result)
					dsl.JSONRPC(func() {})
				case jsonRPCSSE:
					dsl.Payload(payload)
					dsl.StreamingResult(result)
					dsl.JSONRPC(func() { dsl.ServerSentEvents() })
				case jsonRPCWebSocket:
					dsl.StreamingPayload(payload)
					dsl.StreamingResult(result)
					dsl.JSONRPC(func() {})
				default:
					panic("unknown JSON-RPC test mode")
				}
			})
		})
	})
}

// httpHelperCollisionRoot returns one design with mixed HTTP results and a
// mapped file. serviceName controls the output directory and typePrefix keeps
// the service types distinct when two designs share that directory.
func httpHelperCollisionRoot(t *testing.T, serviceName, typePrefix string) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		payload := dsl.Type(typePrefix+"Payload", func() {
			dsl.Attribute("value", dsl.String)
		})
		result := dsl.Type(typePrefix+"Result", func() {
			dsl.Attribute("value", dsl.String)
		})
		event := dsl.Type(typePrefix+"Event", func() {
			dsl.Attribute("value", dsl.String)
		})
		dsl.Service(serviceName, func() {
			dsl.Method("Create", func() {
				dsl.Payload(payload)
				dsl.Result(result)
				dsl.StreamingResult(event)
				dsl.HTTP(func() {
					dsl.POST("/create")
					dsl.ServerSentEvents()
				})
			})
			dsl.Files("/asset.json", "/embedded/file.json")
		})
	})
}
