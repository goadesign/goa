// This file verifies that HTTP package names are requested before Goa assigns
// them and that files are built only after service names are available.
package codegen

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestPlanReservesStaticAliasesBeforeFreeze(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Path", func() {
			dsl.Method("Read", func() {})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	_, err = NewPlans(generation)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	services := servicePlan.Services()

	require.Equal(t, "path2", services.ServiceImport("Path").Name)
}

func TestPlanRejectsFrozenGeneration(t *testing.T) {
	generation, err := codegen.NewGeneration("generated.local/gen", nil)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())

	_, err = NewPlans(generation)
	require.Error(t, err)
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
	_, err = NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	services := servicePlan.Services()

	client := services.PackageImport("generated.local/gen/http/foo/client")
	server := services.PackageImport("generated.local/gen/http/foo/server")
	cli := services.PackageImport(path.Join(
		"generated.local/gen/http/cli",
		codegen.SnakeCase(codegen.Goify(root.API.Servers[0].Name, true)),
	))
	require.NotEqual(t, services.ServiceImport("Fooc").Name, client.Name)
	require.NotEqual(t, services.ServiceImport("Foosvr").Name, server.Name)
	require.NotEmpty(t, cli.Name)
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
	require.NoError(t, example.Plan(generation))
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	plan := plans[0]
	require.NoError(t, plan.Link())
	_, ok := plan.JSONRPCService("Calc")
	require.True(t, ok)
	require.NotEmpty(t, plan.ExampleServerFiles())
	require.NotEmpty(t, plan.ExampleCLIFiles())
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
	require.NoError(t, example.Plan(generation))
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	service, ok := plans[0].JSONRPCService("Calc")
	require.True(t, ok)
	stored := plans[0].jsonServices["Calc"]
	assertIndependentCodecFile(t, stored.clientCodec, service.ClientCodecFile)
	assertIndependentCodecFile(t, stored.serverCodec, service.ServerCodecFile)

	clientPath := stored.clientCodec.Path
	imports := service.FileImports(clientPath)
	require.NotEmpty(t, imports)
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

// TestViewedResultSnapshotsPreserveMissingBodies checks that a successful
// response containing only a mapped header keeps both body values absent. It
// also changes the returned header and confirms a later copy is unchanged.
func TestViewedResultSnapshotsPreserveMissingBodies(t *testing.T) {
	root := expr.RunDSL(t, func() {
		result := dsl.ResultType("application/vnd.header-view", func() {
			dsl.Attribute("id", dsl.String)
			dsl.Required("id")
			dsl.View("default", func() { dsl.Attribute("id") })
		})
		dsl.Service("Headers", func() {
			dsl.Method("Fetch", func() {
				dsl.Result(result)
				dsl.JSONRPC(func() {
					dsl.Response(func() { dsl.Header("id") })
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewJSONRPCPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, example.Plan(generation))
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	viewed, ok := plans[0].ViewedResult("Headers", "Fetch")
	require.True(t, ok)
	require.Len(t, viewed.Representations, 1)
	require.Nil(t, viewed.Representations[0].ServerBody)
	require.Nil(t, viewed.Representations[0].ClientBody)
	require.Len(t, viewed.Representations[0].Headers, 1)
	originalHeader := viewed.Representations[0].Headers[0].CanonicalName
	viewed.Representations[0].Headers[0].CanonicalName = "Changed"
	fresh, ok := plans[0].ViewedResult("Headers", "Fetch")
	require.True(t, ok)
	require.Equal(t, originalHeader, fresh.Representations[0].Headers[0].CanonicalName)
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
	require.NoError(t, example.Plan(generation))
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
			dsl.Error("BadInput", func() { dsl.Attribute("message", dsl.String) })
			dsl.Method("Add", func() {
				dsl.Payload(func() { dsl.Attribute("value", dsl.Int) })
				dsl.Result(func() { dsl.Attribute("total", dsl.Int) })
				dsl.Error("BadInput")
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
	require.NoError(t, example.Plan(generation))
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
	require.NoError(t, example.Plan(generation))
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
	makeRoot := func(serviceName string) *expr.RootExpr {
		return expr.RunDSL(t, func() {
			dsl.Service(serviceName, func() {
				dsl.Method("Add", func() {
					dsl.Payload(func() { dsl.Attribute("value", dsl.Int) })
					dsl.HTTP(func() { dsl.POST("/add") })
				})
			})
		})
	}
	first := makeRoot("Foo Bar")
	second := makeRoot("Foo-Bar")
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
	require.NoError(t, example.Plan(generation))
	require.NoError(t, generation.Freeze())
	for index := range plans {
		require.NoError(t, servicePlans[index].Link())
		require.NoError(t, plans[index].Link())
	}

	firstService := plans[0].services.Get("Foo Bar")
	secondService := plans[1].services.Get("Foo-Bar")
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
	require.NoError(t, example.Plan(generation))
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
