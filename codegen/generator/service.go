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
		r, ok := root.(*expr.RootExpr)
		if !ok {
			continue
		}
		// Create service data
		services := service.NewServicesData(r)

		for _, s := range r.Services {
			// Make sure service is first so name scope is
			// properly initialized.
			files = append(files, service.Files(genpkg, s, services, userTypePkgs)...)
			files = append(files, service.EndpointFile(genpkg, s, services), service.ClientFile(genpkg, s, services))
			if f := service.ViewsFile(genpkg, s, services); f != nil {
				files = append(files, f)
			}
			for _, f := range files {
				if len(f.SectionTemplates) > 0 {
					d := services.Get(s.Name)
					service.AddServiceDataMetaTypeImports(f.SectionTemplates[0], s, d)
					service.AddUserTypeImports(genpkg, f.SectionTemplates[0], d)
				}
			}
			f, err := service.ConvertFile(r, s, services)
			if err != nil {
				return nil, err
			}
			if f != nil {
				files = append(files, f)
			}
		}
	}
	return files, nil
}
