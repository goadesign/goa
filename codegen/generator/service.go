// This file assembles service-owned generated files after every participating
// Goa design root has planned and frozen its package declarations.
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
func Service(generation *codegen.Generation) ([]*codegen.File, error) {
	var files []*codegen.File
	designRoots := serviceRoots(generation.Roots)
	analyses := make([]*service.ServicesData, len(designRoots))
	for i, r := range designRoots {
		services, err := service.NewServicesData(r, generation)
		if err != nil {
			return nil, err
		}
		analyses[i] = services

		for _, s := range r.Services {
			endpointFiles := []*codegen.File{
				service.EndpointFile(generation.GenPkg, s, services),
				service.ClientFile(generation.GenPkg, s, services),
			}
			files = append(files, endpointFiles...)

			if f := service.ViewsFile(generation.GenPkg, s, services); f != nil {
				files = append(files, f)
			}
			convFiles, err := service.ConvertFiles(r, s, services)
			if err != nil {
				return nil, err
			}
			files = append(files, convFiles...)
		}
	}
	svcFiles := service.Files(generation.GenPkg, analyses)
	return append(svcFiles, files...), nil
}

// planServiceData declares service-owned generated package types for every Goa
// design root in generation.
func planServiceData(generation *codegen.Generation) error {
	for _, root := range serviceRoots(generation.Roots) {
		if err := service.Plan(root, generation); err != nil {
			return err
		}
	}
	return nil
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
