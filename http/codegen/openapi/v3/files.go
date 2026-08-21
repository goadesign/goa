// This file renders a prepared HTTP design as OpenAPI 3 JSON and YAML files.
// Callers provide the run-owned example coordinator, and the builder derives
// every example stream from the HTTP expression represented in the document.
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
func Files(root *expr.RootExpr, ver openapi.Version, path string, generator *expr.ExampleGenerator) []*codegen.File {
	return openapi.Files(New(root, ver, generator), root.API.Meta, "openapi_v3", path)
}
