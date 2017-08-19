package generator

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/service"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Service iterates through the roots and returns the files needed to render the
// service code. It returns an error if the roots slice does not include a goa
// design.
func Service(roots []eval.Root) ([]codegen.File, error) {
	var files []codegen.File
	for _, root := range roots {
		switch r := root.(type) {
		case *design.RootExpr:
			for _, s := range r.Services {
				// Make sure service is first so name scope is
				// properly initialized.
				files = append(files, service.File(s))
				files = append(files, service.EndpointFile(s))
			}
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("server: could not find goa design in DSL roots, vendoring issue?")
	}
	return files, nil
}
