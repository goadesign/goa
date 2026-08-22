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
	httpPlans, err := httpcodegen.NewPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	jsonHTTPPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	jsonPlans, err := jsonrpccodegen.NewPlans(generation, jsonrpccodegen.PlanInput{
		Root: root, Service: servicePlan, HTTP: jsonHTTPPlans[0], ApplicationHTTP: httpPlans[0],
	})
	require.NoError(t, err)
	require.NoError(t, example.Plan(generation))
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	require.NoError(t, jsonHTTPPlans[0].Link())
	require.NoError(t, jsonPlans[0].Link())
	tfiles := kitchenSinkTransportFiles(httpPlans[0], jsonPlans[0])
	efiles := kitchenSinkExampleFiles(root, servicePlan, httpPlans[0], jsonPlans[0])

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
func kitchenSinkTransportFiles(httpPlan *httpcodegen.Plan, jsonPlan *jsonrpccodegen.Plan) []*goacodegen.File {
	files := httpPlan.ServerFiles()
	files = append(files, httpPlan.ClientFiles()...)
	files = append(files, httpPlan.ServerTypeFiles()...)
	files = append(files, httpPlan.ClientTypeFiles()...)
	files = append(files, httpPlan.PathFiles()...)
	files = append(files, httpPlan.ClientCLIFiles()...)

	files = append(files, jsonPlan.ServerFiles()...)
	files = append(files, jsonPlan.ClientFiles()...)
	files = append(files, jsonPlan.ServerTypeFiles()...)
	files = append(files, jsonPlan.ClientTypeFiles()...)
	files = append(files, jsonPlan.PathFiles()...)
	return append(files, jsonPlan.ClientCLIFiles()...)
}

// kitchenSinkExampleFiles assembles example service and transport files
// through their public subsystem APIs.
func kitchenSinkExampleFiles(root *expr.RootExpr, plan *service.Plan, httpPlan *httpcodegen.Plan, jsonPlan *jsonrpccodegen.Plan) []*goacodegen.File {
	services := plan.Services()
	files := service.ExampleServiceFiles(plan)
	files = append(files, service.ExampleInterceptorsFiles(plan)...)
	files = append(files, example.ServerFiles(root, services)...)
	files = append(files, example.CLIFiles(root)...)

	if len(root.API.HTTP.Services) > 0 {
		files = append(files, httpPlan.ExampleCLIFiles()...)
	}
	if len(root.API.JSONRPC.Services) > 0 {
		files = append(files, jsonPlan.ExampleServerFiles()...)
		files = append(files, jsonPlan.ExampleCLIFiles()...)
	}
	return files
}
