// This file writes runnable HTTP servers and file-upload helpers with the
// package names already chosen for this generation.
package codegen

import (
	"maps"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

type (
	// exampleServerArgumentData contains one typed parameter accepted by a
	// generated transport helper.
	exampleServerArgumentData struct {
		// Name is the parameter name used inside the helper.
		Name string
		// PkgName is the generated service package name.
		PkgName string
		// TypeName is the generated service or endpoint type name.
		TypeName string
		// Pointer is true when TypeName is an endpoint collection pointer.
		Pointer bool
	}

	// exampleMultipartDecoderData describes the HTTP request body filled by one
	// starter multipart decoder.
	exampleMultipartDecoderData struct {
		*MultipartData
		// BodyType is the request body type as seen from the starter service
		// package.
		BodyType string
	}
)

// exampleServerFiles builds each runnable HTTP server from copied server data.
func exampleServerFiles(root *example.Root, data *ServicesData) []*codegen.File {
	var fw []*codegen.File
	for _, server := range root.Servers {
		if m := exampleServer(server, data); m != nil {
			fw = append(fw, m)
		}
	}
	for _, svc := range data.Expressions.Services {
		if f := dummyMultipartFile(svc, data); f != nil {
			fw = append(fw, f)
		}
	}
	return fw
}

// exampleServer returns an example HTTP server implementation.
func exampleServer(server *example.Data, services *ServicesData) *codegen.File {
	genpkg := services.GenPkg()
	fpath := filepath.Join("cmd", server.Dir, "http.go")
	outputPackage := path.Join(path.Dir(genpkg), "cmd", server.Dir)
	specs := make([]*codegen.ImportSpec, 0, 12+2*len(services.Expressions.Services))
	baseSpecs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "net/http"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "sync"},
		{Path: "time"},
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: "goa.design/clue/debug"},
		{Path: "goa.design/clue/log"},
		codegen.GoaImport("middleware"),
		{Path: "github.com/gorilla/websocket"},
	}
	specs = append(specs, baseSpecs...)

	for _, serviceName := range server.Services {
		sd := services.Get(serviceName)
		if sd == nil {
			continue
		}
		svcName := sd.Service.PathName
		serverImport := services.PackageImport(outputPackage, path.Join(genpkg, "http", svcName, "server"))
		serviceImport := services.ServiceImport(outputPackage, serviceName)
		specs = append(specs, serverImport, serviceImport)
	}

	rootPath := path.Dir(genpkg)
	apiImport := services.PackageImport(outputPackage, rootPath)
	apiPkg := apiImport.Name
	specs = append(specs, apiImport)

	var svcdata []*ServiceData
	for _, svc := range server.Services {
		if data := services.Get(svc); data != nil {
			copy := exampleServiceDataForOutput(data, services, outputPackage)
			copy.ServerPkgName = services.PackageImport(
				outputPackage,
				path.Join(genpkg, "http", data.Service.PathName, "server"),
			).Name
			svcdata = append(svcdata, copy)
		}
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "server-http-start",
			Source: httpTemplates.Read(serverStartT),
			Data: map[string]any{
				"Services":    svcdata,
				"HandlerArgs": exampleServerArguments(server, svcdata, nil),
				// JSONRPCServices must always be set (typed nil when
				// absent) so the template functions receive a valid
				// []*ServiceData value. The JSON-RPC generator
				// overrides it with the JSON-RPC service data.
				"JSONRPCServices": []*ServiceData(nil),
			},
		},
		{
			Name:   "server-http-encoding",
			Source: httpTemplates.Read(serverEncodingT),
		},
		{
			Name:   "server-http-mux",
			Source: httpTemplates.Read(serverMuxT),
		},
		{
			Name:   "server-http-init",
			Source: httpTemplates.Read(serverConfigureT),
			Data: map[string]any{
				"Services":        svcdata,
				"JSONRPCServices": []*ServiceData(nil),
				"APIPkg":          apiPkg,
			},
			FuncMap: map[string]any{"needDialer": NeedDialer, "hasWebSocket": HasWebSocket},
		},
		{
			Name:   "server-http-middleware",
			Source: httpTemplates.Read(serverMiddlewareT),
		},
		{
			Name:   "server-http-end",
			Source: httpTemplates.Read(serverEndT),
			Data: map[string]any{
				"Services":        svcdata,
				"JSONRPCServices": []*ServiceData(nil),
			},
		},
		{
			Name:   "server-http-errorhandler",
			Source: httpTemplates.Read(serverErrorHandlerT),
		},
	}

	return &codegen.File{Path: fpath, SectionTemplates: sections, SkipExist: true}
}

