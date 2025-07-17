package codegen

import (
	"path"
	"path/filepath"
	"slices"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// ExampleServerFiles returns example JSON-RPC server implementation.
func ExampleServerFiles(genpkg string, data *httpcodegen.ServicesData, files []*codegen.File) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range data.Root.API.Servers {
		if m := exampleServer(genpkg, data, svr, files); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

func exampleServer(genpkg string, data *httpcodegen.ServicesData, svr *expr.ServerExpr, files []*codegen.File) *codegen.File {
	svrdata := example.Servers.Get(svr, data.Root)
	httppath := filepath.Join("cmd", svrdata.Dir, "http.go")

	// Retrieve existing HTTP server file or create a new one
	var file *codegen.File
	var hasHTTP bool
	for _, f := range files {
		if f.Path == httppath {
			file = f
			hasHTTP = true
			break
		}
	}
	if file == nil {
		file = httpcodegen.ExampleServer(genpkg, data.Root, svr, data)
		updateHeader(file)
	}

	// Add JSON-RPC imports to the HTTP server file
	header := file.SectionTemplates[0]
	scope := codegen.NewNameScope()
	for _, svc := range data.Root.API.JSONRPC.Services {
		sd := data.Get(svc.Name())
		svcName := sd.Service.PathName
		codegen.AddImport(header, &codegen.ImportSpec{
			Path: path.Join(genpkg, svcName),
			Name: scope.Unique(sd.Service.PkgName),
		})
		codegen.AddImport(header, &codegen.ImportSpec{
			Path: path.Join(genpkg, "jsonrpc", svcName, "server"),
			Name: scope.Unique(sd.Service.PkgName + "jssvr"),
		})
	}

	// Add JSON-RPC to the HTTP server file
	var svcdata []*httpcodegen.ServiceData
	for _, svc := range svr.Services {
		if d := data.Get(svc); d != nil {
			svcdata = append(svcdata, d)
		}
	}
	var sections []*codegen.SectionTemplate
	for _, s := range file.SectionTemplates {
		switch s.Name {
		case "server-http-start":
			// Add JSON-RPC services to the HTTP server data so the
			// generated handleHTTPServer signature includes all the
			// necessary endpoints.
			data := s.Data.(map[string]any)
			httpServices := data["Services"].([]*httpcodegen.ServiceData)
			httpServices = slices.DeleteFunc(httpServices, func(svc *httpcodegen.ServiceData) bool {
				return len(svc.Service.Methods) == 0
			})
			for _, svc := range svcdata {
				if !slices.ContainsFunc(httpServices, func(httpsvc *httpcodegen.ServiceData) bool {
					return httpsvc.Service.Name == svc.Service.Name
				}) {
					httpServices = append(httpServices, svc)
				}
			}
			data["Services"] = httpServices
		case "server-http-end":
			updateData(s, svcdata, hasHTTP)
			mountCode := logJSONRPCMount
			if hasHTTP {
				mountCode = logHTTPMount + "\n" + logJSONRPCMount
			}
			s.Source = strings.Replace(s.Source, logHTTPMount, mountCode, 1)
		case "server-http-init":
			updateData(s, svcdata, hasHTTP)
			s.Source = jsonrpcTemplates.Read(serverConfigureT)
		}
		sections = append(sections, s)
	}
	file.SectionTemplates = sections
	return file
}

func updateData(s *codegen.SectionTemplate, svcdata []*httpcodegen.ServiceData, hasHTTP bool) {
	s.Data.(map[string]any)["JSONRPCServices"] = svcdata
	if !hasHTTP {
		delete(s.Data.(map[string]any), "Services")
	}
}

const logHTTPMount = `{{- range .Services }}
		for _, m := range {{ .Service.VarName }}Server.Mounts {
			log.Printf(ctx, "HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
		}
	{{- end }}`

const logJSONRPCMount = `{{- range .JSONRPCServices }}
		for _, m := range {{ .Service.VarName }}JSONRPCServer.Methods {
		{{- range (index .Endpoints 0).Routes }}
			log.Printf(ctx, "JSON-RPC method %q mounted on {{ .Verb }} {{ .Path }}", m)
		{{- end }}
		}
	{{- end }}`
