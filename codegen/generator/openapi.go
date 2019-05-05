package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// OpenAPI iterates through the roots and returns the files needed to render
// the service OpenAPI spec. It produces OpenAPI specifications only if the
// roots define a HTTP service.
func OpenAPI(_ string, roots []eval.Root) ([]*codegen.File, error) {
	for _, root := range roots {
		if r, ok := root.(*expr.RootExpr); ok {
			return httpcodegen.OpenAPIFiles(r)
		}
	}
	return nil, nil
}
