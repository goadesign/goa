package codegen

import (
	"path"
	"path/filepath"

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
	for _, s := range file.SectionTemplates {
		switch s.Name {
		case "server-http-start":
			// Only set the JSON-RPC services if not already populated.
			data := s.Data.(map[string]any)
			if existing, _ := data["JSONRPCServices"].([]*httpcodegen.ServiceData); len(existing) == 0 {
				data["JSONRPCServices"] = svcdata
			}
		case "server-http-init", "server-http-end":
			updateData(s, svcdata, hasHTTP)
		}
	}
	return file
}

func updateData(s *codegen.SectionTemplate, svcdata []*httpcodegen.ServiceData, hasHTTP bool) {
	s.Data.(map[string]any)["JSONRPCServices"] = svcdata
	if !hasHTTP {
		delete(s.Data.(map[string]any), "Services")
	}
}