// combinedExampleServerFiles builds runnable server files that mount both the
// JSON-RPC services and the ordinary HTTP services from one design. The caller
// may edit every returned file without changing either input plan.
func combinedExampleServerFiles(root *example.Root, jsonrpc, application *ServicesData) []*codegen.File {
	files := make([]*codegen.File, 0, len(root.Servers))
	for _, server := range root.Servers {
		file := combinedExampleServer(server, jsonrpc, application)
		if file != nil {
			files = append(files, file)
		}
	}
	if application != nil {
		for _, service := range application.Expressions.Services {
			if file := dummyMultipartFile(service, application); file != nil {
				files = append(files, cloneGeneratedFile(file))
			}
		}
	}
	return files
}

// combinedExampleServer builds one main-package file for a configured server.
// It reads service membership from server and writes separate HTTP and
// JSON-RPC lists because the code that writes main initializes them differently.
func combinedExampleServer(server *example.Data, jsonrpc, application *ServicesData) *codegen.File {
	outputPackage := path.Join(path.Dir(jsonrpc.GenPkg()), "cmd", server.Dir)
	imports := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "net/http"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "sync"},
		{Path: "time"},
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: "goa.design/clue/debug"},
		{Path: "goa.design/clue/log"},
		codegen.GoaImport("middleware"),
		{Path: "github.com/gorilla/websocket"},
	}
	var ordinaryServices []*ServiceData
	if application != nil {
		for _, name := range server.Services {
			data := application.Get(name)
			if data == nil {
				continue
			}
			copy := exampleServiceDataForOutput(data, application, outputPackage)
			copy.ServerPkgName = application.PackageImport(
				outputPackage,
				path.Join(application.GenPkg(), "http", data.Service.PathName, "server"),
			).Name
			ordinaryServices = append(ordinaryServices, copy)
			imports = append(imports,
				application.PackageImport(outputPackage, path.Join(application.GenPkg(), "http", data.Service.PathName, "server")),
				application.ServiceImport(outputPackage, name),
			)
		}
	}
	var jsonrpcServices []*ServiceData
	for _, name := range server.Services {
		data := jsonrpc.Get(name)
		if data == nil {
			continue
		}
		copy := exampleServiceDataForOutput(data, jsonrpc, outputPackage)
		copy.ServerPkgName = jsonrpc.PackageImport(
			outputPackage,
			path.Join(jsonrpc.GenPkg(), "jsonrpc", data.Service.PathName, "server"),
		).Name
		jsonrpcServices = append(jsonrpcServices, copy)
		imports = append(imports,
			jsonrpc.PackageImport(outputPackage, path.Join(jsonrpc.GenPkg(), "jsonrpc", data.Service.PathName, "server")),
			jsonrpc.ServiceImport(outputPackage, name),
		)
	}
	if len(ordinaryServices) == 0 && len(jsonrpcServices) == 0 {
		return nil
	}
	apiImport := jsonrpc.PackageImport(outputPackage, path.Dir(jsonrpc.GenPkg()))
	imports = append(imports, apiImport)
	imports = uniqueExampleImports(imports)
	data := map[string]any{
		"Services":        ordinaryServices,
		"JSONRPCServices": jsonrpcServices,
		"HandlerArgs":     exampleServerArguments(server, ordinaryServices, jsonrpcServices),
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", imports),
		{Name: "server-http-start", Source: httpTemplates.Read(serverStartT), Data: data},
		{Name: "server-http-encoding", Source: httpTemplates.Read(serverEncodingT)},
		{Name: "server-http-mux", Source: httpTemplates.Read(serverMuxT)},
		{
			Name:   "server-http-init",
			Source: httpTemplates.Read(serverConfigureT),
			Data: map[string]any{
				"Services":        ordinaryServices,
				"JSONRPCServices": jsonrpcServices,
				"APIPkg":          apiImport.Name,
			},
			FuncMap: map[string]any{"needDialer": NeedDialer, "hasWebSocket": HasWebSocket},
		},
		{Name: "server-http-middleware", Source: httpTemplates.Read(serverMiddlewareT)},
		{Name: "server-http-end", Source: httpTemplates.Read(serverEndT), Data: data},
		{Name: "server-http-errorhandler", Source: httpTemplates.Read(serverErrorHandlerT)},
	}
	return &codegen.File{
		Path:             filepath.Join("cmd", server.Dir, "http.go"),
		SectionTemplates: sections,
		SkipExist:        true,
	}
}

