// This file verifies that HTTP package names are requested before Goa assigns
// them and that files are built only after service names are available.
package codegen

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestPlanDoesNotReserveUnusedStaticAliasesBeforeFreeze(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Path", func() {
			dsl.Method("Read", func() { dsl.HTTP(func() { dsl.GET("/") }) })
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.Len(t, plans, 1)
	for filePath, file := range plans[0].fileImports {
		if strings.HasPrefix(filePath, "gen/http/path/client/") {
			require.NotContains(t, file.Paths(), "generated.local/gen/path")
		}
	}
}

func TestPlanRejectsFrozenGeneration(t *testing.T) {
	generation, err := codegen.NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())

	_, err = NewPlans(generation)
	require.Error(t, err)
}

func TestEndpointPayloadConstructorUsesReleasedTypeName(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		payload string
		want    string
	}{
		{
			name:    "named payload",
			method:  "MethodBodyUnion",
			payload: "Union",
			want:    "NewMethodBodyUnionUnion",
		},
		{
			name:    "overlapping payload",
			method:  "MethodA",
			payload: "APayload",
			want:    "NewMethodAAPayload",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			endpoint := &expr.HTTPEndpointExpr{MethodExpr: &expr.MethodExpr{
				Name:    test.method,
				Payload: &expr.AttributeExpr{Type: wireCatalogType(test.payload, test.payload, "value", true)},
			}}
			require.Equal(t, test.want, endpointPayloadConstructorName(endpoint))
		})
	}
}

// TestNewExamplePlanRejectsAnotherServicePlan checks that server names and
// URLs cannot come from a different design with the same authored names.
func TestNewExamplePlanRejectsAnotherServicePlan(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("Service", func() {
			dsl.Method("Read", func() { dsl.HTTP(func() { dsl.GET("/") }) })
		})
	})
	transport := linkedHTTPPlanForRoot(t, root)

	otherRoot := codegen.RunDSL(t, func() {
		dsl.Service("Service", func() {
			dsl.Method("Read", func() { dsl.HTTP(func() { dsl.GET("/") }) })
		})
	})
	otherGeneration, err := codegen.NewGeneration("other.local/gen", []eval.Root{otherRoot})
	require.NoError(t, err)
	otherService, err := service.NewPlan(otherRoot, otherGeneration, expr.NewExampleGenerator(otherRoot.API.RandomizerFactory))
	require.NoError(t, err)
	examples, err := example.NewPlan(otherGeneration, otherService)
	require.NoError(t, err)

	_, err = NewExamplePlan(transport, examples)
	require.EqualError(t, err, "HTTP examples require server data created from the same service design")
}

