package generator

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/http"
	httpdesign "goa.design/goa.v2/design/http"
	"goa.design/goa.v2/eval"
)

// Transport iterates through the roots and returns the files needed to render
// the transport code. It returns an error if the roots slice does not include
// at least one transport design roots.
func Transport(roots ...eval.Root) ([]codegen.File, error) {
	var files []codegen.File
	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			files = http.ServerFiles(r)
			files = append(files, http.ClientFiles(r)...)
			files = append(files, http.ServerTypeFiles(r)...)
			files = append(files, http.ClientTypeFiles(r)...)
			files = append(files, http.PathFiles(r)...)
			break
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("transport: could not find transport design in DSL roots")
	}
	return files, nil
}
