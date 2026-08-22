// This file renders one service's in-process client and keeps its type imports
// scoped to that generated client file.
package service

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
)

// clientFile renders the client for the exact service retained by plan.
func clientFile(plan *Plan, facts *serviceFacts) *codegen.File {
	services := plan.Services()
	svc := services.Get(facts.name)
	data := endpointData(svc)
	path := filepath.Join(codegen.Gendir, svc.PathName, "client.go")
	var (
		sections []*codegen.SectionTemplate
	)
	{
		header := codegen.Header(facts.name+" client", svc.PkgName, facts.imports.client.specs)
		def := &codegen.SectionTemplate{
			Name:   "client-struct",
			Source: serviceTemplates.Read(serviceClientT),
			Data:   data,
		}
		init := &codegen.SectionTemplate{
			Name:   "client-init",
			Source: serviceTemplates.Read(serviceClientInitT),
			Data:   data,
		}
		sections = make([]*codegen.SectionTemplate, 0, 3+len(data.Methods))
		sections = append(sections, header, def, init)
		for _, m := range data.Methods {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-method",
				Source: serviceTemplates.Read(serviceClientMethodT),
				Data:   m,
			})
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}
