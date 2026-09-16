// This file builds Swagger 2.0 JSON and YAML files from one HTTP design. Each
// example comes from the request or response described in the file.
package openapiv2

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
)

// Files returns the Swagger 2.0 specification files in JSON and YAML formats.
// path is the output path of the files relative to the gen directory, without
// extension.
func Files(root *expr.RootExpr, path string) ([]*codegen.File, error) {
	return FilesWithValues(
		root,
		path,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
		openapi.Values{},
	)
}

// FilesWithValues returns Swagger 2.0 files using values in place of matching
// titles, descriptions, and examples from the evaluated design. The generator
// supplies examples for attributes that have no matching value.
func FilesWithValues(root *expr.RootExpr, path string, generator *expr.ExampleGenerator, values openapi.Values) ([]*codegen.File, error) {
	spec, err := NewV2WithValues(root, root.API.Servers[0].Hosts[0], generator, values)
	if err != nil {
		return nil, err
	}
	return openapi.Files(spec, root.API.Meta, "openapi", path), nil
}
