package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// pathFiles builds the service path files read by Plan.Link.
func pathFiles(data *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, 2*len(data.Expressions.Services))
	for i := 0; i < len(data.Expressions.Services); i++ {
		fw[i*2] = serverPath(data.Expressions.Services[i], data)
		fw[i*2+1] = clientPath(data.Expressions.Services[i], data)
	}
	return fw
}

// serverPath returns the server file containing the request path constructors
// for the given service.
func serverPath(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	sd := services.Get(svc.Name())
	path := filepath.Join(codegen.Gendir, services.dir(), sd.Service.PathName, "server", "paths.go")
	return &codegen.File{Path: path, SectionTemplates: pathSections(svc, "server", services)}
}

// clientPath returns the client file containing the request path constructors
// for the given service.
func clientPath(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	sd := services.Get(svc.Name())
	path := filepath.Join(codegen.Gendir, services.dir(), sd.Service.PathName, "client", "paths.go")
	return &codegen.File{Path: path, SectionTemplates: pathSections(svc, "client", services)}
}

// pathSections returns the sections of the file of the pkg package that
// contains the request path constructors for the given service.
func pathSections(svc *expr.HTTPServiceExpr, pkg string, services *ServicesData) []*codegen.SectionTemplate {
	title := fmt.Sprintf("%s request path constructors for the %s service.", services.label(), svc.Name())
	sections := make([]*codegen.SectionTemplate, 0, 1+len(svc.HTTPEndpoints))
	sections = append(sections,
		codegen.Header(title, pkg, []*codegen.ImportSpec{
			{Path: "fmt"},
			{Path: "net/url"},
			{Path: "strconv"},
			{Path: "strings"},
		}),
	)
	sdata := services.Get(svc.Name())
	for _, e := range svc.HTTPEndpoints {
		data := struct {
			*EndpointData
			Client bool
		}{
			EndpointData: sdata.Endpoint(e.Name()),
			Client:       pkg == "client",
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "path",
			Source: httpTemplates.Read(pathT),
			Data:   data,
		})
	}

	return sections
}
