// This file verifies that common example CLI entrypoints render without a
// second generated-module path input.
package example

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/service"
	dsl "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestExampleCLIFiles(t *testing.T) {
	cases := []struct {
		Name               string
		DSL                func()
		HasEndpointResults bool
		HasStreamResults   bool
	}{
		{"no-server", testdata.NoServerDSL, true, false},
		{"single-server-single-host", testdata.SingleServerSingleHostDSL, true, false},
		{"single-server-single-host-with-variables", testdata.SingleServerSingleHostWithVariablesDSL, true, false},
		{"single-server-multiple-hosts", testdata.SingleServerMultipleHostsDSL, true, false},
		{"single-server-multiple-hosts-with-variables", testdata.SingleServerMultipleHostsWithVariablesDSL, true, false},
		{"server-stream", serverStreamClientDSL, false, true},
		{"input-stream", inputStreamClientDSL, false, false},
		{"mixed-results", mixedResultClientDSL, true, false},
		{"files-with-grpc", filesWithGRPCClientDSL, true, false},
		{"https-only", httpsOnlyClientDSL, true, false},
		{"grpcs-only", grpcsOnlyClientDSL, true, false},
		{"https-before-http", httpsBeforeHTTPClientDSL, true, false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			generation, err := codegen.NewGeneration("goa.design/goa/example", []eval.Root{root})
			require.NoError(t, err)
			servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
			require.NoError(t, err)
			plan, err := NewPlan(generation, servicePlan)
			require.NoError(t, err)
			require.NoError(t, generation.Freeze())
			require.NoError(t, servicePlan.Link())
			rootData, ok := plan.Root(servicePlan)
			require.True(t, ok)
			fs := CLIFiles(rootData)
			require.Len(t, fs, 1)
			require.Greater(t, len(fs[0].SectionTemplates), 0)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			require.Equal(t, c.HasEndpointResults, strings.Contains(code, "func writeEndpointResult("))
			require.Equal(t, c.HasStreamResults, strings.Contains(code, "func writeStreamResults["))
			require.Equal(t, c.HasEndpointResults || c.HasStreamResults, strings.Contains(code, "func writeJSON("))
			golden := filepath.Join("testdata", "client-"+c.Name+".golden")
			compareOrUpdateGolden(t, code, golden)
		})
	}
}

func TestExampleCLIOmitsServerWithoutCallableMethods(t *testing.T) {
	plannedRoot := plannedExampleRoot(t, func() {
		dsl.Service("assets", func() {
			dsl.Files("/index.html", "index.html")
		})
	})

	require.True(t, plannedRoot.Servers[0].HasHTTP)
	require.NotNil(t, plannedRoot.Servers[0].DefaultTransport())
	require.Nil(t, plannedRoot.Servers[0].DefaultClientTransport())
	require.Empty(t, CLIFiles(plannedRoot))
}

func TestExampleCLIUsesOnlyTransportsWithCallableMethods(t *testing.T) {
	plannedRoot := plannedExampleRoot(t, filesWithGRPCClientDSL)

	server := plannedRoot.Servers[0]
	require.True(t, server.HasTransport(TransportHTTP))
	require.True(t, server.HasTransport(TransportGRPC))
	require.Equal(t, TransportHTTP, server.DefaultTransport().Type)
	require.Equal(t, Transport(TransportGRPC), server.DefaultClientTransport().Type)
	files := CLIFiles(plannedRoot)
	require.Len(t, files, 1)
	code := renderExampleSections(t, files[0])
	require.Contains(t, code, `case "grpc", "grpcs":`)
	require.NotContains(t, code, `case "http", "https":`)
	require.NotContains(t, code, "doHTTP(")
	require.Contains(t, code, "valid schemes: grpc")
}

