// This file builds JSON-RPC code-generation analysis in tests using the same
// normalize, plan, freeze, and render lifecycle as production generation.
package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// CreateJSONRPCServices creates a new ServicesData instance for JSON-RPC
// testing. The root is normalized first like the production Generate flow
// does before the generators read the design.
func CreateJSONRPCServices(root *expr.RootExpr) *httpcodegen.ServicesData {
	codegen.NormalizeRoot(root)
	services := createServiceServices(root)
	return httpcodegen.NewJSONRPCServicesData(services, &root.API.JSONRPC.HTTPExpr)
}

// createServiceServices performs the complete package declaration lifecycle
// required by transport test helpers.
func createServiceServices(root *expr.RootExpr) *service.ServicesData {
	generation := codegen.NewGeneration("goa.design/goa/example", []eval.Root{root})
	if err := service.Plan(root, generation); err != nil {
		panic(err)
	}
	if err := generation.Freeze(); err != nil {
		panic(err)
	}
	services, err := service.NewServicesData(root, generation)
	if err != nil {
		panic(err)
	}
	return services
}
