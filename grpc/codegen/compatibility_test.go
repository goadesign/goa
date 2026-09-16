// This file pins released gRPC generator entry points that plugins call after
// Goa has built the service data for one generated package.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

var (
	_ func(string, *ServicesData) []*codegen.File = ClientFiles
	_ func(string, *ServicesData) []*codegen.File = ClientCLIFiles
	_ func(string, *ServicesData) []*codegen.File = ProtoFiles
	_ func(string, *ServicesData) []*codegen.File = ServerFiles
	_ func(string, *ServicesData) []*codegen.File = ServerTypeFiles
	_ func(string, *ServicesData) []*codegen.File = ClientTypeFiles
)

// TestReleasedFileFunctionsUsePlannedPackage checks that the compatibility
// entry points render one retained plan and reject a different package.
func TestReleasedFileFunctionsUsePlannedPackage(t *testing.T) {
	services := CreateGRPCServices(RunGRPCDSL(t, testdata.UnaryRPCsDSL))
	genpkg := services.GenPkg()
	require.Len(t, ClientFiles(genpkg, services), len(clientFiles(services)))
	require.Len(t, ClientCLIFiles(genpkg, services), len(clientCLIFiles(services)))
	require.Len(t, ProtoFiles(genpkg, services), len(protoFiles(services)))
	require.Len(t, ServerFiles(genpkg, services), len(serverFiles(services)))
	require.Len(t, ServerTypeFiles(genpkg, services), len(serverTypeFiles(services)))
	require.Len(t, ClientTypeFiles(genpkg, services), len(clientTypeFiles(services)))
	require.PanicsWithValue(
		t,
		`gRPC generation package "other.local/gen" does not match planned package "generated.local/gen"`,
		func() { ClientFiles("other.local/gen", services) },
	)
}
