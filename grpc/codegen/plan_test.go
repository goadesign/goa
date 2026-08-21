// This file verifies gRPC planning reserves static and generated package
// imports before the generation catalog freezes.
package codegen

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestPlanReservesGeneratedGRPCPackages verifies that client, server,
// protobuf, and CLI packages consume exact frozen import records.
func TestPlanReservesGeneratedGRPCPackages(t *testing.T) {
	root := expr.RunDSL(t, func() {
		for _, name := range []string{"Foo", "Fooc", "Foosvr"} {
			dsl.Service(name, func() {
				dsl.Method("Read", func() { dsl.GRPC(func() {}) })
			})
		}
	})
	generation := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, service.Plan(root, generation))
	require.NoError(t, Plan(generation))
	require.NoError(t, generation.Freeze())
	services, err := service.NewServicesData(root, generation)
	require.NoError(t, err)

	client := services.PackageImport("generated.local/gen/grpc/foo/client")
	server := services.PackageImport("generated.local/gen/grpc/foo/server")
	protobuf := services.PackageImport("generated.local/gen/grpc/foo/pb")
	cli := services.PackageImport(path.Join(
		"generated.local/gen/grpc/cli",
		codegen.SnakeCase(codegen.Goify(root.API.Servers[0].Name, true)),
	))
	require.NotEqual(t, services.ServiceImport("Fooc").Name, client.Name)
	require.NotEqual(t, services.ServiceImport("Foosvr").Name, server.Name)
	require.NotEmpty(t, protobuf.Name)
	require.NotEmpty(t, cli.Name)
}
