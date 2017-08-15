package generator

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/eval"
	httpcodegen "goa.design/goa.v2/http/codegen"
	httpdesign "goa.design/goa.v2/http/design"
)

// Transport iterates through the roots and returns the files needed to render
// the transport code. It returns an error if the roots slice does not include
// at least one transport design roots.
func Transport(roots ...eval.Root) ([]codegen.File, error) {
	var files []codegen.File
	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			files = httpcodegen.ServerFiles(r)
			files = append(files, httpcodegen.ClientFiles(r)...)
			files = append(files, httpcodegen.ServerTypeFiles(r)...)
			files = append(files, httpcodegen.ClientTypeFiles(r)...)
			files = append(files, httpcodegen.PathFiles(r)...)
			files = append(files, httpcodegen.ClientCLIFiles(r)...)
			break
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("transport: no HTTP design found")
	}
	return files, nil
}
