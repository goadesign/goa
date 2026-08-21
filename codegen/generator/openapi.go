// This file renders OpenAPI documents from the same evaluated roots and frozen
// service declaration data used by the transport generators.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// openAPIFiles returns OpenAPI files described by plan's frozen package
// declarations and run-owned example state.
func openAPIFiles(plan *Plan) ([]*codegen.File, error) {
	generation := plan.Generation()
	designRoots := serviceRoots(generation.Roots())
	for _, root := range designRoots {
		if _, err := service.NewServicesData(root, generation, plan.exampleGenerator(root)); err != nil {
			return nil, err
		}
	}
	if len(designRoots) > 0 {
		root := designRoots[0]
		return httpcodegen.OpenAPIFiles(root, plan.exampleGenerator(root))
	}
	return nil, nil
}
