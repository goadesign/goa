// This file renders runnable HTTP servers and multipart helpers. Each example
// file imports relocated types and generated service packages with the
// qualifiers selected during planning.
package codegen

import (
	"maps"
	"os"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

// exampleServerFiles builds each runnable HTTP server read by Plan.Link.
func exampleServerFiles(data *ServicesData) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range data.Root.API.Servers {
		if m := exampleServer(data.Root, svr, data); m != nil {
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
func exampleServer(root *expr.RootExpr, svr *expr.ServerExpr, services *ServicesData) *codegen.File {
	genpkg := services.GenPkg()
	svrdata := example.Servers.Get(svr, root)
	fpath := filepath.Join("cmd", svrdata.Dir, "http.go")
	specs := make([]*codegen.ImportSpec, 0, 12+2*len(root.API.HTTP.Services))
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

	for _, svc := range root.API.HTTP.Services {
		sd := services.Get(svc.Name())
		svcName := sd.Service.PathName
		serverImport := services.PackageImport(path.Join(genpkg, "http", svcName, "server"))
		serviceImport := services.ServiceImport(svc.Name())
		specs = append(specs, serverImport, serviceImport)
	}

	rootPath := path.Dir(genpkg)
	apiImport := services.PackageImport(rootPath)
	apiPkg := apiImport.Name
	specs = append(specs, apiImport)

	var svcdata []*ServiceData
	for _, svc := range svr.Services {
		if data := services.Get(svc); data != nil {
			svcdata = append(svcdata, data)
		}
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "server-http-start",
			Source: httpTemplates.Read(serverStartT),
			Data: map[string]any{
				"Services": svcdata,
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
func combinedExampleServerFiles(jsonrpc, application *ServicesData) []*codegen.File {
	root := jsonrpc.Root
	files := make([]*codegen.File, 0, len(root.API.Servers))
	for _, server := range root.API.Servers {
		file := combinedExampleServer(root, server, jsonrpc, application)
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
func combinedExampleServer(root *expr.RootExpr, server *expr.ServerExpr, jsonrpc, application *ServicesData) *codegen.File {
	serverData := example.Servers.Get(server, root)
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
			ordinaryServices = append(ordinaryServices, data)
			imports = append(imports,
				application.PackageImport(path.Join(application.GenPkg(), "http", data.Service.PathName, "server")),
				application.ServiceImport(name),
			)
		}
	}
	var jsonrpcServices []*ServiceData
	for _, name := range server.Services {
		data := jsonrpc.Get(name)
		if data == nil {
			continue
		}
		jsonrpcServices = append(jsonrpcServices, data)
		imports = append(imports,
			jsonrpc.PackageImport(path.Join(jsonrpc.GenPkg(), "jsonrpc", data.Service.PathName, "server")),
			jsonrpc.ServiceImport(name),
		)
	}
	if len(ordinaryServices) == 0 && len(jsonrpcServices) == 0 {
		return nil
	}
	apiImport := jsonrpc.PackageImport(path.Dir(jsonrpc.GenPkg()))
	imports = append(imports, apiImport)
	imports = uniqueExampleImports(imports)
	data := map[string]any{
		"Services":        ordinaryServices,
		"JSONRPCServices": jsonrpcServices,
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
		Path:             filepath.Join("cmd", serverData.Dir, "http.go"),
		SectionTemplates: sections,
		SkipExist:        true,
	}
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
	if _, err := os.Stat(mpath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	var (
		sections []*codegen.SectionTemplate
		mustGen  bool
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
		specs = append(specs, services.ServiceImport(svc.Name()))
		rootPath := path.Dir(genpkg)
		specs = append(specs, services.AttributeImports(rootPath, serviceReferenceAttributes(multipartEndpoints...)...)...)

		apiPkg := services.PackageImport(rootPath).Name
		sections = []*codegen.SectionTemplate{codegen.Header("", apiPkg, specs)}
		for _, e := range data.Endpoints {
			if e.MultipartRequestDecoder != nil {
				mustGen = true
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "dummy-multipart-request-decoder",
					Source: httpTemplates.Read(dummyMultipartRequestDecoderT),
					Data:   e.MultipartRequestDecoder,
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
