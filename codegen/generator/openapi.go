// This file renders OpenAPI documents from the same evaluated roots and frozen
// service declaration data used by the transport generators.
package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// OpenAPI iterates through the roots and returns the files needed to render
// the service OpenAPI spec. It produces OpenAPI specifications only if the
// roots define a HTTP service.
func OpenAPI(generation *codegen.Generation) ([]*codegen.File, error) {
	designRoots := serviceRoots(generation.Roots)
	for _, root := range designRoots {
		if _, err := service.NewServicesData(root, generation); err != nil {
			return nil, err
		}
	}
	if len(designRoots) > 0 {
		return httpcodegen.OpenAPIFiles(designRoots[0])
	}
	return nil, nil
}
