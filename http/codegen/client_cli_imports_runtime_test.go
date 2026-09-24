// These tests compile generated HTTP commands and execute them with map-query
// flags and custom field types. They check the imports used by the final Go
// source, including aliases chosen when a custom type shares an import name.
package codegen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestGeneratedMapQueryCLI(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.API("query", func() {})
		dsl.Service("query", func() {
			for _, name := range []string{"optional", "required", "defaulted", "validated"} {
				dsl.Method(name, func() {
					dsl.Payload(func() {
						dsl.Attribute("values", dsl.MapOf(dsl.String, dsl.String), func() {
							if name == "defaulted" {
								dsl.Default(map[string]string{"preset": "ok"})
							}
							if name == "validated" {
								dsl.Elem(func() {
									dsl.MinLength(2)
								})
							}
						})
						if name == "required" {
							dsl.Required("values")
						}
					})
					dsl.HTTP(func() {
						dsl.GET("/" + name)
						dsl.MapParams("values")
						dsl.Response(dsl.StatusNoContent)
					})
				})
			}
			dsl.Method("inline", func() {
				dsl.Payload(dsl.MapOf(dsl.String, dsl.String))
				dsl.HTTP(func() {
					dsl.GET("/inline")
					dsl.MapParams()
					dsl.Response(dsl.StatusNoContent)
				})
			})
		})
	})
	runGeneratedCLIImportTest(t, root, generatedMapQueryCLITest)
}

func TestGeneratedCLICustomTypeAlias(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("reports", func() {
			dsl.Method("create", func() {
				dsl.Payload(func() {
					dsl.Attribute("at", dsl.String, func() {
						dsl.Meta("struct:field:type", "json.Time", "time", "json")
					})
					dsl.Required("at")
				})
				dsl.HTTP(func() {
					dsl.POST("/reports")
					dsl.Body("at")
				})
			})
		})
	})
	runGeneratedCLIImportTest(t, root, generatedCustomTypeCLITest)
}

// runGeneratedCLIImportTest renders the service, HTTP client, and command
// packages together. The separate module compiles the generated imports and
// runs the supplied tests against this checkout's runtime.
func runGeneratedCLIImportTest(t *testing.T, root *expr.RootExpr, testSource string) {
	t.Helper()
	plan, generation, servicePlan := plannedHTTPPlan(t, root, false)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plan.Link())

	files, err := service.Files(servicePlan)
	require.NoError(t, err)
	files = append(files, plan.ClientFiles()...)
	files = append(files, plan.ClientTypeFiles()...)
	files = append(files, plan.PathFiles()...)
	files = append(files, plan.ClientCLIFiles()...)

	directory := t.TempDir()
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	module := "module generated.local\n\ngo 1.26\n\n" +
		"require goa.design/goa/v3 v3.0.0\n\n" +
		"replace goa.design/goa/v3 => " + filepath.ToSlash(repository) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte(module), 0o600))
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	require.NoError(t, os.WriteFile(filepath.Join(directory, "cli_test.go"), []byte(testSource), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./...")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated CLI tests:\n%s", output)
}

const generatedMapQueryCLITest = `package cli_test

import (
	"flag"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	gencli "generated.local/gen/http/cli/query"
	goahttp "goa.design/goa/v3/http"
)

type captureDoer struct {
	request *http.Request
}

func (d *captureDoer) Do(request *http.Request) (*http.Response, error) {
	d.request = request
	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header: make(http.Header),
		Body: io.NopCloser(strings.NewReader("")),
	}, nil
}

func TestMapQueryFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want url.Values
		wantError string
	}{
		{name: "optional omitted", args: []string{"optional"}, want: url.Values{}},
		{name: "optional empty", args: []string{"optional", "--values", "{}"}, want: url.Values{}},
		{name: "optional supplied", args: []string{"optional", "--values", "{\"a\":\"hello world\"}"}, want: url.Values{"a": {"hello world"}}},
		{name: "malformed JSON", args: []string{"optional", "--values", "{"}, wantError: "invalid JSON"},
		{name: "required omitted", args: []string{"required"}, wantError: "missing required flag"},
		{name: "required empty", args: []string{"required", "--values", "{}"}, want: url.Values{}},
		{name: "default omitted", args: []string{"defaulted"}, want: url.Values{"preset": {"ok"}}},
		{name: "default overridden", args: []string{"defaulted", "--values", "{}"}, want: url.Values{}},
		{name: "valid Unicode length", args: []string{"validated", "--values", "{\"a\":\"éé\"}"}, want: url.Values{"a": {"éé"}}},
		{name: "invalid Unicode length", args: []string{"validated", "--values", "{\"a\":\"é\"}"}, wantError: "length"},
		{name: "whole inline map", args: []string{"inline", "--p", "{\"a\":\"inline\"}"}, want: url.Values{"a": {"inline"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			savedArgs, savedFlags := os.Args, flag.CommandLine
			t.Cleanup(func() {
				os.Args, flag.CommandLine = savedArgs, savedFlags
			})
			os.Args = append([]string{"query-cli", "query"}, test.args...)
			flag.CommandLine = flag.NewFlagSet("query-cli", flag.ContinueOnError)
			doer := &captureDoer{}
			endpoint, payload, err := gencli.ParseEndpoint(
				"http", "example.com", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false,
			)
			if test.wantError != "" {
				require.ErrorContains(t, err, test.wantError)
				require.Nil(t, doer.request)
				return
			}
			require.NoError(t, err)
			_, err = endpoint(t.Context(), payload)
			require.NoError(t, err)
			require.NotNil(t, doer.request)
			require.Equal(t, "/"+test.args[0], doer.request.URL.Path)
			require.Equal(t, test.want, doer.request.URL.Query())
		})
	}
}
`

const generatedCustomTypeCLITest = `package cli_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	genclient "generated.local/gen/http/reports/client"
)

func TestCustomTypeFlag(t *testing.T) {
	text := "\"2026-01-02T03:04:05Z\""
	payload, err := genclient.BuildCreatePayload(&text)
	require.NoError(t, err)
	require.Equal(t, time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC), payload.At)
}
`
