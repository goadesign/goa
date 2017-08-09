package codegen

import (
	"fmt"
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	httpdesign "goa.design/goa.v2/http/design"
)

var (
	pathFuncMap = template.FuncMap{"add": codegen.Add}
	pathTmpl    = template.Must(template.New("path").Funcs(pathFuncMap).Parse(pathT))
)

// PathFiles returns the service path files.
func PathFiles(root *httpdesign.RootExpr) []codegen.File {
	fw := make([]codegen.File, 2*len(root.HTTPServices))
	for i := 0; i < len(root.HTTPServices); i++ {
		fw[i*2] = serverPath(root.HTTPServices[i])
		fw[i*2+1] = clientPath(root.HTTPServices[i])
	}
	return fw
}

// serverPath returns the server file containing the request path constructors
// for the given service.
func serverPath(svc *httpdesign.ServiceExpr) codegen.File {
	path := filepath.Join("http", codegen.SnakeCase(svc.Name()), "server", "paths.go")
	return codegen.NewSource(path, pathSections(svc, "server"))
}

// clientPath returns the client file containing the request path constructors
// for the given service.
func clientPath(svc *httpdesign.ServiceExpr) codegen.File {
	path := filepath.Join("http", codegen.SnakeCase(svc.Name()), "client", "paths.go")
	return codegen.NewSource(path, pathSections(svc, "client"))
}

// pathSections returns the sections of the file of the pkg package that
// contains the request path constructors for the given service.
func pathSections(svc *httpdesign.ServiceExpr, pkg string) codegen.SectionsFunc {
	return func(_ string) []*codegen.Section {
		title := fmt.Sprintf("HTTP request path constructors for the %s service.", svc.Name())
		s := []*codegen.Section{
			codegen.Header(title, pkg, []*codegen.ImportSpec{
				{Path: "fmt"},
				{Path: "net/url"},
				{Path: "strconv"},
				{Path: "strings"},
			}),
		}
		sdata := HTTPServices.Get(svc.Name())
		for _, e := range svc.HTTPEndpoints {
			s = append(s, &codegen.Section{
				Template: pathTmpl,
				Data:     sdata.Endpoint(e.Name()),
			})
		}
		return s
	}
}

// input: EndpointData
const pathT = `{{ range .Routes }}// {{ .PathInit.Description }}
func {{ .PathInit.Name }}({{ range .PathInit.Args }}{{ .Name }} {{ .TypeRef }}, {{ end }}) {{ .PathInit.ReturnTypeRef }} {
{{- .PathInit.ServerCode }}
}
{{ end }}`
