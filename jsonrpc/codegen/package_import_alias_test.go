// This file verifies that transport packages choose import names in their own
// Go package instead of sharing names across the entire generation.
package codegen

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// TestTransportCLIPackagesChooseClientAliasesIndependently proves that the
// HTTP and JSON-RPC command packages may both use the natural client alias.
func TestTransportCLIPackagesChooseClientAliasesIndependently(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			dsl.Method("Add", func() {
				dsl.HTTP(func() { dsl.POST("/add") })
				dsl.JSONRPC(func() {})
			})
		})
	})
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	jsonHTTPPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	_, err = NewPlans(generation, PlanInput{
		Root:            root,
		Service:         servicePlan,
		HTTP:            jsonHTTPPlans[0],
		ApplicationHTTP: httpPlans[0],
	})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())

	serverName := goacodegen.SnakeCase(goacodegen.Goify(root.API.Servers[0].Name, true))
	httpCLI := generation.Package(path.Join(generation.GenPkg(), "http", "cli", serverName))
	jsonrpcCLI := generation.Package(path.Join(generation.GenPkg(), "jsonrpc", "cli", serverName))
	require.Equal(t, "calcc", httpCLI.ImportName(path.Join(generation.GenPkg(), "http", "calc", "client")))
	require.Equal(t, "calcc", jsonrpcCLI.ImportName(path.Join(generation.GenPkg(), "jsonrpc", "calc", "client")))
}
