package service

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

const (
	// clientStructName is the name of the generated client data structure.
	clientStructName = "Client"
)

// ClientFile returns the client file for the given service.
func ClientFile(_ string, service *expr.ServiceExpr, services *ServicesData) *codegen.File {
	svc := services.Get(service.Name)
	data := endpointData(svc)
	path := filepath.Join(codegen.Gendir, svc.PathName, "client.go")
	var (
		sections []*codegen.SectionTemplate
	)
	{
		imports := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "io"},
			codegen.GoaImport(""),
		}
		header := codegen.Header(service.Name+" client", svc.PkgName, imports)
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
		sections = []*codegen.SectionTemplate{header, def, init}
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
