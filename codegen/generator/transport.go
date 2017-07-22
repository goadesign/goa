package generator

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/rest"
	restdesign "goa.design/goa.v2/design/rest"
	"goa.design/goa.v2/eval"
)

// Transport iterates through the roots and returns the files needed to render
// the transport code. It returns an error if the roots slice does not include
// at least one transport design roots.
func Transport(roots ...eval.Root) ([]codegen.File, error) {
	var files []codegen.File
	for _, root := range roots {
		if r, ok := root.(*restdesign.RootExpr); ok {
			files = rest.ServerFiles(r)
			files = append(files, rest.ClientFiles(r)...)
			files = append(files, rest.ServerTypeFiles(r)...)
			files = append(files, rest.ClientTypeFiles(r)...)
			files = append(files, rest.PathFiles(r)...)
			break
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("transport: could not find transport design in DSL roots")
	}
	return files, nil
}
