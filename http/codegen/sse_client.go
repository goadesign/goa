package codegen

import (
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// sseClientFile returns the file implementing the SSE client code for SSE endpoints if any.
// Relies on SSEData (ed.SSE) for all codegen needs.
func sseClientFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if !HasSSE(data) {
		return nil
	}
	path := filepath.Join(codegen.Gendir, "http", data.Service.PathName, "client", "sse.go")
	outputPackage := generatedFileOutputPackage(services, path)
	data = serviceDataForOutput(data, services, outputPackage)
	tmplSections := sseClientTemplateSections(data)
	sections := make([]*codegen.SectionTemplate, 0, 1+len(tmplSections))
	sections = append(sections,
		plannedFileHeader(
			"sse-client",
			"client",
			path,
			services,
		),
	)
	sections = append(sections, tmplSections...) // add SSE client methods
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// sseClientTemplateSections returns section templates for SSE client endpoints.
func sseClientTemplateSections(data *ServiceData) []*codegen.SectionTemplate {
	sections := make([]*codegen.SectionTemplate, 0)
	for _, ed := range data.Endpoints {
		if ed.SSE == nil {
			continue
		}
		funcs := sseTemplateFuncs()
		funcs["dict"] = dict
		funcs["goTypeRef"] = func(dataType expr.DataType) string {
			return data.Scope.GoTypeRef(&expr.AttributeExpr{Type: dataType})
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "client-sse",
			Source:  httpTemplates.Read(clientSseT, sseParseP, queryTypeConversionP, elementSliceConversionP, sliceItemConversionP),
			Data:    ed,
			FuncMap: funcs,
		})
	}
	return sections
}
