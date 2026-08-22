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
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	_, err = Plan(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	services := servicePlan.Services()

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
