// This file verifies that generated example servers and command-line programs
// contain the service and transport wiring required by representative designs.
package example

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/service"
	dsl "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// updateGolden is true when -w is passed to `go test`, e.g. `go test ./... -w`
var updateGolden = false

func init() {
	flag.BoolVar(&updateGolden, "w", false, "update golden files")
}

func compareOrUpdateGolden(t *testing.T, code, golden string) {
	t.Helper()
	if updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0750))
		require.NoError(t, os.WriteFile(golden, []byte(code), 0640))
		return
	}
	data, err := os.ReadFile(golden)
	require.NoError(t, err)
	if runtime.GOOS == "windows" {
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	}
	assert.Equal(t, string(data), code)
}

func TestExampleServerFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"no-server", testdata.NoServerDSL},
		{"same-api-service-name", testdata.SameAPIServiceNameDSL},
		{"single-server-single-host", testdata.SingleServerSingleHostDSL},
		{"single-server-single-host-with-variables", testdata.SingleServerSingleHostWithVariablesDSL},
		{"server-hosting-service-with-file-server", testdata.ServerHostingServiceWithFileServerDSL},
		{"server-hosting-service-subset", testdata.ServerHostingServiceSubsetDSL},
		{"server-hosting-multiple-services", testdata.ServerHostingMultipleServicesDSL},
		{"single-server-multiple-hosts", testdata.SingleServerMultipleHostsDSL},
		{"single-server-multiple-hosts-with-variables", testdata.SingleServerMultipleHostsWithVariablesDSL},
		{"service-name-with-spaces", testdata.NamesWithSpacesDSL},
		{"service-for-only-http", testdata.ServiceForOnlyHTTPDSL},
		{"sercice-for-only-grpc", testdata.ServiceForOnlyGRPCDSL},
		{"service-for-http-and-part-of-grpc", testdata.ServiceForHTTPAndPartOfGRPCDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			generation, err := codegen.NewGeneration("goa.design/goa/example", []eval.Root{root})
			require.NoError(t, err)
			servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
			require.NoError(t, err)
			examplePlan, err := NewPlan(generation, servicePlan)
			require.NoError(t, err)
			rootData, ok := examplePlan.Root(servicePlan)
			require.True(t, ok)
			require.NoError(t, generation.Freeze())
			require.NoError(t, servicePlan.Link())
			services := servicePlan.Services()
			fs := ServerFiles(rootData, services)
			require.Len(t, fs, 1)
			require.Greater(t, len(fs[0].SectionTemplates), 0)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			golden := filepath.Join("testdata", "server-"+c.Name+".golden")
			compareOrUpdateGolden(t, code, golden)
		})
	}
}

func TestGRPCOnlyServerLoggerDoesNotUseHTTPPort(t *testing.T) {
	root := codegen.RunDSL(t, testdata.ServiceForOnlyGRPCDSL)
	generation, err := codegen.NewGeneration("goa.design/goa/example", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	examplePlan, err := NewPlan(generation, servicePlan)
	require.NoError(t, err)
	rootData, ok := examplePlan.Root(servicePlan)
	require.True(t, ok)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	files := ServerFiles(rootData, servicePlan.Services())
	require.Len(t, files, 1)

	var code bytes.Buffer
	for _, section := range files[0].SectionTemplates {
		require.NoError(t, section.Write(&code))
	}
	require.NotContains(t, code.String(), "httpPortF")
	require.Contains(t, code.String(), `log.KV{K: "grpc-port", V: *grpcPortF}`)
}

func TestServerMainUsesPlannedDeclarationsAndDistinctLocals(t *testing.T) {
	root := codegen.RunDSL(t, collidingServerServicesDSL)
	generation, err := codegen.NewGeneration("example.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	examplePlan, err := NewPlan(generation, servicePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	plannedRoot, ok := examplePlan.Root(servicePlan)
	require.True(t, ok)
	services := servicePlan.Services()
	files := ServerFiles(plannedRoot, services)
	require.Len(t, files, 1)
	code := renderExampleSections(t, files[0])

	first := services.Get("foo-bar")
	second := services.Get("foo bar")
	outputPackage := plannedRoot.Servers[0].serverPackage.ImportPath()
	serviceImport := services.ServiceImport(outputPackage, first.Name)
	require.Contains(t, code, serviceImport.Name+` "`+serviceImport.Path+`"`)
	require.Contains(t, code, serviceImport.Name+"."+first.ServiceDeclaration.Name())
	require.Contains(t, code, serviceImport.Name+"."+second.ServiceDeclaration.Name())
	require.Contains(t, code, "."+first.ExampleConstructorDeclaration.Name()+"()")
	require.Contains(t, code, "."+second.ExampleConstructorDeclaration.Name()+"()")
	require.Contains(t, code, "."+first.NewEndpointsDeclaration.Name()+"(")
	require.Contains(t, code, "."+second.NewEndpointsDeclaration.Name()+"(")
	require.Contains(t, code, "."+first.ServerInterceptorsDeclaration.Name())
	require.Contains(t, code, "."+second.ServerInterceptorsDeclaration.Name())
	require.Contains(t, code, "."+first.ExampleServerInterceptorsConstructorDeclaration.Name()+"()")
	require.Contains(t, code, "."+second.ExampleServerInterceptorsConstructorDeclaration.Name()+"()")
	require.Contains(t, code, "fooBarSvc2")
	require.Contains(t, code, "fooBarEndpoints2")
	require.Contains(t, code, "fooBarInterceptors2")
}

// TestServerMainUsesItsOutputPackageImportNames checks that clue/log keeps the
// name log in cmd/public/main.go. The application receives log2, and every
// generated call uses that selected name.
func TestServerMainUsesItsOutputPackageImportNames(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.API("log", func() {
			dsl.Server("public", func() {
				dsl.Services("status")
				dsl.Host("development", func() {
					dsl.URI("http://localhost:8080")
				})
			})
		})
		dsl.Service("status", func() {
			dsl.Method("read", func() {
				dsl.HTTP(func() {
					dsl.GET("/status")
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("example.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	examplePlan, err := NewPlan(generation, servicePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	plannedRoot, ok := examplePlan.Root(servicePlan)
	require.True(t, ok)

	files := ServerFiles(plannedRoot, servicePlan.Services())
	require.Len(t, files, 1)
	code := renderExampleSections(t, files[0])
	require.Contains(t, code, `log2 "example.local"`)
	require.Contains(t, code, `"goa.design/clue/log"`)
	require.Contains(t, code, "log2.NewStatus()")
}

var collidingServerServicesDSL = func() {
	trace := dsl.Interceptor("trace")
	dsl.API("colliding server", func() {
		dsl.Server("public", func() {
			dsl.Services("foo-bar", "foo bar")
			dsl.Host("development", func() {
				dsl.URI("http://localhost:8080")
			})
		})
	})
	dsl.Service("foo-bar", func() {
		dsl.ServerInterceptor(trace)
		dsl.Method("first", func() {
			dsl.HTTP(func() {
				dsl.GET("/first")
			})
		})
	})
	dsl.Service("foo bar", func() {
		dsl.ServerInterceptor(trace)
		dsl.Method("second", func() {
			dsl.HTTP(func() {
				dsl.GET("/second")
			})
		})
	})
}
