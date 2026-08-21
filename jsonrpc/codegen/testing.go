// This file builds JSON-RPC code-generation analysis in tests using the same
// generation construction, planning, freezing, and rendering as production.
package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// CreateJSONRPCServices creates a new ServicesData instance for JSON-RPC
// testing. Generation construction normalizes the root before any planner
// reads it.
func CreateJSONRPCServices(root *expr.RootExpr) *httpcodegen.ServicesData {
	services := createServiceServices(root)
	return httpcodegen.NewJSONRPCServicesData(services, &root.API.JSONRPC.HTTPExpr)
}

// createServiceServices performs the complete package declaration lifecycle
// required by transport test helpers.
func createServiceServices(root *expr.RootExpr) *service.ServicesData {
	generation, err := codegen.NewGeneration("/", []eval.Root{root})
	if err != nil {
		panic(err)
	}
	if err := service.Plan(root, generation); err != nil {
		panic(err)
	}
	if err := Plan(generation); err != nil {
		panic(err)
	}
	if err := generation.Freeze(); err != nil {
		panic(err)
	}
	services, err := service.NewServicesData(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	if err != nil {
		panic(err)
	}
	return services
}
