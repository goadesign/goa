// This file verifies standalone JSON-RPC planning includes the HTTP codecs and
// helpers that JSON-RPC rendering reuses.
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
	require.NoError(t, service.Plan(root, generation))
	require.NoError(t, Plan(generation))
	require.NoError(t, generation.Freeze())
	services, err := service.NewServicesData(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)

	require.Equal(t, "uuid2", services.ServiceImport("UUID").Name)
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
	require.NoError(t, service.Plan(root, generation))
	require.NoError(t, Plan(generation))
	require.NoError(t, generation.Freeze())
	services, err := service.NewServicesData(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)

	client := services.PackageImport("generated.local/gen/jsonrpc/foo/client")
	server := services.PackageImport("generated.local/gen/jsonrpc/foo/server")
	cli := services.PackageImport(path.Join(
		"generated.local/gen/jsonrpc/cli",
		codegen.SnakeCase(codegen.Goify(root.API.Servers[0].Name, true)),
	))
	require.NotEqual(t, services.ServiceImport("Fooc").Name, client.Name)
	require.NotEqual(t, services.ServiceImport("Foojssvr").Name, server.Name)
	require.Equal(t, "cli", cli.Name)
}
