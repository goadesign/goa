/*
Package openapi generates the OpenAPI specification of the service.
*/
package openapi

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/writers"
	"goa.design/goa.v2/eval"
	rest "goa.design/goa.v2/rest/design"
)

// Writers iterates through the roots and returns the writers needed to render
// the service OpenAPI spec. It returns an error if the roots slice does not
// include a rest root.
func Writers(roots ...eval.Root) ([]codegen.FileWriter, error) {
	var (
		ws  codegen.FileWriter
		err error
	)
	for _, root := range roots {
		switch r := root.(type) {
		case *rest.RootExpr:
			ws, err = writers.OpenAPI(r)
		}
	}
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, fmt.Errorf("could not find rest design in DSL roots")
	}
	return []codegen.FileWriter{ws}, nil
}
