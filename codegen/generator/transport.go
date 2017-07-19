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
			files = rest.Servers(r)
			files = append(files, rest.Clients(r)...)
			files = append(files, rest.ServerTypes(r)...)
			files = append(files, rest.ClientTypes(r)...)
			files = append(files, rest.Paths(r)...)
			break
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("transport: could not find transport design in DSL roots")
	}
	return files, nil
}
