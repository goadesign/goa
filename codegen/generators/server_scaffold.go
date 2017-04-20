package generator

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	restfiles "goa.design/goa.v2/codegen/files/rest"
	"goa.design/goa.v2/design/rest"
	"goa.design/goa.v2/eval"
)

// ServerScaffold iterates through the roots and returns the files needed to render
// the service tool code. It returns an error if the roots slice does not
// include at least one transport design roots.
func ServerScaffold(roots ...eval.Root) ([]codegen.File, error) {
	var (
		tran []codegen.File
	)
	for _, root := range roots {
		switch r := root.(type) {
		case *rest.RootExpr:
			tran = append(tran, restfiles.ServerScaffoldFiles(r)...)
		}
		// TBD:
		// case *rpc.RootExpr:
		// tranws = append(tranws, rpccodegen.ServerScaffoldFiles(r))
	}
	if len(tran) == 0 {
		return nil, fmt.Errorf("could not find transport design in DSL roots")
	}
	return tran, nil
}
