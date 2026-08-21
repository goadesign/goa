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
	designRoots := serviceRoots(roots)
	servicesByRoot := service.NewServicesDataForRoots(designRoots)
	for _, r := range designRoots {
		services := servicesByRoot[r]

		for _, s := range r.Services {
			d := services.Get(s.Name)
			service.SetUserTypeImports(genpkg, d)

			// Make sure service is first so name scope is
			// properly initialized.
			svcFiles := service.Files(genpkg, s, services, userTypePkgs)
			addServiceImports(svcFiles, d)
			files = append(files, svcFiles...)

			endpointFiles := []*codegen.File{
				service.EndpointFile(genpkg, s, services),
				service.ClientFile(genpkg, s, services),
			}
			addServiceImports(endpointFiles, d)
			files = append(files, endpointFiles...)

			if f := service.ViewsFile(genpkg, s, services); f != nil {
				addServiceImports([]*codegen.File{f}, d)
				files = append(files, f)
			}
			convFiles, err := service.ConvertFiles(r, s, services)
			if err != nil {
				return nil, err
			}
			files = append(files, convFiles...)
		}
	}
	return files, nil
}

// serviceRoots returns every Goa design root that emits files into the same
// generated package tree.
func serviceRoots(roots []eval.Root) []*expr.RootExpr {
	var designRoots []*expr.RootExpr
	for _, root := range roots {
		if design, ok := root.(*expr.RootExpr); ok {
			designRoots = append(designRoots, design)
		}
	}
	return designRoots
}

func addServiceImports(files []*codegen.File, d *service.Data) {
	for _, f := range files {
		if len(f.SectionTemplates) == 0 {
			continue
		}
		service.AddServiceDataMetaTypeImports(f.SectionTemplates[0], d)
		service.AddUserTypeImports(f.SectionTemplates[0], d)
	}
}

func addServicesImports(files []*codegen.File, services *service.ServicesData, svcs []*expr.ServiceExpr) {
	for _, s := range svcs {
		addServiceImports(files, services.Get(s.Name))
	}
}

func addMetaTypeImports(files []*codegen.File, d *service.Data) {
	for _, f := range files {
		if len(f.SectionTemplates) == 0 {
			continue
		}
		service.AddServiceDataMetaTypeImports(f.SectionTemplates[0], d)
	}
}

func addServicesMetaTypeImports(files []*codegen.File, services *service.ServicesData, svcs []*expr.ServiceExpr) {
	for _, s := range svcs {
		addMetaTypeImports(files, services.Get(s.Name))
	}
}
