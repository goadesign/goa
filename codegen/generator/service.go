package generator

import (
	"fmt"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service"
	"goa.design/goa/expr"
	"goa.design/goa/eval"
)

// Service iterates through the roots and returns the files needed to render the
// service code. It returns an error if the roots slice does not include a goa
// design.
func Service(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
	var files []*codegen.File
	for _, root := range roots {
		switch r := root.(type) {
		case *expr.RootExpr:
			for _, s := range r.Services {
				// Make sure service is first so name scope is
				// properly initialized.
				files = append(files, service.File(genpkg, s))
				files = append(files, service.EndpointFile(genpkg, s))
				files = append(files, service.ClientFile(s))
				if f := service.ViewsFile(genpkg, s); f != nil {
					files = append(files, f)
				}
				f, err := service.ConvertFile(r, s)
				if err != nil {
					return nil, err
				}
				if f != nil {
					files = append(files, f)
				}
			}
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("design must define at least one service")
	}
	return files, nil
}
