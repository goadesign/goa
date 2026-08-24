// This file writes the shared example server from copied server data and the
// package names already chosen for this generation.
package example

import (
	"path"
	"sort"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
)

type (
	// serverMainData contains every import, declaration, and local name used by
	// one generated server main.
	serverMainData struct {
		// Server contains the listener settings and handler arguments written to the main.
		Server *serverMainServerData
		// Services lists the generated declarations and local names for each service.
		Services []*serverMainServiceData
		// APIPkg is the package name used for starter service constructors.
		APIPkg string
		// InterPkg is the package name used for starter interceptor constructors.
		InterPkg string
		// HasServices reports whether the main initializes any service endpoints.
		HasServices bool
		// HasInterceptors reports whether the main initializes server interceptors.
		HasInterceptors bool
	}

	// serverMainServerData keeps the server settings and handler arguments used
	// by one generated main.
	serverMainServerData struct {
		*Data
		// Variables lists every URL variable with its server flag names.
		Variables []*mainVariableData
		// Hosts contains each host and the arguments passed to its handlers.
		Hosts []*serverMainHostData
	}

	// serverMainHostData keeps one host and the arguments passed to each of its
	// generated handlers.
	serverMainHostData struct {
		*HostData
		// Variables lists the URL variables used by this host.
		Variables []*mainVariableData
		// URIs lists the host URLs and the arguments passed to their handlers.
		URIs []*URIData
	}

	// serverMainServiceData contains exact generated declarations and the local
	// names used to connect one service to its handlers.
	serverMainServiceData struct {
		// Name is the design service name used to connect handler arguments.
		Name string
		// PkgName is the package name used for the generated service package.
		PkgName string
		// ServiceVar is the local variable holding the starter service.
		ServiceVar string
		// EndpointsVar is the local variable holding the service endpoints.
		EndpointsVar string
		// InterceptorsVar is the local variable holding server interceptors.
		InterceptorsVar string
		// HasMethods reports whether the service needs a service and endpoint value.
		HasMethods bool
		// HasServerInterceptors reports whether NewEndpoints takes interceptors.
		HasServerInterceptors bool
		// ServiceDeclaration is the exact generated service interface.
		ServiceDeclaration *codegen.NameDeclaration
		// EndpointsDeclaration is the exact generated endpoint collection type.
		EndpointsDeclaration *codegen.NameDeclaration
		// NewEndpointsDeclaration is the exact generated endpoint constructor.
		NewEndpointsDeclaration *codegen.NameDeclaration
		// ServerInterceptorsDeclaration is the exact generated interceptor interface.
		ServerInterceptorsDeclaration *codegen.NameDeclaration
		// ExampleConstructorDeclaration is the exact starter service constructor.
		ExampleConstructorDeclaration *codegen.NameDeclaration
		// ExampleInterceptorsConstructor is the exact starter server interceptor
		// constructor.
		ExampleInterceptorsConstructor *codegen.NameDeclaration
	}
)

// ServerFiles returns one example main program for each copied server.
func ServerFiles(root *Root, services *service.ServicesData) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range root.Servers {
		if m := exampleSvrMain(svr, services); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// RootPath returns the project import path that contains genpkg.
func RootPath(genpkg string) string {
	return path.Dir(genpkg)
}

// exampleSvrMain writes the main program for server.
func exampleSvrMain(server *Data, services *service.ServicesData) *codegen.File {
	mainPath := server.serverMainPath
	outputPackage := server.serverPackage.ImportPath()
	specs := packageImports(server.serverPackage, serverMainFixedImports())

	// Load the generated information for each service hosted by this server.
	svcData := make([]*service.Data, len(server.Services))
	hasInterceptors := false
	serviceImports := make(map[string]struct{}, len(server.Services))
	servicePackages := make(map[string]string, len(server.Services))
	for i, svc := range server.Services {
		sd := services.Get(svc)
		svcData[i] = sd
		serviceImport := services.ServiceImport(outputPackage, svc)
		servicePackages[svc] = serviceImport.Name
		if _, exists := serviceImports[serviceImport.Path]; !exists {
			specs = append(specs, serviceImport)
			serviceImports[serviceImport.Path] = struct{}{}
		}
		hasInterceptors = hasInterceptors || len(sd.ServerInterceptors) > 0
	}
	rootPath := path.Dir(services.GenPkg())
	apiImport := services.PackageImport(outputPackage, rootPath)
	apiPkg := apiImport.Name
	specs = append(specs, apiImport)
	var interPkg string
	if hasInterceptors {
		interceptorImport := services.PackageImport(outputPackage, rootPath+"/interceptors")
		interPkg = interceptorImport.Name
		specs = append(specs, interceptorImport)
	}
	main := planServerMain(server, svcData, servicePackages, apiPkg, interPkg)

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "server-main-start",
			Source: exampleTemplates.Read(serverStartT),
			Data:   main,
			FuncMap: map[string]any{
				"join": strings.Join,
			},
		}, {
			Name:   "server-main-logger",
			Source: exampleTemplates.Read(serverLoggerT),
			Data:   main,
		}, {
			Name:   "server-main-services",
			Source: exampleTemplates.Read(serverServicesT),
			Data:   main,
		}, {
			Name:   "server-main-interceptors",
			Source: exampleTemplates.Read(serverInterceptorsT),
			Data:   main,
		}, {
			Name:   "server-main-endpoints",
			Source: exampleTemplates.Read(serverEndpointsT),
			Data:   main,
		}, {
			Name:   "server-main-interrupts",
			Source: exampleTemplates.Read(serverInterruptsT),
		}, {
			Name:   "server-main-handler",
			Source: exampleTemplates.Read(serverHandlerT),
			Data:   main,
			FuncMap: map[string]any{
				"join":    strings.Join,
				"toUpper": strings.ToUpper,
			},
		},
		{
			Name:   "server-main-end",
			Source: exampleTemplates.Read(serverEndT),
		},
	}

	return &codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}
}

