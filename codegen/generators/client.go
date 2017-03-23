package generator

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	restfiles "goa.design/goa.v2/codegen/files/rest"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
	"goa.design/goa.v2/eval"
)

// Client iterates through the roots and returns the files needed to render
// the service client code. It returns an error if the roots slice does not
// include both a goa design and at least one transport design roots.
func Client(roots ...eval.Root) ([]codegen.File, error) {
	var (
		des, tran []codegen.File
	)
	for _, root := range roots {
		switch r := root.(type) {
		case *design.RootExpr:
			for _, s := range r.Services {
				des = append(des, files.Service(s))
				des = append(des, files.Endpoint(s))
			}
		case *rest.RootExpr:
			tran = append(tran, restfiles.ClientFiles(r)...)
		}
		// TBD:
		// case *rpc.RootExpr:
		// tranws = append(tran, rpccodegen.ClientFiles(r))
	}
	if len(des) == 0 {
		return nil, fmt.Errorf("could not find goa design in DSL roots, vendoring issue?")
	}
	if len(tran) == 0 {
		return nil, fmt.Errorf("could not find transport design in DSL roots")
	}
	return append(des, tran...), nil
}
