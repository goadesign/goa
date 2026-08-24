// This file verifies that service declarations and transport callers keep the
// same Go names when authored types claim the usual generated spellings.
package codegen

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/codegentest"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	gencodegengrpc "goa.design/goa/v3/grpc/codegen"
)

type (
	// plannedNameSection identifies one complete generated section or one file
	// body included in the cross-transport golden output.
	plannedNameSection struct {
		label string
		files []*codegen.File
		file  string
		name  string
		whole bool
	}
)

// TestPlannedServiceNamesUsedAcrossTransports renders each definition and use
// from the same generation so a changed or rebuilt name breaks one fixture.
func TestPlannedServiceNamesUsedAcrossTransports(t *testing.T) {
	root := expr.RunDSL(t, plannedServiceNameUsesDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	grpcPlans, err := gencodegengrpc.NewPlans(generation, gencodegengrpc.PlanInput{
		Root:    root,
		Service: servicePlan,
	})
	require.NoError(t, err)
	examplePlan, err := example.NewPlan(generation, servicePlan)
	require.NoError(t, err)
	httpExamples, err := NewExamplePlan(httpPlans[0], examplePlan)
	require.NoError(t, err)
	grpcExamples, err := gencodegengrpc.NewExamplePlan(grpcPlans[0], examplePlan)
	require.NoError(t, err)

	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	require.NoError(t, grpcPlans[0].Link())

	serviceData := servicePlan.Services().Get("Collisions")
	require.NotNil(t, serviceData)
	method := serviceData.Methods[0]
	require.Equal(t, "Endpoints2", serviceData.EndpointsDeclaration.Name())
	require.Equal(t, "MethodNames2", serviceData.MethodNamesDeclaration.Name())
	require.Equal(t, "ClientInterceptors2", serviceData.ClientInterceptorsDeclaration.Name())
	require.Equal(t, "WrapReadClientEndpoint2", method.ClientEndpointWrapperDeclaration.Name())

	serviceFiles, err := service.Files(servicePlan)
	require.NoError(t, err)
	sections := []plannedNameSection{
		{
			label: "service method names definition",
			files: serviceFiles,
			file:  "service.go",
			name:  "service",
		},
		{
			label: "service endpoints definition",
			files: serviceFiles,
			file:  "endpoints.go",
			name:  "endpoints-struct",
		},
		{
			label: "client interceptors definition",
			files: serviceFiles,
			file:  "client_interceptors.go",
			name:  "client-interceptors-type",
		},
		{
			label: "client endpoint wrapper definition",
			files: serviceFiles,
			file:  "client_interceptors.go",
			name:  "client-wrapper",
		},
		{
			label: "HTTP server endpoints use",
			files: httpPlans[0].ServerFiles(),
			file:  "server.go",
			name:  "server-init",
		},
		{
			label: "HTTP server method names use",
			files: httpPlans[0].ServerFiles(),
			file:  "server.go",
			name:  "server-method-names",
		},
		{
			label: "HTTP command parser",
			files: httpPlans[0].ClientCLIFiles(),
			file:  "cli.go",
			name:  "parse-endpoint",
		},
		{
			label: "HTTP example client interceptor use",
			files: httpExamples.CLIFiles(),
			file:  "http.go",
			whole: true,
		},
		{
			label: "gRPC server endpoints use",
			files: grpcPlans[0].ServerFiles(),
			file:  "server.go",
			name:  "server-init",
		},
		{
			label: "gRPC example server endpoints use",
			files: grpcExamples.ServerFiles(),
			file:  "grpc.go",
			whole: true,
		},
		{
			label: "gRPC command parser",
			files: grpcPlans[0].ClientCLIFiles(),
			file:  "cli.go",
			name:  "parse-endpoint-grpc",
		},
	}

	var source strings.Builder
	for _, section := range sections {
		source.WriteString("===== ")
		source.WriteString(section.label)
		source.WriteString(" =====\n")
		source.WriteString(plannedNameSectionCode(t, section))
		source.WriteString("\n")
	}
	testutil.AssertString(t, "testdata/golden/planned_service_name_uses.go.golden", source.String())
}

// plannedNameSectionCode renders either one complete section or the complete
// body of a file whose opening and closing statements span several sections.
func plannedNameSectionCode(t *testing.T, section plannedNameSection) string {
	t.Helper()
	if !section.whole {
		matches := codegentest.Sections(section.files, section.file, section.name)
		require.Len(t, matches, 1, section.label)
		return codegen.SectionCode(t, matches[0])
	}
	for _, file := range section.files {
		if filepath.Base(file.Path) == section.file {
			var source bytes.Buffer
			for _, part := range file.SectionTemplates[1:] {
				require.NoError(t, part.Write(&source))
			}
			return codegen.FormatTestCode(t, "package foo\n"+source.String())
		}
	}
	require.Fail(t, "missing generated file", section.label)
	return ""
}

// plannedServiceNameUsesDSL makes authored types claim four names normally
// chosen for the generated service and interceptor code.
func plannedServiceNameUsesDSL() {
	endpointType := dsl.Type("Endpoints", dsl.String)
	methodNamesType := dsl.Type("MethodNames", dsl.String)
	interceptorsType := dsl.Type("ClientInterceptors", dsl.String)
	wrapperType := dsl.Type("WrapReadClientEndpoint", dsl.String)
	trace := dsl.Interceptor("Trace")

	dsl.API("Name Test", func() {
		dsl.Server("test", func() {
			dsl.Host("development", func() {
				dsl.URI("http://localhost:80")
			})
		})
	})
	dsl.Service("Collisions", func() {
		dsl.ClientInterceptor(trace)
		dsl.Method("Read", func() {
			dsl.Payload(func() {
				dsl.Field(1, "endpoint", endpointType)
				dsl.Field(2, "method_names", methodNamesType)
				dsl.Field(3, "interceptors", interceptorsType)
				dsl.Field(4, "wrapper", wrapperType)
			})
			dsl.Result(dsl.String)
			dsl.HTTP(func() {
				dsl.POST("/read")
			})
			dsl.GRPC(func() {
			})
		})
	})
}
