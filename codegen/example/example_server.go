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
func ServerFiles(genpkg string, root *expr.RootExpr, services *service.ServicesData) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range root.API.Servers {
		if m := exampleSvrMain(genpkg, root, svr, services); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// exampleSvrMain returns the default main function for the given server
// expression.
func exampleSvrMain(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr, services *service.ServicesData) *codegen.File {
	svrdata := Servers.Get(svr, root)
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
	hasInterceptors := false
	for i, svc := range svr.Services {
		sd := services.Get(svc)
		svcData[i] = sd
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, sd.PathName),
			Name: scope.Unique(sd.PkgName, "svc"),
		})
		hasInterceptors = hasInterceptors || len(sd.ServerInterceptors) > 0
	}
	interPkg := scope.Unique("interceptors", "ex")

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
	if hasInterceptors {
		specs = append(specs, &codegen.ImportSpec{Path: path.Join(rootPath, "interceptors"), Name: interPkg})
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "server-main-start",
			Source: exampleTemplates.Read(serverStartT),
			Data: map[string]any{
				"Server": svrdata,
			},
			FuncMap: map[string]any{
				"join": strings.Join,
			},
		}, {
			Name:   "server-main-logger",
			Source: exampleTemplates.Read(serverLoggerT),
			Data: map[string]any{
				"APIPkg": apiPkg,
			},
		}, {
			Name:   "server-main-services",
			Source: exampleTemplates.Read(serverServicesT),
			Data: map[string]any{
				"APIPkg":   apiPkg,
				"Services": svcData,
			},
			FuncMap: map[string]any{
				"mustInitServices": mustInitServices,
			},
		}, {
			Name:   "server-main-interceptors",
			Source: exampleTemplates.Read(serverInterceptorsT),
			Data: map[string]any{
				"APIPkg":          apiPkg,
				"InterPkg":        interPkg,
				"Services":        svcData,
				"HasInterceptors": hasInterceptors,
			},
			FuncMap: map[string]any{
				"mustInitServices": mustInitServices,
			},
		}, {
			Name:   "server-main-endpoints",
			Source: exampleTemplates.Read(serverEndpointsT),
			Data: map[string]any{
				"Services": svcData,
			},
			FuncMap: map[string]any{
				"mustInitServices": mustInitServices,
			},
		}, {
			Name:   "server-main-interrupts",
			Source: exampleTemplates.Read(serverInterruptsT),
		}, {
			Name:   "server-main-handler",
			Source: exampleTemplates.Read(serverHandlerT),
			Data: map[string]any{
				"Server":   svrdata,
				"Services": svcData,
			},
			FuncMap: map[string]any{
				"goify":   codegen.Goify,
				"join":    strings.Join,
				"toUpper": strings.ToUpper,
				"hasJSONRPCEndpoints": func(svcData *service.Data) bool {
					return hasJSONRPCEndpoints(root, svcData)
				},
			},
		},
		{
			Name:   "server-main-end",
			Source: exampleTemplates.Read(serverEndT),
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

// hasJSONRPCEndpoints returns true if the service has JSON-RPC endpoints.
func hasJSONRPCEndpoints(root *expr.RootExpr, data *service.Data) bool {
	for _, svc := range root.API.JSONRPC.Services {
		if svc.Name() == data.Name {
			return true
		}
	}
	return false
}
