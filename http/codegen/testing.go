// This file builds HTTP code-generation analysis in tests using the same
// normalize, plan, freeze, and render lifecycle as production generation.
package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// CreateHTTPServices creates a new ServicesData instance for testing. The
// root is normalized first like the production Generate flow does before the
// generators read the design.
func CreateHTTPServices(root *expr.RootExpr) *ServicesData {
	codegen.NormalizeRoot(root)
	return NewServicesData(createServiceServices(root), root.API.HTTP)
}

// createServiceServices performs the complete package declaration lifecycle
// required by transport test helpers.
func createServiceServices(root *expr.RootExpr) *service.ServicesData {
	generation := codegen.NewGeneration("/", []eval.Root{root})
	if err := service.Plan(root, generation); err != nil {
		panic(err)
	}
	if err := Plan(generation); err != nil {
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
