package codegen

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

// ExampleServerFiles returns an example http service implementation.
func ExampleServerFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range root.API.Servers {
		if m := exampleServer(genpkg, root, svr); m != nil {
			fw = append(fw, m)
		}
	}
	for _, svc := range root.API.HTTP.Services {
		if f := dummyMultipartFile(genpkg, root, svc); f != nil {
			fw = append(fw, f)
		}
	}
	return fw
}

// exampleServer returns an example HTTP server implementation.
func exampleServer(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr) *codegen.File {
	svrdata := example.Servers.Get(svr)
	fpath := filepath.Join("cmd", svrdata.Dir, "http.go")
	specs := []*codegen.ImportSpec{
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

	scope := codegen.NewNameScope()
	for _, svc := range root.API.HTTP.Services {
		sd := HTTPServices.Get(svc.Name())
		svcName := sd.Service.PathName
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, "http", svcName, "server"),
			Name: scope.Unique(sd.Service.PkgName + "svr"),
		})
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, svcName),
			Name: scope.Unique(sd.Service.PkgName),
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

	var svcdata []*ServiceData
	for _, svc := range svr.Services {
		if data := HTTPServices.Get(svc); data != nil {
			svcdata = append(svcdata, data)
		}
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "server-http-start",
			Source: readTemplate("server_start"),
			Data: map[string]any{
				"Services": svcdata,
			},
		},
		{
			Name:   "server-http-encoding",
			Source: readTemplate("server_encoding"),
		},
		{
			Name:   "server-http-mux",
			Source: readTemplate("server_mux"),
		},
		{
			Name:   "server-http-init",
			Source: readTemplate("server_configure"),
			Data: map[string]any{
				"Services": svcdata,
				"APIPkg":   apiPkg,
			},
			FuncMap: map[string]any{"needStream": needStream, "hasWebSocket": hasWebSocket},
		},
		{
			Name:   "server-http-middleware",
			Source: readTemplate("server_middleware"),
		},
		{
			Name:   "server-http-end",
			Source: readTemplate("server_end"),
			Data: map[string]any{
				"Services": svcdata,
			},
		},
		{
			Name:   "server-http-errorhandler",
			Source: readTemplate("server_error_handler"),
		},
	}

	return &codegen.File{Path: fpath, SectionTemplates: sections, SkipExist: true}
}

// dummyMultipartFile returns a dummy implementation of the multipart decoders
// and encoders.
func dummyMultipartFile(genpkg string, root *expr.RootExpr, svc *expr.HTTPServiceExpr) *codegen.File {
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
	for _, svc := range root.Services {
		s := HTTPServices.Get(svc.Name)
		if s == nil {
			panic("unknown http service, " + svc.Name) // bug
		}
		if s.Service == nil {
			panic("unknown service, " + svc.Name) // bug
		}
		scope.Unique(s.Service.PkgName)
	}
	{
		specs := []*codegen.ImportSpec{
			{Path: "mime/multipart"},
		}
		data := HTTPServices.Get(svc.Name())
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, data.Service.PathName),
			Name: scope.Unique(data.Service.PkgName, "svc"),
		})

		apiPkg := scope.Unique(strings.ToLower(codegen.Goify(root.API.Name, false)), "api")
		sections = []*codegen.SectionTemplate{codegen.Header("", apiPkg, specs)}
		for _, e := range data.Endpoints {
			if e.MultipartRequestDecoder != nil {
				mustGen = true
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "dummy-multipart-request-decoder",
					Source: readTemplate("dummy_multipart_request_decoder"),
					Data:   e.MultipartRequestDecoder,
				})
			}
			if e.MultipartRequestEncoder != nil {
				mustGen = true
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "dummy-multipart-request-encoder",
					Source: readTemplate("dummy_multipart_request_encoder"),
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