// TestNewPlansRequiresEveryHTTPRoot proves package names cannot be requested
// from only some of the HTTP designs in one generation.
func TestNewPlansRequiresEveryHTTPRoot(t *testing.T) {
	first := expr.RunDSL(t, func() {
		dsl.Service("First", func() {
			dsl.Method("Read", func() { dsl.HTTP(func() { dsl.GET("/first") }) })
		})
	})
	second := expr.RunDSL(t, func() {
		dsl.Service("Second", func() {
			dsl.Method("Read", func() { dsl.HTTP(func() { dsl.GET("/second") }) })
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{first, second})
	require.NoError(t, err)
	services, err := service.NewPlans(generation,
		service.PlanInput{Root: first, Examples: expr.NewExampleGenerator(first.API.RandomizerFactory)},
		service.PlanInput{Root: second, Examples: expr.NewExampleGenerator(second.API.RandomizerFactory)},
	)
	require.NoError(t, err)

	_, err = NewPlans(generation, PlanInput{Root: first, Service: services[0]})
	require.EqualError(t, err, "HTTP planning requires all 2 transport roots, got 1")
}

// TestNewPlansRejectsDuplicateRoot proves one service plan cannot be paired
// with the same design twice in one call.
func TestNewPlansRejectsDuplicateRoot(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			dsl.Method("Read", func() { dsl.HTTP(func() { dsl.GET("/calc") }) })
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)

	_, err = NewPlans(generation,
		PlanInput{Root: root, Service: servicePlan},
		PlanInput{Root: root, Service: servicePlan},
	)
	require.EqualError(t, err, fmt.Sprintf("HTTP root %p is planned more than once", root))
}

// TestPlanReservesGeneratedHTTPPackages verifies that client, server, and CLI
// packages receive distinct import names before files are written.
func TestPlanReservesGeneratedHTTPPackages(t *testing.T) {
	root := expr.RunDSL(t, func() {
		for _, name := range []string{"Foo", "Fooc", "Foosvr"} {
			dsl.Service(name, func() {
				dsl.Method("Read", func() {
					dsl.HTTP(func() { dsl.GET("/" + name) })
				})
			})
		}
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	examples, err := example.NewPlan(generation, servicePlan)
	require.NoError(t, err)
	_, err = NewExamplePlan(plans[0], examples)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	services := servicePlan.Services()

	cliOutput := path.Join(
		"generated.local/gen/http/cli",
		codegen.SnakeCase(codegen.Goify(root.API.Servers[0].Name, true)),
	)
	client := services.PackageImport(cliOutput, "generated.local/gen/http/foo/client")
	serverOutput := path.Join("generated.local", "cmd", codegen.SnakeCase(codegen.Goify(root.API.Servers[0].Name, true)))
	server := services.PackageImport(serverOutput, "generated.local/gen/http/foo/server")
	require.Equal(t, "fooc", client.Name)
	require.Equal(t, "foosvr2", server.Name)
}

// TestPlanLinkEagerlyRetainsHTTPFiles proves Link analyzes every HTTP service
// once and decides which generated files exist before callers request them.
func TestPlanLinkEagerlyRetainsHTTPFiles(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			dsl.Method("Add", func() {
				dsl.Payload(func() { dsl.Attribute("value", dsl.Int) })
				dsl.Result(func() { dsl.Attribute("total", dsl.Int) })
				dsl.HTTP(func() { dsl.POST("/add") })
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	examplePlan, err := example.NewPlan(generation, servicePlan)
	require.NoError(t, err)
	examples, err := NewExamplePlan(plans[0], examplePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	plan := plans[0]
	require.NoError(t, plan.Link())
	_, ok := plan.JSONRPCService("Calc")
	require.True(t, ok)
	require.NotEmpty(t, examples.ServerFiles())
	require.NotEmpty(t, examples.CLIFiles())
	serverCount := len(plan.ServerFiles())
	clientCount := len(plan.ClientFiles())

	root.API.HTTP.Services = append(root.API.HTTP.Services, &expr.HTTPServiceExpr{})
	require.Len(t, plan.ServerFiles(), serverCount)
	require.Len(t, plan.ClientFiles(), clientCount)
}

// TestJSONRPCCodecFilesAreIndependent checks that the JSON-RPC file writer can
// change a returned encoder and decoder file without changing a later copy.
func TestJSONRPCCodecFilesAreIndependent(t *testing.T) {
	root := expr.RunDSL(t, func() {
		value := dsl.Type("Value", func() {
			dsl.Meta("struct:pkg:path", "example.com/types")
			dsl.Attribute("number", dsl.Int)
		})
		dsl.Service("Calc", func() {
			dsl.Method("Add", func() {
				dsl.Payload(value)
				dsl.Result(func() { dsl.Attribute("total", dsl.Int) })
				dsl.JSONRPC(func() {})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewJSONRPCPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	service, ok := plans[0].JSONRPCService("Calc")
	require.True(t, ok)
	serverBody := service.Endpoints[0].Payload.Request.ServerBody
	require.NotNil(t, serverBody.Declaration)
	require.Equal(t, serverBody.Declaration.Name(), serverBody.VarName)
	stored := plans[0].jsonServices["Calc"]
	assertIndependentCodecFile(t, stored.clientCodec, service.ClientCodecFile)
	assertIndependentCodecFile(t, stored.serverCodec, service.ServerCodecFile)

	clientPath := stored.clientCodec.Path
	imports := service.FileImports(clientPath)
	require.NotEmpty(t, imports)
	require.Equal(t, imports, service.FileImports(strings.ReplaceAll(clientPath, "/", `\`)))
	original := *imports[0]
	imports[0].Path = "changed.example/package"
	freshImports := service.FileImports(clientPath)
	require.Equal(t, original, *freshImports[0])
	require.PanicsWithValue(t, "JSON-RPC file is not part of this HTTP service plan", func() {
		service.FileImports("gen/jsonrpc/calc/client/unknown.go")
	})

	service.Service.Name = "changed"
	service.Endpoints[0].ServiceName = "changed"
	service.Endpoints[0].Payload.Request.Headers = append(
		service.Endpoints[0].Payload.Request.Headers,
		JSONRPCHeaderData{CanonicalName: "Changed"},
	)
	fresh, ok := plans[0].JSONRPCService("Calc")
	require.True(t, ok)
	require.Equal(t, "Calc", fresh.Service.Name)
	require.Equal(t, "Calc", fresh.Endpoints[0].ServiceName)
	require.Empty(t, fresh.Endpoints[0].Payload.Request.Headers)
}

// TestPlanRetainsAttributeImportsBeforeFreeze proves ordinary HTTP and
// JSON-RPC files use the type packages recorded during planning, even if the
// design expression is changed before linking.
func TestPlanRetainsAttributeImportsBeforeFreeze(t *testing.T) {
	for _, transport := range []struct {
		name string
		plan func(*codegen.Generation, PlanInput) ([]*Plan, error)
		file func(*Plan) []*codegen.ImportSpec
	}{
		{
			name: "HTTP",
			plan: func(generation *codegen.Generation, input PlanInput) ([]*Plan, error) {
				return NewPlans(generation, input)
			},
			file: func(plan *Plan) []*codegen.ImportSpec {
				for _, file := range plan.ClientFiles() {
					if strings.HasSuffix(filepath.ToSlash(file.Path), "/client/encode_decode.go") {
						return file.SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec)
					}
				}
				return nil
			},
		},
		{
			name: "JSON-RPC",
			plan: func(generation *codegen.Generation, input PlanInput) ([]*Plan, error) {
				return NewJSONRPCPlans(generation, input)
			},
			file: func(plan *Plan) []*codegen.ImportSpec {
				service, ok := plan.JSONRPCService("Calc")
				require.True(t, ok)
				return service.FileImports("gen/jsonrpc/calc/client/encode_decode.go")
			},
		},
	} {
		t.Run(transport.name, func(t *testing.T) {
			root := expr.RunDSL(t, func() {
				dsl.Service("Calc", func() {
					dsl.Method("Add", func() {
						dsl.Payload(func() {
							dsl.Attribute("number", dsl.Int, func() {
								dsl.Meta("struct:field:type", "values.Number", "example.com/values", "values")
							})
						})
						dsl.Result(dsl.Int)
						if transport.name == "HTTP" {
							dsl.HTTP(func() { dsl.POST("/add") })
						} else {
							dsl.JSONRPC(func() {})
						}
					})
				})
			})
			generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
			require.NoError(t, err)
			servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
			require.NoError(t, err)
			plans, err := transport.plan(generation, PlanInput{Root: root, Service: servicePlan})
			require.NoError(t, err)

			payload := root.Service("Calc").Method("Add").Payload
			delete(expr.AsObject(payload.Type).Attribute("number").Meta, "struct:field:type")
			require.NoError(t, generation.Freeze())
			require.NoError(t, servicePlan.Link())
			require.NoError(t, plans[0].Link())

			imports := transport.file(plans[0])
			require.Contains(t, importPaths(imports), "example.com/values")
		})
	}
}

// TestHTTPPlanReservesOnlyFixedImportsWrittenByThePackage verifies that a
// package name stays available to user types when no generated HTTP file uses
// the standard library package with that name.
func TestHTTPPlanReservesOnlyFixedImportsWrittenByThePackage(t *testing.T) {
	const customRoot = "example.com/custom/"
	root := expr.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			dsl.Method("Read", func() {
				dsl.Result(func() {
					for _, name := range []string{"errors", "strings", "strconv"} {
						dsl.Attribute(name, dsl.String, func() {
							dsl.Meta("struct:field:type", name+".Value", customRoot+name, name)
						})
					}
				})
				dsl.HTTP(func() { dsl.GET("/") })
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	server := generation.Package("generated.local/gen/http/calc/server")
	for _, name := range []string{"errors", "strings", "strconv"} {
		require.Equal(t, name, server.ImportName(customRoot+name))
	}
}

// TestHTTPPlanReservesFixedImportsWrittenByThePackage verifies that standard
// packages keep the names written by HTTP templates when generated sections
// use them.
func TestHTTPPlanReservesFixedImportsWrittenByThePackage(t *testing.T) {
	const customRoot = "example.com/custom/"
	root := expr.RunDSL(t, func() {
		failure := dsl.Type("Failure", func() {
			dsl.Attribute("message", dsl.String)
		})
		dsl.Service("Calc", func() {
			dsl.Error("failed", failure)
			dsl.Method("Read", func() {
				dsl.Result(func() {
					for _, name := range []string{"errors", "strings", "strconv"} {
						dsl.Attribute(name, dsl.String, func() {
							dsl.Meta("struct:field:type", name+".Value", customRoot+name, name)
						})
					}
					dsl.Attribute("numbers", dsl.ArrayOf(dsl.Int))
				})
				dsl.Error("failed")
				dsl.HTTP(func() {
					dsl.GET("/")
					dsl.Response(func() {
						dsl.Header("numbers:X-Numbers")
					})
					dsl.Response("failed", dsl.StatusBadRequest)
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	server := generation.Package("generated.local/gen/http/calc/server")
	for _, name := range []string{"errors", "strings", "strconv"} {
		require.Equal(t, name+"2", server.ImportName(customRoot+name))
	}
}

// TestHTTPFilePlansMatchRenderedImports verifies that planning and rendering
// agree on every import used by a simple HTTP service.
func TestHTTPFilePlansMatchRenderedImports(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			dsl.Method("Read", func() {
				dsl.Result(dsl.String)
				dsl.HTTP(func() { dsl.GET("/") })
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	files := append(plans[0].ClientFiles(), plans[0].ServerFiles()...)
	files = append(files, plans[0].ClientTypeFiles()...)
	files = append(files, plans[0].ServerTypeFiles()...)
	files = append(files, plans[0].PathFiles()...)
	files = append(files, plans[0].ClientCLIFiles()...)
	directory := t.TempDir()
	for _, file := range files {
		if file == nil || filepath.Ext(file.Path) != ".go" {
			continue
		}
		planned := importPaths(file.SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec))
		renderedPath, err := file.Render(directory)
		require.NoError(t, err)
		source, err := os.ReadFile(renderedPath)
		require.NoError(t, err)
		parsed, err := parser.ParseFile(token.NewFileSet(), renderedPath, source, parser.ImportsOnly)
		require.NoError(t, err)
		actual := make([]string, len(parsed.Imports))
		for index, spec := range parsed.Imports {
			actual[index], err = strconv.Unquote(spec.Path.Value)
			require.NoError(t, err)
		}
		require.ElementsMatch(t, actual, planned, file.Path)
	}
}

// TestHTTPFilePlansIncludeGeneratedUses verifies that files reserve the
// packages named by copied validators, constructors, and query conversion.
func TestHTTPFilePlansIncludeGeneratedUses(t *testing.T) {
	t.Run("required basic authentication", func(t *testing.T) {
		root := expr.RunDSL(t, func() {
			basic := dsl.BasicAuthSecurity("basic")
			dsl.API("secure", func() {
				dsl.Security(basic)
			})
			dsl.Service("Account", func() {
				dsl.Method("Read", func() {
					dsl.Payload(func() {
						dsl.Username("username")
						dsl.Password("password")
						dsl.Required("username", "password")
					})
					dsl.HTTP(func() {
						dsl.GET("/account")
					})
				})
			})
		})
		plan := linkedHTTPPlanForRoot(t, root)
		serverImports := plannedHTTPFileImports(t, plan.ServerFiles(), "/server/encode_decode.go")
		clientCLIImports := plannedHTTPFileImports(t, plan.ClientCLIFiles(), "/client/cli.go")
		require.Contains(t, serverImports, codegen.GoaImport("").Path)
		require.Contains(t, clientCLIImports, "fmt")
	})

	t.Run("authorization bearer header", func(t *testing.T) {
		root := expr.RunDSL(t, testdata.PayloadJWTAuthorizationHeaderDSL)
		plan := linkedHTTPPlanForRoot(t, root)
		clientImports := plannedHTTPFileImports(t, plan.ClientFiles(), "/client/encode_decode.go")
		require.Contains(t, clientImports, "strings")
	})

	t.Run("path default ignored by client command", func(t *testing.T) {
		root := expr.RunDSL(t, testdata.PayloadPathStringDefaultDSL)
		plan := linkedHTTPPlanForRoot(t, root)
		clientImports := plannedHTTPFileImports(t, plan.ClientCLIFiles(), "/client/cli.go")
		require.Contains(t, clientImports, "fmt")
	})

	t.Run("nested string validation", func(t *testing.T) {
		root := expr.RunDSL(t, testdata.ResultWithResultCollectionDSL)
		plan := linkedHTTPPlanForRoot(t, root)
		file := plan.ClientTypeFiles()[0]
		imports := importPaths(file.SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec))
		require.Contains(t, imports, "unicode/utf8")
	})

	t.Run("numeric map query", func(t *testing.T) {
		root := expr.RunDSL(t, testdata.PayloadMapQueryObjectDSL)
		plan := linkedHTTPPlanForRoot(t, root)
		client := plan.ClientFiles()[1]
		server := plan.ServerFiles()[1]
		clientImports := importPaths(client.SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec))
		serverImports := importPaths(server.SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec))
		require.Contains(t, clientImports, "strconv")
		require.Contains(t, serverImports, "strconv")
	})

	t.Run("built-in error constructor", func(t *testing.T) {
		root := expr.RunDSL(t, func() {
			dsl.Service("Records", func() {
				dsl.Method("Read", func() {
					dsl.Error("not_found")
					dsl.HTTP(func() {
						dsl.GET("/records")
						dsl.Response("not_found", dsl.StatusNotFound)
					})
				})
			})
		})
		plan := linkedHTTPPlanForRoot(t, root)
		file := plan.ServerTypeFiles()[0]
		imports := importPaths(file.SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec))
		require.Contains(t, imports, codegen.GoaImport("").Path)
	})
}

// importPaths returns the package paths from one generated file header.
func importPaths(imports []*codegen.ImportSpec) []string {
	paths := make([]string, len(imports))
	for index, spec := range imports {
		paths[index] = spec.Path
	}
	return paths
}

// plannedHTTPFileImports returns the packages recorded for the generated file
// whose path ends in suffix.
func plannedHTTPFileImports(t *testing.T, files []*codegen.File, suffix string) []string {
	t.Helper()
	for _, file := range files {
		if strings.HasSuffix(filepath.ToSlash(file.Path), suffix) {
			return importPaths(file.SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec))
		}
	}
	require.Fail(t, "generated HTTP file was not planned", suffix)
	return nil
}

// TestJSONRPCSnapshotsExposeReleasedNames checks that copied JSON-RPC data
// gives existing plugins the final Go name stored in each declaration.
func TestJSONRPCSnapshotsExposeReleasedNames(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			for _, name := range []string{"read-data", "read_data"} {
				dsl.Method(name, func() {
					dsl.Payload(func() {
						dsl.Attribute("value", dsl.Int)
					})
					dsl.Result(dsl.Int)
					dsl.JSONRPC(func() {})
				})
			}
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewJSONRPCPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	snapshot, ok := plans[0].JSONRPCService("Calc")
	require.True(t, ok)
	assertReleasedName(t, snapshot.ServerStruct, snapshot.ServerStructDeclaration)
	assertReleasedName(t, snapshot.ServerInit, snapshot.ServerInitDeclaration)
	assertReleasedName(t, snapshot.MountServer, snapshot.MountServerDeclaration)
	assertReleasedName(t, snapshot.ClientStruct, snapshot.ClientStructDeclaration)
	require.Len(t, snapshot.Endpoints, 2)
	for index := range snapshot.Endpoints {
		endpoint := &snapshot.Endpoints[index]
		assertReleasedName(t, endpoint.HandlerInit, endpoint.HandlerInitDeclaration)
		assertReleasedName(t, endpoint.ClientStruct, endpoint.ClientStructDeclaration)
		assertReleasedName(t, endpoint.RequestEncoder, endpoint.RequestEncoderDeclaration)
		assertReleasedName(t, endpoint.RequestDecoder, endpoint.RequestDecoderDeclaration)
		assertReleasedName(t, endpoint.ResponseDecoder, endpoint.ResponseDecoderDeclaration)
	}
	require.NotEqual(t, snapshot.Endpoints[0].HandlerInit, snapshot.Endpoints[1].HandlerInit)
}

// TestViewedResultSnapshotsReturnIndependentBodies changes one returned body
// and confirms that a later copy still contains the planned value.
func TestViewedResultSnapshotsReturnIndependentBodies(t *testing.T) {
	root := expr.RunDSL(t, func() {
		result := dsl.ResultType("application/vnd.body-view", func() {
			dsl.Attribute("id", dsl.String)
			dsl.Required("id")
			dsl.View("default", func() { dsl.Attribute("id") })
		})
		dsl.Service("Bodies", func() {
			dsl.Method("Fetch", func() {
				dsl.Result(result)
				dsl.JSONRPC(func() {})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewJSONRPCPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	viewed, ok := plans[0].ViewedResult("Bodies", "Fetch")
	require.True(t, ok)
	require.Len(t, viewed.Representations, 1)
	require.NotNil(t, viewed.Representations[0].ServerBody)
	require.NotNil(t, viewed.Representations[0].ClientBody)
	originalRef := viewed.Representations[0].ServerBody.Ref
	viewed.Representations[0].ServerBody.Ref = "Changed"
	fresh, ok := plans[0].ViewedResult("Bodies", "Fetch")
	require.True(t, ok)
	require.Equal(t, originalRef, fresh.Representations[0].ServerBody.Ref)
}

// TestViewedResultCopiesBodyFieldSelection checks that JSON-RPC receives the
// Go field selected by Body("value"). An empty field keeps the whole-result
// body constructor responsible for the server conversion.
func TestViewedResultCopiesBodyFieldSelection(t *testing.T) {
	root := expr.RunDSL(t, func() {
		result := dsl.ResultType("application/vnd.body-field", func() {
			dsl.TypeName("BodyField")
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
			dsl.View("default", func() { dsl.Attribute("value") })
			dsl.View("summary", func() { dsl.Attribute("value") })
		})
		dsl.Service("Values", func() {
			dsl.Method("Field", func() {
				dsl.Result(result)
				dsl.JSONRPC(func() {
					dsl.Response(func() { dsl.Body("value") })
				})
			})
			dsl.Method("Whole", func() {
				dsl.Result(result)
				dsl.JSONRPC(func() { dsl.Response(dsl.StatusOK) })
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewJSONRPCPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	field, ok := plans[0].ViewedResult("Values", "Field")
	require.True(t, ok)
	require.NotEmpty(t, field.Representations)
	for _, representation := range field.Representations {
		require.Equal(t, "Value", representation.ResultAttr)
		require.NotNil(t, representation.ServerBody)
		require.Nil(t, representation.ServerBody.Init)
	}

	whole, ok := plans[0].ViewedResult("Values", "Whole")
	require.True(t, ok)
	require.NotEmpty(t, whole.Representations)
	for _, representation := range whole.Representations {
		require.Empty(t, representation.ResultAttr)
		require.NotNil(t, representation.ServerBody)
		require.NotNil(t, representation.ServerBody.Init)
	}
}

// TestEndpointConstructorsUsePackageDeclarations checks that request payload,
// response result, and error result functions all use the names chosen before
// files are written.
func TestEndpointConstructorsUsePackageDeclarations(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			dsl.Method("Add", func() {
				dsl.Payload(func() { dsl.Attribute("value", dsl.Int) })
				dsl.Result(func() { dsl.Attribute("total", dsl.Int) })
				dsl.Error("BadInput", func() { dsl.Attribute("message", dsl.String) })
				dsl.HTTP(func() {
					dsl.POST("/add")
					dsl.Response("BadInput", dsl.StatusBadRequest)
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	endpoint := plans[0].services.Get("Calc").Endpoints[0]
	require.NotNil(t, endpoint.Payload.Request.PayloadInit.Declaration)
	require.Equal(t, endpoint.Payload.Request.PayloadInit.Declaration.Name(), endpoint.Payload.Request.PayloadInit.Name)
	require.NotNil(t, endpoint.Result.Responses[0].ResultInit.Declaration)
	require.Equal(t, endpoint.Result.Responses[0].ResultInit.Declaration.Name(), endpoint.Result.Responses[0].ResultInit.Name)
	require.NotNil(t, endpoint.Errors[0].Errors[0].Response.ResultInit.Declaration)
	require.Equal(t, endpoint.Errors[0].Errors[0].Response.ResultInit.Declaration.Name(), endpoint.Errors[0].Errors[0].Response.ResultInit.Name)
}

// TestHTTPTypeAndConstructorNamesShareOnePackage checks a body type and a
// payload constructor that request the same spelling. Goa must give them
// different names, and generated definitions and calls must use those names.
func TestHTTPTypeAndConstructorNamesShareOnePackage(t *testing.T) {
	root := expr.RunDSL(t, func() {
		payload := dsl.Type("NewAddPayload", func() {
			dsl.Attribute("value", dsl.Int)
		})
		dsl.Service("Calc", func() {
			dsl.Method("Add", func() {
				dsl.Payload(payload)
				dsl.HTTP(func() { dsl.POST("/add") })
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	request := plans[0].services.Get("Calc").Endpoints[0].Payload.Request
	require.NotEqual(t, request.ServerBody.Name, request.PayloadInit.Name)
	definitions := renderedFiles(t, plans[0].ServerTypeFiles())
	calls := renderedFiles(t, plans[0].ServerFiles())
	require.Contains(t, definitions, "type "+request.ServerBody.Name+" ")
	require.Contains(t, definitions, "func "+request.PayloadInit.Name+"(")
	require.Contains(t, calls, request.PayloadInit.Name+"(")
}

// TestNewPlansAssignsNamesAcrossRoots checks two designs whose service names
// resolve to the same generated directory. NewPlans must submit both sets of
// function names together so definitions and calls remain distinct.
func TestNewPlansAssignsNamesAcrossRoots(t *testing.T) {
	makeRoot := func(apiName string) *expr.RootExpr {
		return expr.RunDSL(t, func() {
			dsl.API(apiName, func() {})
			dsl.Service("Shared", func() {
				dsl.Method("Add", func() {
					dsl.Payload(func() { dsl.Attribute("value", dsl.Int) })
					dsl.HTTP(func() { dsl.POST("/add") })
				})
			})
		})
	}
	first := makeRoot("First")
	second := makeRoot("Second")
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{first, second})
	require.NoError(t, err)
	servicePlans, err := service.NewPlans(generation,
		service.PlanInput{Root: first, Examples: expr.NewExampleGenerator(first.API.RandomizerFactory)},
		service.PlanInput{Root: second, Examples: expr.NewExampleGenerator(second.API.RandomizerFactory)},
	)
	require.NoError(t, err)
	plans, err := NewPlans(generation,
		PlanInput{Root: first, Service: servicePlans[0]},
		PlanInput{Root: second, Service: servicePlans[1]},
	)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	for index := range plans {
		require.NoError(t, servicePlans[index].Link())
		require.NoError(t, plans[index].Link())
	}

	firstService := plans[0].services.Get("Shared")
	secondService := plans[1].services.Get("Shared")
	firstInit := firstService.Endpoints[0].Payload.Request.PayloadInit
	secondInit := secondService.Endpoints[0].Payload.Request.PayloadInit
	require.NotEqual(t, firstInit.Name, secondInit.Name)
	require.Equal(t, plans[0].ServerTypeFiles()[0].Path, plans[1].ServerTypeFiles()[0].Path)
	for index, init := range []*InitData{firstInit, secondInit} {
		definitions := renderedFiles(t, plans[index].ServerTypeFiles())
		calls := renderedFiles(t, plans[index].ServerFiles())
		require.Contains(t, definitions, "func "+init.Name+"(")
		require.Contains(t, calls, init.Name+"(")
	}
}

// TestHTTPHelperDefinitionsUseAssignedNames checks two designs that write the
// same server package. File helpers and mixed-result stream helpers must define
// the same names that their call sites use.
func TestHTTPHelperDefinitionsUseAssignedNames(t *testing.T) {
	makeRoot := func(serviceName, typePrefix string) *expr.RootExpr {
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
	first := makeRoot("Foo Bar", "First")
	second := makeRoot("Foo-Bar", "Second")
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{first, second})
	require.NoError(t, err)
	servicePlans, err := service.NewPlans(generation,
		service.PlanInput{Root: first, Examples: expr.NewExampleGenerator(first.API.RandomizerFactory)},
		service.PlanInput{Root: second, Examples: expr.NewExampleGenerator(second.API.RandomizerFactory)},
	)
	require.NoError(t, err)
	plans, err := NewPlans(generation,
		PlanInput{Root: first, Service: servicePlans[0]},
		PlanInput{Root: second, Service: servicePlans[1]},
	)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	for index := range plans {
		require.NoError(t, servicePlans[index].Link())
		require.NoError(t, plans[index].Link())
	}

	for index, name := range []string{"Foo Bar", "Foo-Bar"} {
		data := plans[index].services.Get(name)
		endpoint := data.Endpoints[0]
		code := renderedFiles(t, plans[index].ServerFiles())
		require.Contains(t, code, "type "+endpoint.DiscardStreamDeclaration.Name()+" struct{}")
		require.Contains(t, code, "type "+data.AppendFSDeclaration.Name()+" struct {")
		require.Contains(t, code, "func "+data.AppendPrefixDeclaration.Name()+"(")
		require.Contains(t, code, "return "+data.AppendFSDeclaration.Name()+"{")
	}
}

// TestNewPlansRejectsDifferentServiceRoot checks the pairing before HTTP
// planning changes any generated package. The valid retry proves the rejected
// call did not reserve imports or names for the wrong design.
func TestNewPlansRejectsDifferentServiceRoot(t *testing.T) {
	first := expr.RunDSL(t, func() {
		dsl.Service("First", func() {
			dsl.Method("Read", func() { dsl.HTTP(func() { dsl.GET("/first") }) })
		})
	})
	second := expr.RunDSL(t, func() {
		dsl.Service("Second", func() {
			dsl.Method("Read", func() { dsl.HTTP(func() { dsl.GET("/second") }) })
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{first, second})
	require.NoError(t, err)
	plans, err := service.NewPlans(generation,
		service.PlanInput{Root: first, Examples: expr.NewExampleGenerator(first.API.RandomizerFactory)},
		service.PlanInput{Root: second, Examples: expr.NewExampleGenerator(second.API.RandomizerFactory)},
	)
	require.NoError(t, err)
	_, err = NewPlans(generation,
		PlanInput{Root: first, Service: plans[1]},
		PlanInput{Root: second, Service: plans[0]},
	)
	require.EqualError(t, err, "HTTP root does not match its service plan root")

	_, err = NewPlans(generation,
		PlanInput{Root: first, Service: plans[0]},
		PlanInput{Root: second, Service: plans[1]},
	)
	require.NoError(t, err)
}

// TestPlanRequiresLinkedServicePlan proves HTTP generation cannot read service
// names before Goa has assigned every package name.
func TestPlanRequiresLinkedServicePlan(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			dsl.Method("Add", func() { dsl.HTTP(func() { dsl.POST("/add") }) })
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.PanicsWithValue(t, "service render model requested before plan linking", func() {
		_ = plans[0].Link()
	})
}

// assertIndependentCodecFile changes every file field that the JSON-RPC
// generator edits, then checks that both the saved file and a new copy keep
// their original values.
func assertIndependentCodecFile(t *testing.T, saved *codegen.File, copyFile func() *codegen.File) {
	t.Helper()
	require.NotNil(t, saved)
	require.NotEmpty(t, saved.SectionTemplates)

	originalPath := saved.Path
	originalName := saved.SectionTemplates[0].Name
	originalSource := saved.SectionTemplates[0].Source
	originalImports := append([]*codegen.ImportSpec(nil), saved.SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec)...)

	changed := copyFile()
	changed.Path = "changed.go"
	changed.SectionTemplates[0].Name = "changed"
	changed.SectionTemplates[0].Source = "changed"
	changed.SectionTemplates[0].FuncMap = map[string]any{"changed": true}
	codegen.AddImport(changed.SectionTemplates[0], &codegen.ImportSpec{Path: "changed.example/package"})
	var changedEndpoint bool
	for _, section := range changed.SectionTemplates {
		if endpoint := codecEndpointData(section.Data); endpoint != nil {
			endpoint.Method.Name = "Changed"
			changedEndpoint = true
			break
		}
	}
	require.True(t, changedEndpoint)

	fresh := copyFile()
	require.Equal(t, originalPath, saved.Path)
	require.Equal(t, originalPath, fresh.Path)
	require.Equal(t, originalName, saved.SectionTemplates[0].Name)
	require.Equal(t, originalName, fresh.SectionTemplates[0].Name)
	require.Equal(t, originalSource, saved.SectionTemplates[0].Source)
	require.Equal(t, originalSource, fresh.SectionTemplates[0].Source)
	require.Equal(t, originalImports, saved.SectionTemplates[0].Data.(map[string]any)["Imports"])
	require.Equal(t, originalImports, fresh.SectionTemplates[0].Data.(map[string]any)["Imports"])
	for _, section := range fresh.SectionTemplates {
		if endpoint := codecEndpointData(section.Data); endpoint != nil {
			require.Equal(t, "Add", endpoint.Method.Name)
			return
		}
	}
	require.Fail(t, "copied codec file has no endpoint section")
}

// codecEndpointData returns the copied endpoint value used by one encoder or
// decoder section.
func codecEndpointData(data any) *JSONRPCEndpointSnapshot {
	switch actual := data.(type) {
	case *JSONRPCEndpointSnapshot:
		return actual
	case *jsonRPCRequestCodecData:
		return actual.JSONRPCEndpointSnapshot
	default:
		return nil
	}
}
