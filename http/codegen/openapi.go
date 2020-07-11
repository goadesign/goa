package codegen

import (
	"encoding/json"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v2"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/http/codegen/openapi"
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
