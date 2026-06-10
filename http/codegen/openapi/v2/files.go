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
	spec, err := NewV2(root, root.API.Servers[0].Hosts[0])
	if err != nil {
		return nil, err
	}
	return openapi.Files(spec, root.API.Meta, "openapi", path), nil
}