// plannedExampleRoot evaluates design and returns the server data used by
// generated example programs.
func plannedExampleRoot(t *testing.T, design func()) *Root {
	t.Helper()
	root := codegen.RunDSL(t, design)
	generation, err := codegen.NewGeneration("example.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plan, err := NewPlan(generation, servicePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	plannedRoot, ok := plan.Root(servicePlan)
	require.True(t, ok)
	return plannedRoot
}

var filesWithGRPCClientDSL = func() {
	dsl.Service("assets", func() {
		dsl.Files("/index.html", "index.html")
		dsl.Method("read_status", func() {
			dsl.Result(dsl.String)
			dsl.GRPC(func() {})
		})
	})
}

var httpsOnlyClientDSL = func() {
	dsl.API("secure HTTP", func() {
		dsl.Server("secure", func() {
			dsl.Services("status")
			dsl.Host("local", func() {
				dsl.URI("https://localhost:8443")
			})
		})
	})
	dsl.Service("status", func() {
		dsl.Method("read", func() {
			dsl.Result(dsl.String)
			dsl.HTTP(func() {
				dsl.GET("/status")
			})
		})
	})
}

var grpcsOnlyClientDSL = func() {
	dsl.API("secure gRPC", func() {
		dsl.Server("secure", func() {
			dsl.Services("status")
			dsl.Host("local", func() {
				dsl.URI("grpcs://localhost:8443")
			})
		})
	})
	dsl.Service("status", func() {
		dsl.Method("read", func() {
			dsl.Result(dsl.String)
			dsl.GRPC(func() {})
		})
	})
}

var httpsBeforeHTTPClientDSL = func() {
	dsl.API("ordered HTTP", func() {
		dsl.Server("ordered", func() {
			dsl.Services("status")
			dsl.Host("public", func() {
				dsl.URI("https://api.example.com:9443/base/status")
				dsl.URI("http://localhost:8080")
			})
		})
	})
	dsl.Service("status", func() {
		dsl.Method("read", func() {
			dsl.Result(dsl.String)
			dsl.HTTP(func() {
				dsl.GET("/status")
			})
		})
	})
}

var serverStreamClientDSL = func() {
	dsl.Service("events", func() {
		dsl.Method("watch", func() {
			dsl.StreamingResult(dsl.String)
			dsl.GRPC(func() {})
		})
	})
}

var inputStreamClientDSL = func() {
	dsl.Service("events", func() {
		dsl.Method("upload", func() {
			dsl.StreamingPayload(dsl.String)
			dsl.Result(dsl.String)
			dsl.HTTP(func() {
				dsl.POST("/upload")
			})
			dsl.GRPC(func() {})
		})
	})
}

var mixedResultClientDSL = func() {
	dsl.Service("events", func() {
		dsl.Method("create", func() {
			dsl.Result(dsl.String)
			dsl.StreamingResult(dsl.Int)
			dsl.HTTP(func() {
				dsl.POST("/create")
				dsl.ServerSentEvents()
			})
		})
	})
}

func TestMixedClientRoutesJSONRPCCommandsFromPlannedEndpoints(t *testing.T) {
	root := codegen.RunDSL(t, mixedClientRoutingDSL)
	generation, err := codegen.NewGeneration("example.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plan, err := NewPlan(generation, servicePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	plannedRoot, ok := plan.Root(servicePlan)
	require.True(t, ok)

	files := CLIFiles(plannedRoot)
	require.Len(t, files, 1)
	code := renderExampleSections(t, files[0])
	require.NotContains(t, code, "strings.HasPrefix(err.Error()")
	require.Contains(t, code, `case "catalog":`)
	require.Contains(t, code, `case "watch":`)
	require.Contains(t, code, "err = doJSONRPC(context.Background(), scheme, host, timeout, debug, os.Stdout)")
	require.Contains(t, code, `usageCommands := []string{`)
	require.NotContains(t, code, "sort.Strings(usageCommands)")
	require.NotContains(t, code, "slices.Compact(usageCommands)")
	first := strings.Index(code, `"catalog read"`)
	second := strings.Index(code, `"catalog watch"`)
	require.GreaterOrEqual(t, first, 0)
	require.Greater(t, second, first)
}

func TestMixedClientRoutesCollidingServicesByPlannedPaths(t *testing.T) {
	root := codegen.RunDSL(t, mixedClientCollidingServicesDSL)
	generation, err := codegen.NewGeneration("example.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plan, err := NewPlan(generation, servicePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	plannedRoot, ok := plan.Root(servicePlan)
	require.True(t, ok)

	first := servicePlan.Services().Get("read_value").PathName
	second := servicePlan.Services().Get("read-value").PathName
	require.NotEqual(t, first, second)
	code := renderExampleSections(t, CLIFiles(plannedRoot)[0])
	require.Contains(t, code, `case "`+codegen.KebabCase(first)+`":`)
	require.Contains(t, code, `case "`+codegen.KebabCase(second)+`":`)
	require.Contains(t, code, `"`+codegen.KebabCase(first)+` read"`)
	require.Contains(t, code, `"`+codegen.KebabCase(first)+` watch"`)
	require.Contains(t, code, `"`+codegen.KebabCase(second)+` read"`)
	require.Contains(t, code, `"`+codegen.KebabCase(second)+` watch"`)
}

func TestClientHostVariableValidationEmitsFixedCases(t *testing.T) {
	root := codegen.RunDSL(t, testdata.SingleServerMultipleHostsWithVariablesDSL)
	generation, err := codegen.NewGeneration("example.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plan, err := NewPlan(generation, servicePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	plannedRoot, ok := plan.Root(servicePlan)
	require.True(t, ok)

	code := renderExampleSections(t, CLIFiles(plannedRoot)[0])
	require.Contains(t, code, `switch *versionF`)
	require.Contains(t, code, `case "v1", "v2":`)
	require.NotContains(t, code, "for _, v := range []string")
}

var mixedClientRoutingDSL = func() {
	dsl.API("mixed client", func() {
		dsl.Server("public", func() {
			dsl.Services("catalog")
			dsl.Host("development", func() {
				dsl.URI("http://localhost:8080")
			})
		})
	})
	dsl.Service("catalog", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("read", func() {
			dsl.HTTP(func() {
				dsl.GET("/catalog")
			})
		})
		dsl.Method("watch", func() {
			dsl.JSONRPC(func() {})
		})
	})
}

var mixedClientCollidingServicesDSL = func() {
	dsl.API("mixed client collisions", func() {
		dsl.Server("public", func() {
			dsl.Services("read_value", "read-value")
		})
	})
	for _, name := range []string{"read_value", "read-value"} {
		dsl.Service(name, func() {
			dsl.JSONRPC(func() {
				dsl.POST("/" + name)
			})
			dsl.Method("read", func() {
				dsl.HTTP(func() {
					dsl.GET("/" + name)
				})
			})
			dsl.Method("watch", func() {
				dsl.JSONRPC(func() {})
			})
		})
	}
}
