package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv2 "goa.design/goa/v3/http/codegen/openapi/v2"
	openapiv3 "goa.design/goa/v3/http/codegen/openapi/v3"
)

// OpenAPIFiles returns the files for the OpenAPI specs of the given HTTP API.
// The "openapi:versions" API meta selects the generated specification
// versions and the "openapi:path:<version>" API meta overrides their output
// paths, see openapi.Specs.
func OpenAPIFiles(root *expr.RootExpr) ([]*codegen.File, error) {
	// Only create a OpenAPI specification if there are HTTP services.
	if len(root.API.HTTP.Services) == 0 {
		return nil, nil
	}

	specs, err := openapi.Specs(root.API.Meta)
	if err != nil {
		return nil, err
	}
	var files []*codegen.File
	for _, spec := range specs {
		var fs []*codegen.File
		switch spec.Version {
		case openapi.Version20:
			fs, err = openapiv2.Files(root, spec.Path)
			if err != nil {
				return nil, err
			}
		default: // Version30, Version32
			fs = openapiv3.Files(root, spec.Version, spec.Path)
		}
		files = append(files, fs...)
	}
	return files, nil
}
