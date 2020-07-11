package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	openapiv2 "goa.design/goa/v3/http/codegen/openapi/v2"
	openapiv3 "goa.design/goa/v3/http/codegen/openapi/v3"
)

// OpenAPIFiles returns the files for the OpenAPIFile spec of the given HTTP API.
func OpenAPIFiles(root *expr.RootExpr) ([]*codegen.File, error) {
	// Only create a OpenAPI specification if there are HTTP services.
	if len(root.API.HTTP.Services) == 0 {
		return nil, nil
	}

	var files []*codegen.File
	{
		// OpenAPI v2
		fs, err := openapiv2.Files(root)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)

		// OpenAPI v3
		fs, err = openapiv3.Files(root)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}
	return files, nil
}
