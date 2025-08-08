package codegen

import (
	"path"
	"path/filepath"
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
	sections := make([]*codegen.SectionTemplate, 0, len(file.SectionTemplates)+2)
	for _, s := range file.SectionTemplates {
		switch s.Name {
		case "server-http-start":
			// Check if the main template already has JSONRPCServices data
			data := s.Data.(map[string]any)
			if _, hasJSONRPCServices := data["JSONRPCServices"]; !hasJSONRPCServices {
				// Main template doesn't have JSON-RPC services, so we need to add them
				data["JSONRPCServices"] = svcdata
				// Replace with JSON-RPC template that includes service parameters in function signature
				s.Source = jsonrpcTemplates.Read(serverHttpStartT)
			}
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
			s.FuncMap = map[string]any{
				"needDialer":   httpcodegen.NeedDialer,
				"hasWebSocket": httpcodegen.HasWebSocket,
			}
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
