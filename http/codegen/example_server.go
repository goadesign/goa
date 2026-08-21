// This file renders example HTTP server wiring and multipart stubs, attaching
// relocated type imports only to the example file that references them.
package codegen

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	servicecodegen "goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// ExampleServerFiles returns an example http service implementation.
func ExampleServerFiles(genpkg string, data *ServicesData) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range data.Root.API.Servers {
		if m := ExampleServer(genpkg, data.Root, svr, data); m != nil {
			fw = append(fw, m)
		}
	}
	for _, svc := range data.Expressions.Services {
		if f := dummyMultipartFile(genpkg, data.Root, svc, data); f != nil {
			fw = append(fw, f)
		}
	}
	return fw
}

// ExampleServer returns an example HTTP server implementation.
func ExampleServer(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr, services *ServicesData) *codegen.File {
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

	scope := codegen.NewNameScope()
	for _, svc := range root.API.HTTP.Services {
		sd := services.Get(svc.Name())
		svcName := sd.Service.PathName
		specs = append(specs,
			&codegen.ImportSpec{
				Path: path.Join(genpkg, "http", svcName, "server"),
				Name: scope.Unique(sd.Service.PkgName + "svr"),
			},
			&codegen.ImportSpec{
				Path: path.Join(genpkg, svcName),
				Name: scope.Unique(sd.Service.PkgName),
			})
	}

	rootPath := example.RootPath(genpkg)
	apiPkg := scope.Unique(strings.ToLower(codegen.Goify(services.Root.API.Name, false) + "api"))
	specs = append(specs, &codegen.ImportSpec{Path: rootPath, Name: apiPkg})

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
func dummyMultipartFile(genpkg string, root *expr.RootExpr, svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	mpath := "multipart.go"
	if _, err := os.Stat(mpath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	var (
		sections []*codegen.SectionTemplate
		mustGen  bool

		scope = codegen.NewNameScope()
	)
	// determine the unique API package name different from the service names
	for _, httpSvc := range root.API.HTTP.Services {
		s := services.Get(httpSvc.Name())
		if s == nil {
			panic("unknown http service, " + httpSvc.Name()) // bug
		}
		if s.Service == nil {
			panic("unknown service, " + httpSvc.Name()) // bug
		}
		scope.Unique(s.Service.PkgName)
	}
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
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, data.Service.PathName),
			Name: scope.Unique(data.Service.PkgName, "svc"),
		})
		specs = append(specs, servicecodegen.AttributeImports(genpkg, example.RootPath(genpkg), httpEndpointAttributes(multipartEndpoints...)...)...)

		apiPkg := example.APIPkg(root, scope)
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
