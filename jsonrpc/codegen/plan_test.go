// This file verifies standalone JSON-RPC planning includes the HTTP codecs and
// helpers that JSON-RPC rendering reuses.
package codegen

import (
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

func TestPlanIncludesSharedHTTPImportAliases(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("UUID", func() {
			dsl.Method("Read", func() {
				dsl.JSONRPC(func() {})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	_, err = NewPlans(generation, PlanInput{Root: root, Service: servicePlan, HTTP: httpPlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	services := servicePlan.Services()

	clientOutput := "generated.local/gen/jsonrpc/uuid/client"
	require.Equal(t, "uuid2", services.ServiceImport(clientOutput, "UUID").Name)
}

// TestPlanRetainsServicePackageImport verifies JSON-RPC output packages use
// the service package selected before later expression changes. The retained
// preferred name must still resolve around the client's sync import.
func TestPlanRetainsServicePackageImport(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Sync", func() {
			dsl.Method("Read", func() {
				dsl.JSONRPC(func() {})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)

	transport := root.API.JSONRPC.Services[0]
	transport.ServiceExpr.Name = "Changed"
	_, err = NewPlans(generation, PlanInput{Root: root, Service: servicePlan, HTTP: httpPlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())

	client := generation.Package("generated.local/gen/jsonrpc/sync/client")
	require.Equal(t, "sync2", client.ImportName("generated.local/gen/sync"))
}

// TestPlanSelectsPositionalParams checks that strings, bytes, and Any values
// are wrapped once while objects, arrays, maps, and unions keep their JSON shape.
func TestPlanSelectsPositionalParams(t *testing.T) {
	_, transport := linkedJSONRPCPlan(t, jsonRPCParamsPlanDSL)
	endpoints := transport.services[0].data.Endpoints

	want := map[string]struct {
		request     bool
		positional  bool
		allowAbsent bool
		omitAbsent  bool
		stream      bool
	}{
		"text":                 {request: true, positional: true},
		"alias":                {request: true, positional: true},
		"anything":             {request: true, positional: true},
		"bytes":                {request: true, positional: true},
		"object":               {request: true},
		"array":                {request: true},
		"map":                  {request: true},
		"union":                {request: true},
		"optional_text":        {request: true, positional: true, allowAbsent: true},
		"optional_object":      {request: true, allowAbsent: true, omitAbsent: true},
		"optional_array":       {request: true, allowAbsent: true, omitAbsent: true},
		"optional_map":         {request: true, allowAbsent: true, omitAbsent: true},
		"optional_union":       {request: true, allowAbsent: true, omitAbsent: true},
		"stream_text":          {stream: true, positional: true},
		"stream_any":           {stream: true, positional: true},
		"stream_array":         {stream: true},
		"stream_optional_text": {stream: true, positional: true, allowAbsent: true},
		"stream_default_text":  {stream: true, positional: true, allowAbsent: true},
		"stream_object_data":   {stream: true, allowAbsent: true, omitAbsent: true},
		"stream_array_data":    {stream: true, allowAbsent: true, omitAbsent: true},
		"stream_map_data":      {stream: true, allowAbsent: true, omitAbsent: true},
		"stream_union_data":    {stream: true, allowAbsent: true, omitAbsent: true},
	}
	require.Len(t, endpoints, len(want))
	for _, endpoint := range endpoints {
		expected, ok := want[endpoint.Method.Name]
		require.True(t, ok, endpoint.Method.Name)
		request := endpoint.Payload != nil && endpoint.Payload.Request != nil && endpoint.Payload.Request.Params != nil
		stream := endpoint.SSE != nil && endpoint.SSE.Params != nil
		require.Equal(t, expected.request, request, endpoint.Method.Name)
		if expected.request {
			require.Equal(t, expected.positional, endpoint.Payload.Request.Params.Positional, endpoint.Method.Name)
			require.Equal(t, expected.allowAbsent, endpoint.Payload.Request.Params.AllowAbsent, endpoint.Method.Name)
			require.Equal(t, expected.omitAbsent, endpoint.Payload.Request.Params.OmitAbsent, endpoint.Method.Name)
		}
		require.Equal(t, expected.stream, stream, endpoint.Method.Name)
		if expected.stream {
			require.Equal(t, expected.positional, endpoint.SSE.Params.Positional, endpoint.Method.Name)
			require.Equal(t, expected.allowAbsent, endpoint.SSE.Params.AllowAbsent, endpoint.Method.Name)
			require.Equal(t, expected.omitAbsent, endpoint.SSE.Params.OmitAbsent, endpoint.Method.Name)
		}
	}
}

// TestNewExamplePlanRejectsAnotherServicePlan checks that server names and
// URLs cannot come from a different design with the same authored names.
func TestNewExamplePlanRejectsAnotherServicePlan(t *testing.T) {
	_, transport := linkedJSONRPCPlan(t, viewedJSONRPCPlanDSL)
	otherRoot := expr.RunDSL(t, viewedJSONRPCPlanDSL)
	otherGeneration, err := codegen.NewGeneration("other.local/gen", []eval.Root{otherRoot})
	require.NoError(t, err)
	otherService, err := service.NewPlan(otherRoot, otherGeneration, expr.NewExampleGenerator(otherRoot.API.RandomizerFactory))
	require.NoError(t, err)
	examples, err := example.NewPlan(otherGeneration, otherService)
	require.NoError(t, err)

	_, err = NewExamplePlan(transport, examples)
	require.EqualError(t, err, "JSON-RPC examples require server data created from the same service design")
}

// TestPlanReservesGeneratedJSONRPCPackages verifies that the JSON-RPC client,
// server, and CLI imports are frozen by their complete generated paths.
func TestPlanReservesGeneratedJSONRPCPackages(t *testing.T) {
	root := expr.RunDSL(t, func() {
		for _, name := range []string{"Foo", "Fooc", "Foojssvr"} {
			dsl.Service(name, func() {
				dsl.Method("Read", func() { dsl.JSONRPC(func() {}) })
			})
		}
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	_, err = NewPlans(generation, PlanInput{Root: root, Service: servicePlan, HTTP: httpPlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	services := servicePlan.Services()

	cliOutput := path.Join(
		"generated.local/gen/jsonrpc/cli",
		codegen.SnakeCase(codegen.Goify(root.API.Servers[0].Name, true)),
	)
	client := services.PackageImport(cliOutput, "generated.local/gen/jsonrpc/foo/client")
	serverOutput := path.Join("generated.local", "cmd", codegen.SnakeCase(codegen.Goify(root.API.Servers[0].Name, true)))
	server := services.PackageImport(serverOutput, "generated.local/gen/jsonrpc/foo/server")
	require.Equal(t, "fooc", client.Name)
	require.Equal(t, "foojssvr", server.Name)
}

// TestNewPlansRequiresEveryJSONRPCRoot verifies that planning cannot reserve
// names from only part of a generation. The caller must supply each root that
// declares a JSON-RPC service exactly once.
func TestNewPlansRequiresEveryJSONRPCRoot(t *testing.T) {
	generation, roots, services, jsonPlans, applicationPlans := jsonRPCPlanningInputs(t)

	_, err := NewPlans(generation, PlanInput{
		Root:            roots[0],
		Service:         services[0],
		HTTP:            jsonPlans[0],
		ApplicationHTTP: applicationPlans[0],
	})
	require.EqualError(t, err, "JSON-RPC planning requires all 2 JSON-RPC roots, got 1")
	assertViewedHelperNameAvailable(t, generation)
}

// TestNewPlansRejectsDuplicateRoot verifies that two inputs cannot plan the
// same JSON-RPC root. The rejected call must not consume a helper name that a
// later generator can use.
func TestNewPlansRejectsDuplicateRoot(t *testing.T) {
	generation, roots, services, jsonPlans, applicationPlans := jsonRPCPlanningInputs(t)
	input := PlanInput{
		Root:            roots[0],
		Service:         services[0],
		HTTP:            jsonPlans[0],
		ApplicationHTTP: applicationPlans[0],
	}

	_, err := NewPlans(generation, input, input)
	require.EqualError(t, err, "JSON-RPC root is planned more than once: First")
	assertViewedHelperNameAvailable(t, generation)
}

// TestNewPlansRejectsRootWithoutJSONRPC verifies that inputs contain only
// roots that declare JSON-RPC services. Ordinary HTTP roots are planned by the
// HTTP generator and must not influence JSON-RPC names.
func TestNewPlansRejectsRootWithoutJSONRPC(t *testing.T) {
	generation, roots, services, jsonPlans, applicationPlans := jsonRPCPlanningInputs(t)

	_, err := NewPlans(generation,
		PlanInput{Root: roots[0], Service: services[0], HTTP: jsonPlans[0], ApplicationHTTP: applicationPlans[0]},
		PlanInput{Root: roots[1], Service: services[1], HTTP: jsonPlans[1], ApplicationHTTP: applicationPlans[1]},
		PlanInput{Root: roots[2], Service: services[2], HTTP: jsonPlans[0], ApplicationHTTP: applicationPlans[2]},
	)
	require.EqualError(t, err, "root does not declare JSON-RPC services")
	assertViewedHelperNameAvailable(t, generation)
}

// TestNewPlansRejectsMismatchedInputPlans verifies that every service and HTTP
// plan belongs to the root in the same input. Validation finishes before any
// JSON-RPC helper name is submitted.
func TestNewPlansRejectsMismatchedInputPlans(t *testing.T) {
	tests := []struct {
		name   string
		change func([]*expr.RootExpr, []*service.Plan, []*httpcodegen.Plan, []*httpcodegen.Plan) []PlanInput
		error  string
	}{
		{
			name: "service",
			change: func(roots []*expr.RootExpr, services []*service.Plan, jsonPlans, applicationPlans []*httpcodegen.Plan) []PlanInput {
				return []PlanInput{
					{Root: roots[0], Service: services[1], HTTP: jsonPlans[1], ApplicationHTTP: applicationPlans[1]},
					{Root: roots[1], Service: services[1], HTTP: jsonPlans[1], ApplicationHTTP: applicationPlans[1]},
				}
			},
			error: "JSON-RPC service plan does not belong to root First",
		},
		{
			name: "JSON-RPC HTTP",
			change: func(roots []*expr.RootExpr, services []*service.Plan, jsonPlans, applicationPlans []*httpcodegen.Plan) []PlanInput {
				return []PlanInput{
					{Root: roots[0], Service: services[0], HTTP: jsonPlans[1], ApplicationHTTP: applicationPlans[0]},
					{Root: roots[1], Service: services[1], HTTP: jsonPlans[1], ApplicationHTTP: applicationPlans[1]},
				}
			},
			error: "JSON-RPC HTTP plan does not belong to root First and its service plan",
		},
		{
			name: "ordinary HTTP plan in JSON-RPC role",
			change: func(roots []*expr.RootExpr, services []*service.Plan, jsonPlans, applicationPlans []*httpcodegen.Plan) []PlanInput {
				return []PlanInput{
					{Root: roots[0], Service: services[0], HTTP: applicationPlans[0], ApplicationHTTP: applicationPlans[0]},
					{Root: roots[1], Service: services[1], HTTP: jsonPlans[1], ApplicationHTTP: applicationPlans[1]},
				}
			},
			error: "JSON-RPC HTTP plan does not belong to root First and its service plan",
		},
		{
			name: "application HTTP",
			change: func(roots []*expr.RootExpr, services []*service.Plan, jsonPlans, applicationPlans []*httpcodegen.Plan) []PlanInput {
				return []PlanInput{
					{Root: roots[0], Service: services[0], HTTP: jsonPlans[0], ApplicationHTTP: applicationPlans[1]},
					{Root: roots[1], Service: services[1], HTTP: jsonPlans[1], ApplicationHTTP: applicationPlans[1]},
				}
			},
			error: "application HTTP plan does not belong to root First and its service plan",
		},
		{
			name: "JSON-RPC plan in application HTTP role",
			change: func(roots []*expr.RootExpr, services []*service.Plan, jsonPlans, applicationPlans []*httpcodegen.Plan) []PlanInput {
				return []PlanInput{
					{Root: roots[0], Service: services[0], HTTP: jsonPlans[0], ApplicationHTTP: jsonPlans[0]},
					{Root: roots[1], Service: services[1], HTTP: jsonPlans[1], ApplicationHTTP: applicationPlans[1]},
				}
			},
			error: "application HTTP plan does not belong to root First and its service plan",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generation, roots, services, jsonPlans, applicationPlans := jsonRPCPlanningInputs(t)
			_, err := NewPlans(generation, test.change(roots, services, jsonPlans, applicationPlans)...)
			require.EqualError(t, err, test.error)
			assertViewedHelperNameAvailable(t, generation)
		})
	}
}

// TestPlanEmitsViewedEncoderForJSONRPCMethodOnHTTPService verifies that a
// service with ordinary HTTP and JSON-RPC methods writes the viewed-result
// encoder called by its generated JSON-RPC server.
func TestPlanEmitsViewedEncoderForJSONRPCMethodOnHTTPService(t *testing.T) {
	root := expr.RunDSL(t, viewedJSONRPCWithHTTPServiceDSL)
	plan := CreateJSONRPCPlan(root)
	require.Len(t, plan.services, 1)
	require.Len(t, plan.services[0].endpoints, 1)
	require.NotNil(t, plan.services[0].endpoints[0].viewed)
	helper := plan.services[0].helpers["JSONRPC"].encode.Name()

	var source strings.Builder
	for _, file := range plan.ServerFiles() {
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(&source))
		}
	}
	require.Contains(t, source.String(), "func "+helper+"(")
}

// TestPlanRequiresLinkBeforeRender verifies callers cannot read files before
// Link finishes or ask Link to build the same files twice.
func TestPlanRequiresLinkBeforeRender(t *testing.T) {
	root := expr.RunDSL(t, viewedJSONRPCPlanDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan, HTTP: httpPlans[0]})
	require.NoError(t, err)
	require.PanicsWithValue(t, "JSON-RPC files requested before Plan.Link", func() {
		plans[0].ServerFiles()
	})

	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	require.NoError(t, plans[0].Link())
	require.EqualError(t, plans[0].Link(), "JSON-RPC plan is already linked")
}

// TestPlanBuildsCombinedExampleWithoutChangingHTTP verifies Link creates a new
// runnable server with ordinary HTTP and JSON-RPC services and leaves the HTTP
// plan's file unchanged.
func TestPlanBuildsCombinedExampleWithoutChangingHTTP(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("mixed", func() {
			dsl.Method("read", func() {
				dsl.HTTP(func() { dsl.GET("/read") })
				dsl.JSONRPC(func() {})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	applicationPlans, err := httpcodegen.NewPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{
		Root:            root,
		Service:         servicePlan,
		HTTP:            httpPlans[0],
		ApplicationHTTP: applicationPlans[0],
	})
	require.NoError(t, err)
	examplePlan, err := example.NewPlan(generation, servicePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, applicationPlans[0].Link())
	require.NoError(t, httpPlans[0].Link())

	httpExamples, err := httpcodegen.NewExamplePlan(applicationPlans[0], examplePlan)
	require.NoError(t, err)
	httpFile := httpExamples.ServerFiles()[0]
	httpImports := append([]*codegen.ImportSpec(nil), httpFile.SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec)...)
	require.NoError(t, plans[0].Link())

	require.Equal(t, httpImports, httpFile.SectionTemplates[0].Data.(map[string]any)["Imports"])
	for _, section := range httpFile.SectionTemplates {
		switch section.Name {
		case "server-http-start", "server-http-init", "server-http-end":
			require.Empty(t, section.Data.(map[string]any)["JSONRPCServices"])
		}
	}
	examples, err := NewExamplePlan(plans[0], examplePlan)
	require.NoError(t, err)
	combined := examples.ServerFiles()[0]
	require.NotSame(t, httpFile, combined)
	for _, section := range combined.SectionTemplates {
		switch section.Name {
		case "server-http-start", "server-http-init", "server-http-end":
			require.Len(t, section.Data.(map[string]any)["JSONRPCServices"], 1)
		}
	}
}

// TestPlanUsesHTTPViewedRepresentationBranches verifies each method uses the
// body types and constructors that the HTTP plan prepared for its result views.
func TestPlanUsesHTTPViewedRepresentationBranches(t *testing.T) {
	shared, plan := linkedJSONRPCPlan(t, viewedJSONRPCPlanDSL)
	require.Len(t, plan.services, 1)

	endpoints := make(map[string]*endpointPlan)
	for _, endpoint := range plan.services[0].endpoints {
		endpoints[endpoint.Method.Name] = endpoint
	}
	variable := endpoints["fetch"].viewed
	require.True(t, variable.variable)
	httpViewed, ok := shared.ViewedResult("retained", "fetch")
	require.True(t, ok)
	require.Len(t, variable.branches, len(httpViewed.Representations))
	for index, branch := range variable.branches {
		require.Equal(t, httpViewed.Representations[index].View, branch.view)
	}
	for _, branch := range variable.branches {
		require.NotNil(t, branch.serverBody)
		require.NotNil(t, branch.clientBody)
		require.NotEmpty(t, branch.resultInit.Name)
	}

	fixed := endpoints["fixed"].viewed
	require.False(t, fixed.variable)
	require.Equal(t, "detailed", fixed.fixedView)
	require.Len(t, fixed.branches, 1)
	require.Equal(t, "detailed", fixed.branches[0].view)
}

// TestPlanUsesEveryViewForMappedResultField verifies that a JSON-RPC response
// mapped to one result field still carries and checks every view that the
// service method may return. Each branch uses the mapped field's JSON body and
// result constructor supplied by the HTTP plan.
func TestPlanUsesEveryViewForMappedResultField(t *testing.T) {
	shared, plan := linkedJSONRPCPlan(t, viewedJSONRPCMappedFieldPlanDSL)
	httpViewed, ok := shared.ViewedResult("mapped", "fetch")
	require.True(t, ok)
	representations := httpViewed.Representations
	require.Len(t, representations, 2)
	require.Len(t, plan.services, 1)
	require.Len(t, plan.services[0].endpoints, 1)
	viewed := plan.services[0].endpoints[0].viewed
	require.True(t, viewed.variable)
	require.Empty(t, viewed.fixedView)
	require.Len(t, viewed.branches, 2)
	for index, name := range []string{"summary", "default"} {
		require.Equal(t, name, viewed.branches[index].view)
		require.Equal(t, representations[index].ServerBody, viewed.branches[index].serverBody)
		require.Equal(t, representations[index].ClientBody, viewed.branches[index].clientBody)
		require.Equal(t, representations[index].ResultInit, viewed.branches[index].resultInit)
	}
	require.Equal(t, viewed.branches[0].serverBody, viewed.branches[1].serverBody)
	require.Equal(t, viewed.branches[0].clientBody, viewed.branches[1].clientBody)
	require.Equal(t, viewed.branches[0].resultInit, viewed.branches[1].resultInit)
}

// TestPlanTreatsSoleResultViewAsFixed verifies that a result type with one view
// produces a response body without a per-response view field. The service plan
// supplies that one view name, and JSON-RPC copies it without deriving a value
// from the first response branch.
func TestPlanTreatsSoleResultViewAsFixed(t *testing.T) {
	shared, plan := linkedJSONRPCPlan(t, viewedJSONRPCSoleViewPlanDSL)
	httpViewed, ok := shared.ViewedResult("sole", "fetch")
	require.True(t, ok)
	representations := httpViewed.Representations
	require.Len(t, representations, 1)
	require.Len(t, plan.services, 1)
	require.Len(t, plan.services[0].endpoints, 1)
	viewed := plan.services[0].endpoints[0].viewed
	require.False(t, viewed.variable)
	require.Equal(t, "default", viewed.fixedView)
	require.Len(t, viewed.branches, 1)
	require.Equal(t, "default", viewed.branches[0].view)
}

// TestPlanUsesAssignedViewedHelperNames verifies that result-view helpers use
// the unique names assigned when two method spellings produce the same Go name
// or another generated function already uses the requested name.
func TestPlanUsesAssignedViewedHelperNames(t *testing.T) {
	root := expr.RunDSL(t, viewedJSONRPCCollisionPlanDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	client, err := generation.ClaimPackage("generated.local/gen/jsonrpc/collisions/client")
	require.NoError(t, err)
	server, err := generation.ClaimPackage("generated.local/gen/jsonrpc/collisions/server")
	require.NoError(t, err)
	require.NoError(t, client.DeclareName(codegen.NewPreferredName(
		codegen.NameFunction,
		"decodeFetchItemViewedResult",
		codegen.UnexportedName,
		jsonRPCNameOrder{role: viewedResultDecoderRole},
	)))
	require.NoError(t, client.DeclareName(codegen.NewPreferredName(
		codegen.NameFunction,
		"decodeJSONRPCResult",
		codegen.UnexportedName,
		jsonRPCNameOrder{},
	)))
	require.NoError(t, server.DeclareName(codegen.NewPreferredName(
		codegen.NameFunction,
		"encodeFetchItemViewedResult",
		codegen.UnexportedName,
		jsonRPCNameOrder{role: viewedResultEncoderRole},
	)))

	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan, HTTP: httpPlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	require.NoError(t, plans[0].Link())

	helpers := plans[0].services[0].helpers
	require.Equal(t, "decodeJSONRPCResult2", plans[0].services[0].bodyDecoder.Name())
	require.Equal(t, "decodeFetchItemViewedResult2", helpers["fetch-item"].decode.Name())
	require.Equal(t, "encodeFetchItemViewedResult2", helpers["fetch-item"].encode.Name())
	require.NotEqual(t, helpers["fetch-item"].decode.Name(), helpers["fetch_item"].decode.Name())
	require.NotEqual(t, helpers["fetch-item"].encode.Name(), helpers["fetch_item"].encode.Name())
}

// linkedJSONRPCPlan evaluates one design, assigns all generated Go names, and
// links the service, HTTP, and JSON-RPC plans used by a test.
func linkedJSONRPCPlan(t *testing.T, design func()) (*httpcodegen.Plan, *Plan) {
	t.Helper()
	root := expr.RunDSL(t, design)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan, HTTP: httpPlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	require.NoError(t, plans[0].Link())
	return httpPlans[0], plans[0]
}

// jsonRPCPlanningInputs builds two roots and their matching service, ordinary
// HTTP, and JSON-RPC HTTP plans. Tests change one input before calling NewPlans
// to prove the constructor rejects an incomplete or mismatched set.
func jsonRPCPlanningInputs(t *testing.T) (*codegen.Generation, []*expr.RootExpr, []*service.Plan, []*httpcodegen.Plan, []*httpcodegen.Plan) {
	t.Helper()
	roots := []*expr.RootExpr{
		expr.RunDSL(t, jsonRPCPlanningRootDSL("First", "/first")),
		expr.RunDSL(t, jsonRPCPlanningRootDSL("Second", "/second")),
		expr.RunDSL(t, ordinaryHTTPPlanningRootDSL),
	}
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{roots[0], roots[1], roots[2]})
	require.NoError(t, err)
	servicePlans, err := service.NewPlans(generation,
		service.PlanInput{Root: roots[0], Examples: expr.NewExampleGenerator(roots[0].API.RandomizerFactory)},
		service.PlanInput{Root: roots[1], Examples: expr.NewExampleGenerator(roots[1].API.RandomizerFactory)},
		service.PlanInput{Root: roots[2], Examples: expr.NewExampleGenerator(roots[2].API.RandomizerFactory)},
	)
	require.NoError(t, err)
	httpInputs := []httpcodegen.PlanInput{
		{Root: roots[0], Service: servicePlans[0]},
		{Root: roots[1], Service: servicePlans[1]},
		{Root: roots[2], Service: servicePlans[2]},
	}
	applicationPlans, err := httpcodegen.NewPlans(generation, httpInputs...)
	require.NoError(t, err)
	jsonPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpInputs[:2]...)
	require.NoError(t, err)
	return generation, roots, servicePlans, jsonPlans, applicationPlans
}

// assertViewedHelperNameAvailable submits the helper name that a rejected plan
// would have used and verifies no earlier JSON-RPC input consumed it.
func assertViewedHelperNameAvailable(t *testing.T, generation *codegen.Generation) {
	t.Helper()
	client, err := generation.ClaimPackage(path.Join("generated.local/gen/jsonrpc/first/client"))
	require.NoError(t, err)
	declaration := codegen.NewPreferredName(
		codegen.NameFunction,
		"decodeReadViewedResult",
		codegen.UnexportedName,
		jsonRPCNameOrder{service: "zzzz", method: "read", role: viewedResultDecoderRole},
	)
	require.NoError(t, client.DeclareName(declaration))
	require.NoError(t, generation.Freeze())
	require.Equal(t, "decodeReadViewedResult", declaration.Name())
}

// jsonRPCPlanningRootDSL defines one viewed method exposed through both
// ordinary HTTP and JSON-RPC so tests can also validate ApplicationHTTP.
func jsonRPCPlanningRootDSL(name, route string) func() {
	return func() {
		result := dsl.ResultType("application/vnd."+strings.ToLower(name), func() {
			dsl.Attribute("id", dsl.String)
			dsl.Required("id")
			dsl.View("default", func() {
				dsl.Attribute("id")
			})
		})
		dsl.Service(name, func() {
			dsl.Method("read", func() {
				dsl.Result(result)
				dsl.HTTP(func() {
					dsl.GET(route)
				})
				dsl.JSONRPC(func() {})
			})
		})
	}
}

// ordinaryHTTPPlanningRootDSL defines a root that must be excluded from
// JSON-RPC inputs even though it participates in the same generation.
func ordinaryHTTPPlanningRootDSL() {
	dsl.Service("Ordinary", func() {
		dsl.Method("read", func() {
			dsl.HTTP(func() {
				dsl.GET("/ordinary")
			})
		})
	})
}

// viewedJSONRPCWithHTTPServiceDSL defines one service whose ordinary HTTP and
// JSON-RPC methods return the same one-view result type.
func viewedJSONRPCWithHTTPServiceDSL() {
	result := dsl.ResultType("application/vnd.viewed-jsonrpc-http-service", func() {
		dsl.Attribute("value", dsl.String)
		dsl.Required("value")
		dsl.View("default", func() {
			dsl.Attribute("value")
		})
	})
	dsl.Service("ViewedHTTPJSON", func() {
		dsl.Method("HTTP", func() {
			dsl.Result(result)
			dsl.HTTP(func() {
				dsl.GET("/http")
			})
		})
		dsl.Method("JSONRPC", func() {
			dsl.Result(result)
			dsl.JSONRPC(func() {})
		})
	})
}

// viewedJSONRPCPlanDSL defines one variable and one fixed viewed result.
func viewedJSONRPCPlanDSL() {
	result := dsl.ResultType("application/vnd.retained-view", func() {
		dsl.TypeName("RetainedView")
		dsl.Attribute("id", dsl.String)
		dsl.Attribute("detail", dsl.String)
		dsl.Required("id", "detail")
		dsl.View("summary", func() {
			dsl.Attribute("id")
		})
		dsl.View("detailed", func() {
			dsl.Attribute("id")
			dsl.Attribute("detail")
		})
	})
	dsl.Service("retained", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("fetch", func() {
			dsl.Result(result)
			dsl.JSONRPC(func() {})
		})
		dsl.Method("fixed", func() {
			dsl.Result(result, func() {
				dsl.View("detailed")
			})
			dsl.JSONRPC(func() {})
		})
	})
}

// jsonRPCParamsPlanDSL declares single and container request and stream values
// so planning tests can compare their JSON-RPC parameter shape.
func jsonRPCParamsPlanDSL() {
	alias := dsl.Type("Alias", dsl.String)
	object := dsl.Type("Object", func() {
		dsl.Attribute("name", dsl.String)
		dsl.Required("name")
	})
	union := dsl.Type("Choice", func() {
		dsl.OneOf("value", func() {
			dsl.Field(1, "name", dsl.String)
			dsl.Field(2, "count", dsl.Int)
		})
	})
	objects := dsl.Type("Objects", dsl.ArrayOf(object))
	objectMap := dsl.Type("ObjectMap", dsl.MapOf(dsl.String, object))
	objectEvent := dsl.Type("ObjectEvent", func() {
		dsl.Attribute("data", object)
	})
	arrayEvent := dsl.Type("ArrayEvent", func() {
		dsl.Attribute("data", objects)
	})
	mapEvent := dsl.Type("MapEvent", func() {
		dsl.Attribute("data", objectMap)
	})
	unionEvent := dsl.Type("UnionEvent", func() {
		dsl.OneOf("data", func() {
			dsl.TypeName("StreamUnionData")
			dsl.Field(1, "name", dsl.String)
			dsl.Field(2, "count", dsl.Int)
		})
	})
	optionalTextEvent := dsl.Type("OptionalTextEvent", func() {
		dsl.Attribute("data", dsl.String, func() {
			dsl.Pattern("^[a-z]*$")
		})
	})
	defaultStreamData := dsl.Type("DefaultStreamData", dsl.String, func() {
		dsl.Pattern("^[a-z]*$")
		dsl.Default("fallback")
	})
	defaultTextEvent := dsl.Type("DefaultTextEvent", func() {
		dsl.Attribute("data", defaultStreamData)
	})
	dsl.Service("params", func() {
		for _, method := range []struct {
			name  string
			type_ any
		}{
			{name: "text", type_: dsl.String},
			{name: "alias", type_: alias},
			{name: "anything", type_: dsl.Any},
			{name: "bytes", type_: dsl.Bytes},
			{name: "object", type_: object},
			{name: "array", type_: dsl.ArrayOf(dsl.String)},
			{name: "map", type_: dsl.MapOf(dsl.String, dsl.Int)},
			{name: "union", type_: union},
		} {
			dsl.Method(method.name, func() {
				dsl.Payload(method.type_)
				dsl.JSONRPC(func() {})
			})
		}
		dsl.Method("optional_text", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", dsl.String)
			})
			dsl.JSONRPC(func() {
				dsl.Body("value")
			})
		})
		for _, method := range []struct {
			name  string
			type_ any
		}{
			{name: "optional_object", type_: object},
			{name: "optional_array", type_: objects},
			{name: "optional_map", type_: objectMap},
		} {
			dsl.Method(method.name, func() {
				dsl.Payload(func() {
					dsl.Attribute("value", method.type_)
				})
				dsl.JSONRPC(func() {
					dsl.Body("value")
				})
			})
		}
		dsl.Method("optional_union", func() {
			dsl.Payload(func() {
				dsl.OneOf("value", func() {
					dsl.TypeName("OptionalRequestValue")
					dsl.Field(1, "name", dsl.String)
					dsl.Field(2, "count", dsl.Int)
				})
			})
			dsl.JSONRPC(func() {
				dsl.Body("value")
			})
		})
		dsl.Method("stream_text", func() {
			dsl.StreamingResult(dsl.String)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents()
			})
		})
		dsl.Method("stream_any", func() {
			dsl.StreamingResult(dsl.Any)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents()
			})
		})
		dsl.Method("stream_array", func() {
			dsl.StreamingResult(dsl.ArrayOf(dsl.String))
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents()
			})
		})
		dsl.Method("stream_optional_text", func() {
			dsl.StreamingResult(optionalTextEvent)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSEEventData("data")
				})
			})
		})
		dsl.Method("stream_default_text", func() {
			dsl.StreamingResult(defaultTextEvent)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSEEventData("data")
				})
			})
		})
		for _, method := range []struct {
			name  string
			type_ any
		}{
			{name: "stream_object_data", type_: objectEvent},
			{name: "stream_array_data", type_: arrayEvent},
			{name: "stream_map_data", type_: mapEvent},
			{name: "stream_union_data", type_: unionEvent},
		} {
			dsl.Method(method.name, func() {
				dsl.StreamingResult(method.type_)
				dsl.JSONRPC(func() {
					dsl.ServerSentEvents(func() {
						dsl.SSEEventData("data")
					})
				})
			})
		}
	})
}

