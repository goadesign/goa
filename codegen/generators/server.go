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

// Server iterates through the roots and returns the files needed to render
// the service server code. It returns an error if the roots slice does not
// include both a goa design and at least one transport design roots.
func Server(roots ...eval.Root) ([]codegen.File, error) {
	var (
		des, tran []codegen.File
	)
	for _, root := range roots {
		switch r := root.(type) {
		case *design.RootExpr:
			for _, s := range r.Services {
				// Make sure service is first so name scope is
				// properly initialized.
				des = append(des, files.Service(s))
				des = append(des, files.Endpoint(s))
			}
		case *rest.RootExpr:
			tran = append(tran, restfiles.ServerFiles(r)...)
			tran = append(tran, restfiles.PathFile(r))
			tran = append(tran, restfiles.PathFile(r))
		}
		// TBD:
		// case *rpc.RootExpr:
		// tranws = append(tranws, rpccodegen.ServerFiles(r))
	}
	if len(des) == 0 {
		return nil, fmt.Errorf("could not find goa design in DSL roots, vendoring issue?")
	}
	if len(tran) == 0 {
		return nil, fmt.Errorf("could not find transport design in DSL roots")
	}
	return append(des, tran...), nil
}