// exampleServerArguments adds generated Go names and types to the ordered
// service values copied by the shared example plan.
func exampleServerArguments(
	server *example.Data,
	ordinary, jsonrpc []*ServiceData,
) []*exampleServerArgumentData {
	services := make(map[string]*ServiceData, len(ordinary)+len(jsonrpc))
	for _, service := range ordinary {
		services[service.Service.Name] = service
	}
	for _, service := range jsonrpc {
		services[service.Service.Name] = service
	}
	planned := server.HandlerArgs(example.TransportHTTP)
	arguments := make([]*exampleServerArgumentData, len(planned))
	for index, argument := range planned {
		service := services[argument.Service].Service
		data := &exampleServerArgumentData{
			PkgName: service.PkgName,
		}
		if argument.Endpoint {
			data.Name = service.VarName + "Endpoints"
			data.TypeName = service.EndpointsDeclaration.Name()
			data.Pointer = true
		} else {
			data.Name = service.VarName + "Svc"
			data.TypeName = service.ServiceDeclaration.Name()
		}
		arguments[index] = data
	}
	return arguments
}

// uniqueExampleImports keeps the first import for each Go package path. A
// service exposed over both protocols uses the same generated service package.
func uniqueExampleImports(imports []*codegen.ImportSpec) []*codegen.ImportSpec {
	result := make([]*codegen.ImportSpec, 0, len(imports))
	seen := make(map[string]struct{}, len(imports))
	for _, spec := range imports {
		if _, ok := seen[spec.Path]; ok {
			continue
		}
		seen[spec.Path] = struct{}{}
		result = append(result, spec)
	}
	return result
}

// cloneGeneratedFile copies a generated file and its section records so the
// caller may change the copy without changing the source plan.
func cloneGeneratedFile(source *codegen.File) *codegen.File {
	if source == nil {
		return nil
	}
	clone := *source
	clone.SectionTemplates = make([]*codegen.SectionTemplate, len(source.SectionTemplates))
	for index, section := range source.SectionTemplates {
		sectionClone := *section
		sectionClone.FuncMap = maps.Clone(section.FuncMap)
		sectionClone.Data = cloneRenderData(section.Data)
		clone.SectionTemplates[index] = &sectionClone
	}
	return &clone
}

// cloneJSONRPCCodecFile copies an encoder and decoder file and replaces each
// HTTP endpoint value with the smaller value read by JSON-RPC code.
func cloneJSONRPCCodecFile(source *codegen.File) *codegen.File {
	if source == nil {
		return nil
	}
	clone := *source
	clone.SectionTemplates = make([]*codegen.SectionTemplate, len(source.SectionTemplates))
	for index, section := range source.SectionTemplates {
		sectionCopy := *section
		sectionCopy.FuncMap = maps.Clone(section.FuncMap)
		if endpoint, ok := section.Data.(*EndpointData); ok {
			switch section.Name {
			case "response-decoder":
				data := copyJSONRPCEndpoint(endpoint)
				sectionCopy.Data = &data
			case "request-builder", "request-encoder", "request-decoder":
				sectionCopy.Data = copyJSONRPCRequestCodec(endpoint)
			default:
				panic("JSON-RPC codec contains an unsupported endpoint section " + section.Name)
			}
		} else if helper, ok := section.Data.(*codegen.TransformFunctionData); ok {
			if section.Name != "client-transform-helper" && section.Name != "server-transform-helper" {
				panic("JSON-RPC codec contains a transform helper in unsupported section " + section.Name)
			}
			sectionCopy.Data = copyJSONRPCTransformFunction(helper)
		} else {
			sectionCopy.Data = cloneRenderData(section.Data)
		}
		clone.SectionTemplates[index] = &sectionCopy
	}
	return &clone
}

