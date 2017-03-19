/*
Package server generates the code for a server. This includes:

    - A service package which defines the service interfaces
    - An endpoint package which defines endpoints that wrap the services methods.
    - transport packages for each of the transports defined in the design.
*/
package server

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/writers"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
	restcodegen "goa.design/goa.v2/rest/codegen"
	rest "goa.design/goa.v2/rest/design"
)

// Writers iterates through the roots and returns the writers needed to render
// the service server code. It returns an error if the roots slice does not
// include both a goa design and at least one transport design roots.
func Writers(roots ...eval.Root) ([]codegen.FileWriter, error) {
	var (
		desws, tranws []codegen.FileWriter
	)
	for _, root := range roots {
		switch r := root.(type) {
		case *design.RootExpr:
			for _, s := range r.Services {
				desws = append(desws, writers.Service(r.API, s))
				desws = append(desws, writers.Endpoint(r.API, s))
			}
		case *rest.RootExpr:
			tranws = append(tranws, restcodegen.ServerWriters(r)...)
		}
		// TBD:
		// case *rpc.RootExpr:
		// tranws = append(tranws, rpccodegen.ServerWriters(r))
	}
	if len(desws) == 0 {
		return nil, fmt.Errorf("could not find goa design in DSL roots, vendoring issue?")
	}
	if len(tranws) == 0 {
		return nil, fmt.Errorf("could not find transport design in DSL roots")
	}
	return append(desws, tranws...), nil
}
