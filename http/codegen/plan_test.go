// This file verifies HTTP import planning participates in the shared
// generation lifecycle before service aliases are frozen.
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

func TestPlanReservesStaticAliasesBeforeFreeze(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Path", func() {
			dsl.Method("Read", func() {})
		})
	})
	generation := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, service.Plan(root, generation))
	require.NoError(t, Plan(generation))
	require.NoError(t, generation.Freeze())
	services, err := service.NewServicesData(root, generation)
	require.NoError(t, err)

	require.Equal(t, "path2", services.ServiceImport("Path").Name)
}

func TestPlanRejectsFrozenGeneration(t *testing.T) {
	generation := codegen.NewGeneration("generated.local/gen", nil)
	require.NoError(t, generation.Freeze())

	require.Error(t, Plan(generation))
}

// TestPlanReservesGeneratedHTTPPackages verifies that client, server, and CLI
// packages receive aliases from the generation catalog before it freezes.
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
	generation := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, service.Plan(root, generation))
	require.NoError(t, Plan(generation))
	require.NoError(t, generation.Freeze())
	services, err := service.NewServicesData(root, generation)
	require.NoError(t, err)

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