// viewedJSONRPCCollisionPlanDSL defines two viewed methods whose authored
// names normalize to the same preferred Go helper spelling.
func viewedJSONRPCCollisionPlanDSL() {
	result := dsl.ResultType("application/vnd.retained-collision", func() {
		dsl.Attribute("id", dsl.String)
		dsl.Required("id")
		dsl.View("summary", func() {
			dsl.Attribute("id")
		})
		dsl.View("detailed", func() {
			dsl.Attribute("id")
		})
	})
	dsl.Service("collisions", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		for _, name := range []string{"fetch-item", "fetch_item"} {
			dsl.Method(name, func() {
				dsl.Result(result)
				dsl.JSONRPC(func() {})
			})
		}
	})
}

// viewedJSONRPCMappedFieldPlanDSL defines a result whose JSON-RPC response body
// contains only the id field while the selected view still accompanies the
// response and must be either the generated default view or the summary view.
func viewedJSONRPCMappedFieldPlanDSL() {
	result := dsl.ResultType("application/vnd.mapped-view", func() {
		dsl.Attribute("id", dsl.String)
		dsl.Attribute("detail", dsl.String)
		dsl.Required("id")
		dsl.View("summary", func() {
			dsl.Attribute("id")
		})
	})
	dsl.Service("mapped", func() {
		dsl.Method("fetch", func() {
			dsl.Result(result)
			dsl.JSONRPC(func() {
				dsl.Response(func() {
					dsl.Body("id")
				})
			})
		})
	})
}

// viewedJSONRPCSoleViewPlanDSL defines a result whose only legal view is the
// default view and does not repeat that choice on the method.
func viewedJSONRPCSoleViewPlanDSL() {
	result := dsl.ResultType("application/vnd.sole-view", func() {
		dsl.Attribute("id", dsl.String)
		dsl.Required("id")
		dsl.View("default", func() {
			dsl.Attribute("id")
		})
	})
	dsl.Service("sole", func() {
		dsl.Method("fetch", func() {
			dsl.Result(result)
			dsl.JSONRPC(func() {})
		})
	})
}
