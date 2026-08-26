// This file writes runnable gRPC command-line examples with the package names
// already chosen for this generation.
package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// exampleCLIFiles returns an example gRPC client tool implementation.
func exampleCLIFiles(root *example.Root, services *ServicesData) []*codegen.File {
	var files []*codegen.File
	for _, server := range root.Servers {
		if f := exampleCLI(services, server); f != nil {
			files = append(files, f)
		}
	}
	return files
}

// exampleCLI writes the gRPC command-line program for server.
func exampleCLI(services *ServicesData, server *example.Data) *codegen.File {
	genpkg := services.GenPkg()
	mainPath := filepath.Join("cmd", server.Dir+"-cli", "grpc.go")
	rootPath := path.Dir(genpkg)
	outputPackage := path.Join(rootPath, "cmd", server.Dir+"-cli")
	cliImport := services.PackageImport(outputPackage, path.Join(genpkg, "grpc", "cli", server.Dir))
	parser := services.cliPlan.parser(server.Name)
	if parser == nil {
		panic("gRPC command parser names are missing for server " + server.Name)
	}

	var svcData []*ServiceData
	hasClientInterceptors := false
	for _, svc := range server.Services {
		if data := services.Get(svc); data != nil {
			svcData = append(svcData, services.exampleServiceData(data, outputPackage, false))
			hasClientInterceptors = hasClientInterceptors || len(data.Service.ClientInterceptors) > 0
		}
	}
	var interceptorsPkg string
	if hasClientInterceptors {
		interceptorImport := services.PackageImport(outputPackage, rootPath+"/interceptors")
		interceptorsPkg = interceptorImport.Name
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", nil),
		{
			Name:   "do-grpc-cli",
			Source: grpcTemplates.Read(grpcDoGRPCCLIT),
			Data: map[string]any{
				"DefaultTransport": server.DefaultTransport(),
				"Services":         svcData,
				"InterceptorsPkg":  interceptorsPkg,
				"CLIPkg":           cliImport.Name,
				"Parser":           parser.Declarations,
			},
			FuncMap: map[string]any{
				"hasAnyInputStreams": cliHasAnyInputStreams,
				"hasInputStreams":    cliHasInputStreams,
				"hasRunnable":        cliHasRunnableCommands,
				"hasRunnableService": cliHasRunnableService,
				"kebab":              codegen.KebabCase,
				"streamsInput":       cliStreamsInput,
				"streamsOutput":      cliStreamsOutput,
			},
		},
	}

	return addEndpointImports(&codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}, services)
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

// parser returns the command parser saved for the named server.
func (p *grpcCLIPlan) parser(serverName string) *cli.ParserPlan {
	for _, server := range p.servers {
		if server.name == serverName {
			return server.parser
		}
	}
	return nil
}
