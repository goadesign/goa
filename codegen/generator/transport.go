package generator

import (
	"fmt"

	"goa.design/goa/codegen"
	"goa.design/goa/eval"
	"goa.design/goa/expr"
	httpcodegen "goa.design/goa/http/codegen"
)

// Transport iterates through the roots and returns the files needed to render
// the transport code. It returns an error if the roots slice does not include
// at least one transport design roots.
func Transport(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
	var files []*codegen.File
	for _, root := range roots {
		r, ok := root.(*expr.RootExpr)
		if !ok {
			continue // could be a plugin root expression
		}

		// HTTP
		files = append(files, httpcodegen.ServerFiles(genpkg, r)...)
		files = append(files, httpcodegen.ClientFiles(genpkg, r)...)
		files = append(files, httpcodegen.ServerTypeFiles(genpkg, r)...)
		files = append(files, httpcodegen.ClientTypeFiles(genpkg, r)...)
		files = append(files, httpcodegen.PathFiles(r)...)
		files = append(files, httpcodegen.ClientCLIFiles(genpkg, r)...)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("transport: no HTTP design found")
	}
	return files, nil
}
