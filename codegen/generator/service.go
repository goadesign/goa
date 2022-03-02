package generator

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Service iterates through the roots and returns the files needed to render
// the service code. It returns an error if the roots slice does not include
// a goa design.
func Service(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
	var files []*codegen.File
	var userTypePkgs = make(map[string][]string)
	for _, root := range roots {
		switch r := root.(type) {
		case *expr.RootExpr:
			for _, s := range r.Services {
				// Make sure service is first so name scope is
				// properly initialized.
				files = append(files, service.Files(genpkg, s, userTypePkgs)...)
				files = append(files, service.EndpointFile(genpkg, s))
				files = append(files, service.ClientFile(genpkg, s))
				if f := service.ViewsFile(genpkg, s); f != nil {
					files = append(files, f)
				}
				for _, f := range files {
					if len(f.SectionTemplates) > 0 {
						service.AddServiceDataMetaTypeImports(f.SectionTemplates[0], s)
					}
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
	return files, nil
}
