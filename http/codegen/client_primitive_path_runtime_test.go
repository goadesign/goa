// This file renders an HTTP client for a primitive path payload. The generated
// test checks that the request builder validates and copies the payload into the
// URL instead of silently using the primitive type's zero value.
package codegen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestGeneratedPrimitivePathRequestBuilder checks the public behavior of a
// generated request builder for both a valid string and a value of the wrong
// type.
func TestGeneratedPrimitivePathRequestBuilder(t *testing.T) {
	root := expr.RunDSL(t, primitivePathDSL)
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())

	serviceFiles, err := service.Files(servicePlan)
	require.NoError(t, err)
	files := slices.Clone(serviceFiles)
	files = append(files, httpPlans[0].ClientFiles()...)
	files = append(files, httpPlans[0].ClientTypeFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)
	runPrimitivePathRuntimeTest(t, files)
}

// primitivePathDSL defines a string payload used directly as a path value.
func primitivePathDSL() {
	dsl.Service("primitive_path", func() {
		dsl.Method("download", func() {
			dsl.Payload(dsl.String)
			dsl.HTTP(func() {
				dsl.GET("/files/{filename}")
				dsl.Param("filename")
			})
		})
	})
}

// runPrimitivePathRuntimeTest writes the generated client and runs its test in
// an isolated module that uses this Goa checkout.
func runPrimitivePathRuntimeTest(t *testing.T, files []*goacodegen.File) {
	t.Helper()
	directory := t.TempDir()
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	module := fmt.Sprintf(
		"module generated.local\n\ngo 1.24\n\nrequire goa.design/goa/v3 v3.0.0\n\nreplace goa.design/goa/v3 => %s\n",
		filepath.ToSlash(repository),
	)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte(module), 0o600))
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	testPath := filepath.Join(directory, "gen", "http", "primitive_path", "client", "primitive_path_test.go")
	require.NoError(t, os.WriteFile(testPath, []byte(generatedPrimitivePathTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./gen/http/primitive_path/client")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated primitive path test:\n%s", output)
}

const generatedPrimitivePathTest = `package client

import (
	"context"
	"errors"
	"testing"

	goahttp "goa.design/goa/v3/http"
)

func TestPrimitivePathRequest(t *testing.T) {
	client := NewClient("https", "example.com", nil, nil, nil, false)
	request, err := client.BuildDownloadRequest(context.Background(), "report.csv")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := request.URL.Path, "/files/report.csv"; got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}

	_, err = client.BuildDownloadRequest(context.Background(), 42)
	var clientError *goahttp.ClientError
	if !errors.As(err, &clientError) {
		t.Fatalf("error = %v, want *goahttp.ClientError", err)
	}
	if got, want := clientError.Name, "invalid_type"; got != want {
		t.Fatalf("error name = %q, want %q", got, want)
	}
}
`
