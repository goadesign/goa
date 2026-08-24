// This file compiles a representative JSON-RPC design and compares every
// generated package with its checked-in golden contract.
package codegen_test

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

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
// required and optional IDs, custom errors, SSE, and mixed HTTP and JSON-RPC
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
	examplePlan, err := example.NewPlan(generation, servicePlan)
	require.NoError(t, err)
	httpExamples, err := httpcodegen.NewExamplePlan(httpPlans[0], examplePlan)
	require.NoError(t, err)
	jsonExamples, err := jsonrpccodegen.NewExamplePlan(jsonPlans[0], examplePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	require.NoError(t, jsonHTTPPlans[0].Link())
	require.NoError(t, jsonPlans[0].Link())
	tfiles := kitchenSinkTransportFiles(httpPlans[0], jsonPlans[0])
	rootData, ok := examplePlan.Root(servicePlan)
	require.True(t, ok)
	efiles := kitchenSinkExampleFiles(rootData, servicePlan, httpExamples, jsonExamples)

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
	feedCodec, err := os.ReadFile(filepath.Join(tmp, "gen", "jsonrpc", "feed", "client", "encode_decode.go"))
	require.NoError(t, err)
	require.NotContains(t, string(feedCodec), "DecodeWatchResponse")
	require.Contains(t, string(feedCodec), "DecodeSnapshotResponse")
	compileKitchenSink(t, servicePlan, tfiles)
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
func kitchenSinkExampleFiles(root *example.Root, plan *service.Plan, httpPlan *httpcodegen.ExamplePlan, jsonPlan *jsonrpccodegen.ExamplePlan) []*goacodegen.File {
	services := plan.Services()
	files := service.ExampleServiceFiles(plan)
	files = append(files, service.ExampleInterceptorsFiles(plan)...)
	files = append(files, example.ServerFiles(root, services)...)
	files = append(files, example.CLIFiles(root)...)

	files = append(files, httpPlan.CLIFiles()...)
	files = append(files, jsonPlan.ServerFiles()...)
	files = append(files, jsonPlan.CLIFiles()...)
	return files
}

// compileKitchenSink renders the service and transport packages together so
// every generated conversion must use a concrete value accepted by the
// service contract it returns.
func compileKitchenSink(t *testing.T, servicePlan *service.Plan, transportFiles []*goacodegen.File) {
	t.Helper()
	serviceFiles, err := service.Files(servicePlan)
	require.NoError(t, err)
	dir := t.TempDir()
	for _, file := range append(serviceFiles, transportFiles...) {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}

	goaDir := goaModuleDirectory(t)
	module := "module generated.local\n\ngo 1.25\n\nrequire goa.design/goa/v3 v3.0.0\n\n" +
		"replace goa.design/goa/v3 => " + filepath.ToSlash(goaDir) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(module), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./gen/...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

// goaModuleDirectory returns the local Goa checkout used to build this test.
func goaModuleDirectory(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "goa.design/goa/v3")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	dir := strings.TrimSpace(string(output))
	require.NotEmpty(t, dir)
	return dir
}
