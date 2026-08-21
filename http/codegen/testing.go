// This file builds HTTP code-generation analysis in tests using the same
// generation construction, planning, freezing, and rendering as production.
package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// CreateHTTPServices creates a new ServicesData instance for testing.
// Generation construction normalizes the root before any planner reads it.
func CreateHTTPServices(root *expr.RootExpr) *ServicesData {
	return NewServicesData(createServiceServices(root), root.API.HTTP)
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
	if err := example.Plan(generation); err != nil {
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