// serverMainFixedImports lists packages whose names are written directly by
// the server main templates.
func serverMainFixedImports() []*codegen.ImportSpec {
	return []*codegen.ImportSpec{
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
}

// planServerMain chooses every local name once and connects each handler
// argument to the matching service variable.
func planServerMain(
	server *Data,
	services []*service.Data,
	packages map[string]string,
	apiPkg, interPkg string,
) *serverMainData {
	scope := codegen.NewNameScope()
	importNames := map[string]struct{}{apiPkg: {}}
	if interPkg != "" {
		importNames[interPkg] = struct{}{}
	}
	for _, packageName := range packages {
		importNames[packageName] = struct{}{}
	}
	orderedImports := make([]string, 0, len(importNames))
	for name := range importNames {
		orderedImports = append(orderedImports, name)
	}
	sort.Strings(orderedImports)
	for _, name := range orderedImports {
		scope.Unique(name)
	}
	for _, name := range []string{
		"addr", "c", "cancel", "context", "ctx", "debug", "err", "errc",
		"flag", "fmt", "format", "h", "log", "net", "os", "signal", "strings",
		"sync", "syscall", "time", "u", "url", "wg",
	} {
		scope.Unique(name)
	}
	byName := make(map[string]*serverMainServiceData, len(services))
	main := &serverMainData{
		APIPkg:          apiPkg,
		InterPkg:        interPkg,
		HasInterceptors: interPkg != "",
	}
	for _, serviceData := range services {
		planned := &serverMainServiceData{
			Name:                           serviceData.Name,
			PkgName:                        packages[serviceData.Name],
			HasMethods:                     len(serviceData.Methods) > 0,
			HasServerInterceptors:          len(serviceData.ServerInterceptors) > 0,
			ServiceDeclaration:             serviceData.ServiceDeclaration,
			EndpointsDeclaration:           serviceData.EndpointsDeclaration,
			NewEndpointsDeclaration:        serviceData.NewEndpointsDeclaration,
			ServerInterceptorsDeclaration:  serviceData.ServerInterceptorsDeclaration,
			ExampleConstructorDeclaration:  serviceData.ExampleConstructorDeclaration,
			ExampleInterceptorsConstructor: serviceData.ExampleServerInterceptorsConstructorDeclaration,
		}
		if planned.HasMethods {
			base := codegen.Goify(serviceData.Name, false)
			planned.ServiceVar = scope.Unique(base + "Svc")
			planned.EndpointsVar = scope.Unique(base + "Endpoints")
			if planned.HasServerInterceptors {
				planned.InterceptorsVar = scope.Unique(base + "Interceptors")
			}
			main.HasServices = true
		}
		main.Services = append(main.Services, planned)
		byName[planned.Name] = planned
	}
	main.Server = planServerMainHandlers(server, byName)
	return main
}

// planServerMainHandlers copies each host and replaces service names with the
// local variables chosen for this main function.
func planServerMainHandlers(server *Data, services map[string]*serverMainServiceData) *serverMainServerData {
	fixedFlags := make([]string, 0, 4+len(server.Transports))
	fixedFlags = append(fixedFlags, "host", "domain", "secure", "debug")
	for _, transport := range server.Transports {
		fixedFlags = append(fixedFlags, string(transport.Type)+"-port")
	}
	variables := planMainVariables(server.Variables, fixedFlags)
	planned := &serverMainServerData{
		Data:      server,
		Variables: variables.all,
		Hosts:     make([]*serverMainHostData, len(server.Hosts)),
	}
	for hostIndex, host := range server.Hosts {
		plannedHost := &serverMainHostData{
			HostData:  host,
			Variables: make([]*mainVariableData, len(host.Variables)),
			URIs:      make([]*URIData, len(host.URIs)),
		}
		for variableIndex, variable := range host.Variables {
			plannedHost.Variables[variableIndex] = variables.byName[variable.Name]
		}
		for uriIndex, uri := range host.URIs {
			plannedURI := *uri
			plannedURI.HandlerArgs = make([]HandlerArg, len(uri.HandlerArgs))
			for argIndex, arg := range uri.HandlerArgs {
				plannedArg := arg
				service := services[arg.Service]
				if arg.Endpoint {
					plannedArg.Variable = service.EndpointsVar
				} else {
					plannedArg.Variable = service.ServiceVar
				}
				plannedURI.HandlerArgs[argIndex] = plannedArg
			}
			plannedHost.URIs[uriIndex] = &plannedURI
		}
		planned.Hosts[hostIndex] = plannedHost
	}
	return planned
}
