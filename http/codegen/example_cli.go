// This file writes runnable HTTP and JSON-RPC command-line examples with the
// package names already chosen for this generation.
package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// exampleCLIFiles returns an example command-line client for the HTTP services
// on each configured server.
func exampleCLIFiles(root *example.Root, services *ServicesData) []*codegen.File {
	var files []*codegen.File
	for _, server := range root.Servers {
		if f := exampleCLI(server, services); f != nil {
			files = append(files, f)
		}
	}
	return files
}

// exampleCLI returns an example command-line client for the HTTP services on
// the given server.
func exampleCLI(server *example.Data, services *ServicesData) *codegen.File {
	genpkg := services.GenPkg()
	outputPath := filepath.Join("cmd", server.Dir+"-cli", services.dir()+".go")
	outputPackage := path.Join(path.Dir(genpkg), "cmd", server.Dir+"-cli")
	funcSuffix := "HTTP"
	if services.jsonrpc {
		funcSuffix = "JSONRPC"
	}
	rootPath := path.Dir(genpkg)
	cliImport := services.PackageImport(outputPackage, path.Join(genpkg, services.dir(), "cli", server.Dir))
	parser := services.cliParsers[server.Name]
	if parser == nil {
		panic("HTTP command parser names are missing for server " + server.Name)
	}
	var svcData []*ServiceData
	hasClientInterceptors := false
	hasMultipart := false
	for _, name := range server.Services {
		data := services.Get(name)
		if data == nil {
			continue
		}
		copy := exampleServiceDataForOutput(data, services, outputPackage)
		svcData = append(svcData, copy)
		hasClientInterceptors = hasClientInterceptors || len(data.Service.ClientInterceptors) > 0
		for _, endpoint := range data.Endpoints {
			hasMultipart = hasMultipart || endpoint.MultipartRequestDecoder != nil
		}
	}
	var interceptorsPkg string
	if hasClientInterceptors {
		interceptorImport := services.PackageImport(outputPackage, rootPath+"/interceptors")
		interceptorsPkg = interceptorImport.Name
	}
	var apiPkg string
	if hasMultipart {
		apiPkg = services.PackageImport(outputPackage, rootPath).Name
	}
	sections := []*codegen.SectionTemplate{
		plannedFileHeader("", "main", outputPath, services),
		{
			Name:   "cli-http-start",
			Source: httpTemplates.Read(cliStartT),
			Data: map[string]any{
				"Services":        svcData,
				"InterceptorsPkg": interceptorsPkg,
				"FuncSuffix":      funcSuffix,
			},
		},
		{
			Name:   "cli-http-streaming",
			Source: httpTemplates.Read(cliStreamingT),
			Data: map[string]any{
				"Services": svcData,
			},
			FuncMap: map[string]any{
				"needDialer": NeedDialer,
			},
		},
		{
			Name:   "cli-http-end",
			Source: httpTemplates.Read(cliEndT),
			Data: map[string]any{
				"Services":  svcData,
				"APIPkg":    apiPkg,
				"CLIPkg":    cliImport.Name,
				"Parser":    parser.Declarations,
				"Transport": services.label(),
			},
			FuncMap: map[string]any{
				"hasAnyInputStreams": cliHasAnyInputStreams,
				"hasInputStreams":    cliHasInputStreams,
				"hasRunnable":        cliHasRunnableCommands,
				"hasRunnableService": cliHasRunnableService,
				"needDialer":         NeedDialer,
				"hasWebSocket":       HasWebSocket,
				"kebab":              codegen.KebabCase,
				"streamsInput":       cliStreamsInput,
				"streamsOutput":      cliStreamsOutput,
			},
		},
		{
			Name:   "cli-http-usage",
			Source: httpTemplates.Read(cliUsageT),
			Data: map[string]any{
				"VarPrefix": services.dir(),
				"CLIPkg":    cliImport.Name,
				"Parser":    parser.Declarations,
			},
		},
	}
	return &codegen.File{
		Path:             outputPath,
		SectionTemplates: sections,
		SkipExist:        true,
	}
}

// cliStreamsInput reports whether an example command would need to send more
// payload values after the endpoint call starts.
func cliStreamsInput(method *service.MethodData) bool {
	return method.StreamKind == expr.ClientStreamKind || method.StreamKind == expr.BidirectionalStreamKind
}

// cliStreamsOutput reports whether an example command receives a sequence of
// results from the server.
func cliStreamsOutput(method *service.MethodData) bool {
	return method.StreamKind == expr.ServerStreamKind
}

// cliHasInputStreams reports whether a service has commands that the example
// client must reject before parsing an endpoint.
func cliHasInputStreams(data *ServiceData) bool {
	for _, endpoint := range data.Endpoints {
		if cliStreamsInput(endpoint.Method) {
			return true
		}
	}
	return false
}

// cliHasAnyInputStreams reports whether any service has a command that the
// example client must reject before parsing an endpoint.
func cliHasAnyInputStreams(services []*ServiceData) bool {
	for _, data := range services {
		if cliHasInputStreams(data) {
			return true
		}
	}
	return false
}

// cliHasRunnableCommands reports whether the example client can invoke at
// least one generated endpoint.
func cliHasRunnableCommands(services []*ServiceData) bool {
	for _, data := range services {
		if cliHasRunnableService(data) {
			return true
		}
	}
	return false
}

// cliHasRunnableService reports whether the example client can invoke at
// least one endpoint in the service.
func cliHasRunnableService(data *ServiceData) bool {
	for _, endpoint := range data.Endpoints {
		if !cliStreamsInput(endpoint.Method) {
			return true
		}
	}
	return false
}
