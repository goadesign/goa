// This file builds OpenAPI 3 JSON and YAML files from one HTTP design. Each
// example comes from the request or response described in the file.
package openapiv3

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

// Files returns the OpenAPI specification files conforming to the given
// version (openapi.Version30 or openapi.Version32) in JSON and YAML formats.
// path is the output path of the files relative to the gen directory, without
// extension.
func Files(root *expr.RootExpr, ver openapi.Version, path string) []*codegen.File {
	return FilesWithValues(
		root,
		ver,
		path,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		openapi.Values{},
	)
}

// FilesWithValues returns OpenAPI files using values in place of matching
// titles, descriptions, and examples from the evaluated design. The generator
// supplies examples for attributes that have no matching value.
func FilesWithValues(root *expr.RootExpr, ver openapi.Version, path string, generator *expr.ExampleGenerator, values openapi.Values) []*codegen.File {
	return openapi.Files(NewWithValues(root, ver, generator, values), root.API.Meta, "openapi_v3", path)
}
