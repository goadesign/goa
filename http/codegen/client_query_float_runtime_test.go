// This file runs generated HTTP client query encoders and checks the exact
// values that a server receives after URL parsing.
package codegen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestGeneratedClientFormatsFloatQueriesCompactly catches fixed-point query
// formatting that expands values which have a shorter exponent form.
func TestGeneratedClientFormatsFloatQueriesCompactly(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("float_query", func() {
			dsl.Method("format", func() {
				dsl.Payload(func() {
					dsl.Attribute("scalar32", dsl.Float32)
					dsl.Attribute("scalar64", dsl.Float64)
					dsl.Attribute("repeated32", dsl.ArrayOf(dsl.Float32))
					dsl.Attribute("repeated64", dsl.ArrayOf(dsl.Float64))
					dsl.Required("scalar32", "scalar64", "repeated32", "repeated64")
				})
				dsl.HTTP(func() {
					dsl.GET("/")
					dsl.Param("scalar32")
					dsl.Param("scalar64")
					dsl.Param("repeated32")
					dsl.Param("repeated64")
				})
			})
		})
	})

	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
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
	files := append(serviceFiles, httpPlans[0].ClientFiles()...)
	files = append(files, httpPlans[0].ClientTypeFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)
	runGeneratedFloatQueryTest(t, files)
}

// runGeneratedFloatQueryTest writes the generated packages and a test in the
// generated client package, then runs that test in an isolated module.
func runGeneratedFloatQueryTest(t *testing.T, files []*codegen.File) {
	t.Helper()
	directory := t.TempDir()
	goaRoot := floatQueryModuleDirectory(t)
	module := "module generated.local\n\ngo 1.24\n\n" +
		"require goa.design/goa/v3 v3.0.0\n\n" +
		"replace goa.design/goa/v3 => " + filepath.ToSlash(goaRoot) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte(module), 0o600))
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}

	testPath := filepath.Join(directory, "gen", "http", "float_query", "client", "float_query_test.go")
	require.NoError(t, os.WriteFile(testPath, []byte(generatedFloatQueryTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./gen/http/float_query/client")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated float query test:\n%s", output)
}

// floatQueryModuleDirectory returns this Goa checkout so the temporary module
// tests the generated code against the same runtime as the generator.
func floatQueryModuleDirectory(t *testing.T) string {
	t.Helper()
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "goa.design/goa/v3")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "resolve Goa module:\n%s", output)
	directory := strings.TrimSpace(string(output))
	require.NotEmpty(t, directory)
	return directory
}

const generatedFloatQueryTest = `package client

import (
	"net/http"
	"testing"

	genfloatquery "generated.local/gen/float_query"
)

func TestFloatQueryValues(t *testing.T) {
	payload := &genfloatquery.FormatPayload{
		Scalar32:   12.5,
		Scalar64:   1e100,
		Repeated32: []float32{1e20, 0.25},
		Repeated64: []float64{1e100, 0.25},
	}
	request, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := EncodeFormatRequest(nil)(request, payload); err != nil {
		t.Fatal(err)
	}
	query := request.URL.Query()
	if got := query.Get("scalar32"); got != "12.5" {
		t.Fatalf("scalar32 query = %q, want %q", got, "12.5")
	}
	if got := query.Get("scalar64"); got != "1e+100" {
		t.Fatalf("scalar64 query = %q, want %q", got, "1e+100")
	}
	if got := query["repeated32"]; len(got) != 2 || got[0] != "1e+20" || got[1] != "0.25" {
		t.Fatalf("repeated32 query = %#v, want %#v", got, []string{"1e+20", "0.25"})
	}
	if got := query["repeated64"]; len(got) != 2 || got[0] != "1e+100" || got[1] != "0.25" {
		t.Fatalf("repeated64 query = %#v, want %#v", got, []string{"1e+100", "0.25"})
	}
}
`
