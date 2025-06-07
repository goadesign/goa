package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// PathFiles returns the service path files.
func PathFiles(services *ServicesData) []*codegen.File {
	root := services.Root
	fw := make([]*codegen.File, 2*len(root.API.HTTP.Services))
	for i := 0; i < len(root.API.HTTP.Services); i++ {
		fw[i*2] = serverPath(root.API.HTTP.Services[i], services)
		fw[i*2+1] = clientPath(root.API.HTTP.Services[i], services)
	}
	return fw
}

// serverPath returns the server file containing the request path constructors
// for the given service.
func serverPath(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	sd := services.Get(svc.Name())
	path := filepath.Join(codegen.Gendir, "http", sd.Service.PathName, "server", "paths.go")
	return &codegen.File{Path: path, SectionTemplates: pathSections(svc, "server", services)}
}

// clientPath returns the client file containing the request path constructors
// for the given service.
func clientPath(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	sd := services.Get(svc.Name())
	path := filepath.Join(codegen.Gendir, "http", sd.Service.PathName, "client", "paths.go")
	return &codegen.File{Path: path, SectionTemplates: pathSections(svc, "client", services)}
}

// pathSections returns the sections of the file of the pkg package that
// contains the request path constructors for the given service.
func pathSections(svc *expr.HTTPServiceExpr, pkg string, services *ServicesData) []*codegen.SectionTemplate {
	title := fmt.Sprintf("HTTP request path constructors for the %s service.", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, pkg, []*codegen.ImportSpec{
			{Path: "fmt"},
			{Path: "net/url"},
			{Path: "strconv"},
			{Path: "strings"},
		}),
	}
	sdata := services.Get(svc.Name())
	for _, e := range svc.HTTPEndpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "path",
			Source: readTemplate("path"),
			Data:   sdata.Endpoint(e.Name()),
		})
	}

	return sections
}
