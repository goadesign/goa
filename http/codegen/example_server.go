// This file renders runnable HTTP servers and multipart helpers. Each example
// file imports relocated types and generated service packages with the
// qualifiers selected during planning.
package codegen

import (
	"os"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

// ExampleServerFiles returns an example http service implementation.
func ExampleServerFiles(data *ServicesData) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range data.Root.API.Servers {
		if m := ExampleServer(data.Root, svr, data); m != nil {
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

// ExampleServer returns an example HTTP server implementation.
func ExampleServer(root *expr.RootExpr, svr *expr.ServerExpr, services *ServicesData) *codegen.File {
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
		specs = append(specs, services.AttributeImports(rootPath, ServiceReferenceAttributes(multipartEndpoints...)...)...)

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
