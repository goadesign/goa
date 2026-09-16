// This file checks that gRPC definitions and their callers use the package
// names selected after preferred names are already taken.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/expr"
)

// TestGRPCPlannedNamesSurvivePackageCollisions covers endpoint functions,
// stream types, conversion constructors, and validators in both definitions
// and calls.
func TestGRPCPlannedNamesSurvivePackageCollisions(t *testing.T) {
	fixture := grpcPlanRetentionDSL(t)
	generation, services := grpcServicePlans(t, []*expr.RootExpr{fixture.root})
	clientPackage, err := generation.ClaimPackage("generated.local/gen/grpc/saved_transport/client")
	require.NoError(t, err)
	serverPackage, err := generation.ClaimPackage("generated.local/gen/grpc/saved_transport/server")
	require.NoError(t, err)
	for _, declaration := range []*codegen.NameDeclaration{
		codegen.NewExactName(codegen.NameFunction, "BuildWatchFunc"),
		codegen.NewExactName(codegen.NameFunction, "EncodeWatchRequest"),
		codegen.NewExactName(codegen.NameFunction, "DecodeWatchResponse"),
		codegen.NewExactName(codegen.NameFunction, "NewProtoWatchRequest"),
		codegen.NewExactName(codegen.NameType, "WatchClientStream"),
	} {
		require.NoError(t, clientPackage.DeclareName(declaration))
	}
	for _, declaration := range []*codegen.NameDeclaration{
		codegen.NewExactName(codegen.NameFunction, "NewWatchHandler"),
		codegen.NewExactName(codegen.NameFunction, "DecodeWatchRequest"),
		codegen.NewExactName(codegen.NameFunction, "EncodeWatchResponse"),
		codegen.NewExactName(codegen.NameFunction, "ValidateWatchRequest"),
		codegen.NewExactName(codegen.NameType, "WatchServerStream"),
	} {
		require.NoError(t, serverPackage.DeclareName(declaration))
	}
	plans, err := newPlans(
		generation,
		fixedProtobufToolResolver(),
		PlanInput{Root: fixture.root, Service: services[0]},
	)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())
	require.NoError(t, plans[0].Link())

	data, ok := plans[0].ServiceData(fixture.service)
	require.True(t, ok)
	endpoint := data.Endpoints[0]
	require.NotEqual(t, "BuildWatchFunc", endpoint.ClientBuildDeclaration.Name())
	require.NotEqual(t, "EncodeWatchRequest", endpoint.ClientEncodeDeclaration.Name())
	require.NotEqual(t, "DecodeWatchResponse", endpoint.ClientDecodeDeclaration.Name())
	require.NotEqual(t, "NewProtoWatchRequest", endpoint.Request.ClientConvert.Init.Declaration.Name())
	require.NotEqual(t, "NewWatchHandler", endpoint.ServerHandlerDeclaration.Name())
	require.NotEqual(t, "DecodeWatchRequest", endpoint.ServerDecodeDeclaration.Name())
	require.NotEqual(t, "EncodeWatchResponse", endpoint.ServerEncodeDeclaration.Name())
	require.NotEqual(t, "ValidateWatchRequest", endpoint.Request.ServerConvert.Validation.Declaration.Name())
	require.NotEqual(t, "WatchClientStream", endpoint.ClientStream.Declaration.Name())
	require.NotEqual(t, "WatchServerStream", endpoint.ServerStream.Declaration.Name())

	var source strings.Builder
	for _, selection := range []struct {
		files   []*codegen.File
		section string
	}{
		{plans[0].ClientFiles(), "remote-method-builder"},
		{plans[0].ClientFiles(), "request-encoder"},
		{plans[0].ClientFiles(), "response-decoder"},
		{plans[0].ClientFiles(), "client-endpoint-init"},
		{plans[0].ClientFiles(), "client-stream-struct-type"},
		{plans[0].ServerFiles(), "request-decoder"},
		{plans[0].ServerFiles(), "response-encoder"},
		{plans[0].ServerFiles(), "grpc-handler-init"},
		{plans[0].ServerFiles(), "server-grpc-interface"},
		{plans[0].ServerFiles(), "server-stream-struct-type"},
	} {
		sections := matchingGRPCSections(selection.files, selection.section, endpoint)
		require.NotEmpty(t, sections, "missing %s section", selection.section)
		source.WriteString(codegen.SectionsCode(t, sections))
		source.WriteString("\n")
	}
	for _, selection := range []struct {
		files   []*codegen.File
		section string
	}{
		{plans[0].ClientTypeFiles(), "client-type-init"},
		{plans[0].ServerTypeFiles(), "server-type-init"},
		{plans[0].ServerTypeFiles(), "server-validate"},
	} {
		sections := namedGRPCSections(selection.files, selection.section)
		require.NotEmpty(t, sections, "missing %s section", selection.section)
		source.WriteString(codegen.SectionsCode(t, sections))
		source.WriteString("\n")
	}
	testutil.AssertGo(t, "testdata/golden/planned_name_collisions.go.golden", strings.TrimSpace(source.String())+"\n")
}

// namedGRPCSections returns every section with name in file order.
func namedGRPCSections(files []*codegen.File, name string) []*codegen.SectionTemplate {
	result := make([]*codegen.SectionTemplate, 0, len(files))
	for _, file := range files {
		result = append(result, file.Section(name)...)
	}
	return result
}

// matchingGRPCSections returns sections for endpoint without depending on file
// order or on the number of other sections in the file.
func matchingGRPCSections(files []*codegen.File, name string, endpoint *EndpointData) []*codegen.SectionTemplate {
	var result []*codegen.SectionTemplate
	for _, file := range files {
		for _, section := range file.Section(name) {
			switch data := section.Data.(type) {
			case *EndpointData:
				if data == endpoint {
					result = append(result, section)
				}
			case *StreamData:
				if data.Endpoint == endpoint {
					result = append(result, section)
				}
			}
		}
	}
	return result
}
