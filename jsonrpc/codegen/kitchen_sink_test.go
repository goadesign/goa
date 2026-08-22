// This file compiles a representative JSON-RPC design and compares every
// generated package with its checked-in golden contract.
package codegen_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	grpccodegen "goa.design/goa/v3/grpc/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
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
	generation, err := goacodegen.NewGeneration("generated.local/gen", roots)
	require.NoError(t, err)
	examples := expr.NewExampleGenerator(root.API.RandomizerFactory)
	servicePlan, err := service.NewPlan(root, generation, examples)
	require.NoError(t, err)
	require.NoError(t, jsonrpccodegen.Plan(generation))
	require.NoError(t, example.Plan(generation))
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	services := servicePlan.Services()
	tfiles := kitchenSinkTransportFiles(root, services)
	efiles := kitchenSinkExampleFiles(root, servicePlan)

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

// kitchenSinkTransportFiles assembles every transport file through the public
// subsystem APIs exercised by the golden fixture.
func kitchenSinkTransportFiles(root *expr.RootExpr, services *service.ServicesData) []*goacodegen.File {
	httpServices := httpcodegen.NewServicesData(services, root.API.HTTP)
	files := httpcodegen.ServerFiles(httpServices)
	files = append(files, httpcodegen.ClientFiles(httpServices)...)
	files = append(files, httpcodegen.ServerTypeFiles(httpServices)...)
	files = append(files, httpcodegen.ClientTypeFiles(httpServices)...)
	files = append(files, httpcodegen.PathFiles(httpServices)...)
	files = append(files, httpcodegen.ClientCLIFiles(httpServices)...)

	grpcServices := grpccodegen.NewServicesData(services)
	files = append(files, grpccodegen.ProtoFiles(grpcServices)...)
	files = append(files, grpccodegen.ServerFiles(grpcServices)...)
	files = append(files, grpccodegen.ClientFiles(grpcServices)...)
	files = append(files, grpccodegen.ServerTypeFiles(grpcServices)...)
	files = append(files, grpccodegen.ClientTypeFiles(grpcServices)...)
	files = append(files, grpccodegen.ClientCLIFiles(grpcServices)...)

	jsonrpcServices := httpcodegen.NewJSONRPCServicesData(services, &root.API.JSONRPC.HTTPExpr)
	files = append(files, jsonrpccodegen.ServerFiles(jsonrpcServices)...)
	files = append(files, jsonrpccodegen.ClientFiles(jsonrpcServices)...)
	files = append(files, httpcodegen.ServerTypeFiles(jsonrpcServices)...)
	files = append(files, httpcodegen.ClientTypeFiles(jsonrpcServices)...)
	files = append(files, httpcodegen.PathFiles(jsonrpcServices)...)
	return append(files, httpcodegen.ClientCLIFiles(jsonrpcServices)...)
}

// kitchenSinkExampleFiles assembles example service and transport files
// through their public subsystem APIs.
func kitchenSinkExampleFiles(root *expr.RootExpr, plan *service.Plan) []*goacodegen.File {
	services := plan.Services()
	files := service.ExampleServiceFiles(plan)
	files = append(files, service.ExampleInterceptorsFiles(plan)...)
	files = append(files, example.ServerFiles(root, services)...)
	files = append(files, example.CLIFiles(root)...)

	if len(root.API.HTTP.Services) > 0 {
		httpServices := httpcodegen.NewServicesData(services, root.API.HTTP)
		files = append(files, httpcodegen.ExampleServerFiles(httpServices)...)
		files = append(files, httpcodegen.ExampleCLIFiles(httpServices)...)
	}
	if len(root.API.JSONRPC.Services) > 0 {
		jsonrpcServices := httpcodegen.NewJSONRPCServicesData(services, &root.API.JSONRPC.HTTPExpr)
		files = append(files, jsonrpccodegen.ExampleServerFiles(jsonrpcServices, files)...)
		files = append(files, httpcodegen.ExampleCLIFiles(jsonrpcServices)...)
	}
	if len(root.API.GRPC.Services) > 0 {
		grpcServices := grpccodegen.NewServicesData(services)
		files = append(files, grpccodegen.ExampleServerFiles(grpcServices)...)
		files = append(files, grpccodegen.ExampleCLIFiles(grpcServices)...)
	}
	return files
}
