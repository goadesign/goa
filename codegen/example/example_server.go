package example

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// ServerFiles returns an example server main implementation for every server
// expression in the service design.
func ServerFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range root.API.Servers {
		if m := exampleSvrMain(genpkg, root, svr); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// exampleSvrMain returns the default main function for the given server
// expression.
func exampleSvrMain(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr) *codegen.File {
	svrdata := Servers.Get(svr)
	mainPath := filepath.Join("cmd", svrdata.Dir, "main.go")
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "net"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "os/signal"},
		{Path: "strings"},
		{Path: "sync"},
		{Path: "syscall"},
		{Path: "time"},
		{Path: "goa.design/clue/debug"},
		{Path: "goa.design/clue/log"},
	}

	// Iterate through services listed in the server expression.
	svcData := make([]*service.Data, len(svr.Services))
	scope := codegen.NewNameScope()
	for i, svc := range svr.Services {
		sd := service.Services.Get(svc)
		svcData[i] = sd
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, sd.PathName),
			Name: scope.Unique(sd.PkgName),
		})
	}

	var (
		rootPath string
		apiPkg   string
	)
	{
		// genpkg is created by path.Join so the separator is / regardless of operating system
		idx := strings.LastIndex(genpkg, string("/"))
		rootPath = "."
		if idx > 0 {
			rootPath = genpkg[:idx]
		}
		apiPkg = scope.Unique(strings.ToLower(codegen.Goify(root.API.Name, false)), "api")
	}
	specs = append(specs, &codegen.ImportSpec{Path: rootPath, Name: apiPkg})

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "server-main-start",
			Source: readTemplate("server_start"),
			Data: map[string]any{
				"Server": svrdata,
			},
			FuncMap: map[string]any{
				"join": strings.Join,
			},
		}, {
			Name:   "server-main-logger",
			Source: readTemplate("server_logger"),
			Data: map[string]any{
				"APIPkg": apiPkg,
			},
		}, {
			Name:   "server-main-services",
			Source: readTemplate("server_services"),
			Data: map[string]any{
				"APIPkg":   apiPkg,
				"Services": svcData,
			},
			FuncMap: map[string]any{
				"mustInitServices": mustInitServices,
			},
		}, {
			Name:   "server-main-endpoints",
			Source: readTemplate("server_endpoints"),
			Data: map[string]any{
				"Services": svcData,
			},
			FuncMap: map[string]any{
				"mustInitServices": mustInitServices,
			},
		}, {
			Name:   "server-main-interrupts",
			Source: readTemplate("server_interrupts"),
		}, {
			Name:   "server-main-handler",
			Source: readTemplate("server_handler"),
			Data: map[string]any{
				"Server":   svrdata,
				"Services": svcData,
			},
			FuncMap: map[string]any{
				"goify":   codegen.Goify,
				"join":    strings.Join,
				"toUpper": strings.ToUpper,
			},
		},
		{
			Name:   "server-main-end",
			Source: readTemplate("server_end"),
		},
	}

	return &codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}
}

// mustInitServices returns true if at least one of the services defines methods.
// It is used by the template to initialize service variables.
func mustInitServices(data []*service.Data) bool {
	for _, svc := range data {
		if len(svc.Methods) > 0 {
			return true
		}
	}
	return false
}