// dummyMultipartFile returns a dummy implementation of the multipart decoders
// and encoders.
func dummyMultipartFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	genpkg := services.GenPkg()
	mpath := "multipart.go"
	outputPackage := path.Dir(genpkg)
	var (
		sections    []*codegen.SectionTemplate
		decoderData = make(map[*MultipartData]*exampleMultipartDecoderData)
		mustGen     bool
	)
	{
		specs := make([]*codegen.ImportSpec, 0, 2)
		specs = append(specs, &codegen.ImportSpec{Path: "mime/multipart"})
		data := services.Get(svc.Name())
		var multipartEndpoints []*expr.HTTPEndpointExpr
		for _, endpoint := range data.Endpoints {
			if endpoint.MultipartRequestDecoder != nil || endpoint.MultipartRequestEncoder != nil {
				multipartEndpoints = append(multipartEndpoints, svc.Endpoint(endpoint.Method.Name))
			}
		}
		specs = append(specs, services.ServiceImport(outputPackage, svc.Name()))
		rootPath := path.Dir(genpkg)
		specs = append(specs, services.AttributeImports(rootPath, serviceReferenceAttributes(multipartEndpoints...)...)...)
		for _, endpoint := range data.Endpoints {
			if endpoint.MultipartRequestDecoder == nil {
				continue
			}
			bodyType, bodyImport := exampleMultipartBodyType(svc, endpoint, services, outputPackage)
			if bodyImport != nil {
				specs = append(specs, bodyImport)
			}
			decoderData[endpoint.MultipartRequestDecoder] = &exampleMultipartDecoderData{
				MultipartData: endpoint.MultipartRequestDecoder,
				BodyType:      bodyType,
			}
		}

		apiPkg := examplePackageImportName(services.Root)
		sections = []*codegen.SectionTemplate{codegen.Header("", apiPkg, specs)}
		for _, e := range data.Endpoints {
			if e.MultipartRequestDecoder != nil {
				mustGen = true
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "dummy-multipart-request-decoder",
					Source: httpTemplates.Read(dummyMultipartRequestDecoderT),
					Data:   decoderData[e.MultipartRequestDecoder],
				})
			}
			if e.MultipartRequestEncoder != nil {
				mustGen = true
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "dummy-multipart-request-encoder",
					Source: httpTemplates.Read(dummyMultipartRequestEncoderT),
					Data:   e.MultipartRequestEncoder,
				})
			}
		}
	}
	if !mustGen {
		return nil
	}
	return &codegen.File{
		Path:             mpath,
		SectionTemplates: sections,
		SkipExist:        true,
	}
}

// exampleMultipartBodyType returns the request body type visible from the
// starter service package and the generated server import needed to name it.
func exampleMultipartBodyType(svc *expr.HTTPServiceExpr, endpoint *EndpointData, services *ServicesData, outputPackage string) (string, *codegen.ImportSpec) {
	serverBody := endpoint.Payload.Request.ServerBody
	service := services.Get(svc.Name())
	serverPath := path.Join(services.GenPkg(), "http", service.Service.PathName, "server")
	serverImport := services.PackageImport(outputPackage, serverPath)
	if serverBody.Declaration != nil {
		return serverImport.Name + "." + serverBody.Declaration.Name(), serverImport
	}
	body := serverBody.attribute
	usesServerType := false
	collectUserTypes(body.Type, func(expr.UserType) {
		usesServerType = true
	})
	if usesServerType {
		resolver := &wireAttributeScope{
			catalog:         service.serverWireTypes,
			base:            codegen.NewAttributeScope(service.serverWireTypes.scope),
			pkg:             serverImport.Name,
			policy:          jsonBodyPolicy(true, true, true, ""),
			exactOccurrence: true,
		}
		return resolver.Name(body, serverImport.Name, true, false), serverImport
	}
	return serverBody.VarName, nil
}
