package generator

import (
	"fmt"

	"goa.design/goa/codegen"
	"goa.design/goa/eval"
	httpcodegen "goa.design/goa/http/codegen"
	httpdesign "goa.design/goa/http/design"
)

// OpenAPI iterates through the roots and returns the files needed to render
// the service OpenAPI spec. It returns an error if the roots slice does not
// include a HTTP root.
func OpenAPI(_ string, roots []eval.Root) ([]*codegen.File, error) {
	var (
		files []*codegen.File
		err   error
	)
	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			files, err = httpcodegen.OpenAPIFiles(r)
			break
		}
	}
	if err != nil {
		return nil, err
	}
	if files == nil {
		return nil, fmt.Errorf("openapi: could not find HTTP design in DSL roots")
	}
	return files, nil
}
