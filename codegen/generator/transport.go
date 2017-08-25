package generator

import (
	"fmt"

	"goa.design/goa/codegen"
	"goa.design/goa/eval"
	httpcodegen "goa.design/goa/http/codegen"
	httpdesign "goa.design/goa/http/design"
)

// Transport iterates through the roots and returns the files needed to render
// the transport code. It returns an error if the roots slice does not include
// at least one transport design roots.
func Transport(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
	var files []*codegen.File
	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			files = httpcodegen.ServerFiles(genpkg, r)
			files = append(files, httpcodegen.ClientFiles(genpkg, r)...)
			files = append(files, httpcodegen.ServerTypeFiles(genpkg, r)...)
			files = append(files, httpcodegen.ClientTypeFiles(genpkg, r)...)
			files = append(files, httpcodegen.PathFiles(r)...)
			files = append(files, httpcodegen.ClientCLIFiles(genpkg, r)...)
			break
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("transport: no HTTP design found")
	}
	return files, nil
}
