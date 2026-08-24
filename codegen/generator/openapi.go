// This file builds each requested OpenAPI document while the evaluated design
// is available. File generation later returns the documents already built.
package generator

import (
	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// openAPIFiles returns the OpenAPI files built during planning.
func openAPIFiles(plan *Plan) ([]*codegen.File, error) {
	if len(plan.openapiReplacements) > 0 {
		var files []*codegen.File
		for _, openapi := range plan.openapiReplacements {
			files = append(files, openapi.Files()...)
		}
		return files, nil
	}
	return plan.openapi.Files(), nil
}

// planOpenAPIData builds the OpenAPI files for the application's design root.
// Later roots contain generated support services and do not describe another
// application API.
func planOpenAPIData(plan *Plan) error {
	roots := serviceRoots(plan.Generation().Roots())
	if len(roots) == 0 {
		plan.openapi = new(httpcodegen.OpenAPIPlan)
		plan.openapiRoot = nil
		return nil
	}
	openapi, err := httpcodegen.NewOpenAPIPlan(roots[0], plan.exampleGenerator(roots[0]))
	if err != nil {
		return err
	}
	plan.openapi = openapi
	plan.openapiRoot = roots[0]
	return nil
}
