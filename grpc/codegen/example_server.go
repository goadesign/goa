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

// ExampleServerFiles returns an example gRPC server implementation.
func ExampleServerFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range root.API.Servers {
		if m := exampleServer(genpkg, root, svr); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// exampleServer returns an example gRPC server implementation.
func exampleServer(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr) *codegen.File {
	var (
		mainPath string

		svrdata = example.Servers.Get(svr)
	)
	{
		mainPath = filepath.Join("cmd", svrdata.Dir, "grpc.go")
		if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
			return nil // file already exists, skip it.
		}
	}

	var (
		specs []*codegen.ImportSpec

		scope = codegen.NewNameScope()
	)
	{
		specs = []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "log"},
			{Path: "net"},
			{Path: "net/url"},
			{Path: "os"},
			{Path: "sync"},
			codegen.GoaImport("middleware"),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			codegen.GoaNamedImport("grpc/middleware", "grpcmdlwr"),
			{Path: "google.golang.org/grpc"},
			{Path: "google.golang.org/grpc/reflection"},
			{Path: "github.com/grpc-ecosystem/go-grpc-middleware", Name: "grpcmiddleware"},
		}
		for _, svc := range root.API.GRPC.Services {
			sd := GRPCServices.Get(svc.Name())
			svcName := sd.Service.PathName
			specs = append(specs, &codegen.ImportSpec{
				Path: path.Join(genpkg, "grpc", svcName, "server"),
				Name: scope.Unique(sd.Service.PkgName + "svr"),
			})
			specs = append(specs, &codegen.ImportSpec{
				Path: path.Join(genpkg, svcName),
				Name: scope.Unique(sd.Service.PkgName),
			})
			specs = append(specs, &codegen.ImportSpec{
				Path: path.Join(genpkg, "grpc", svcName, pbPkgName),
				Name: scope.Unique(svcName + pbPkgName),
			})
		}
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

	var (
		sections []*codegen.SectionTemplate
	)
	{
		var svcdata []*ServiceData
		for _, svc := range svr.Services {
			if data := GRPCServices.Get(svc); data != nil {
				svcdata = append(svcdata, data)
			}
		}
		sections = []*codegen.SectionTemplate{
			codegen.Header("", "main", specs),
			{
				Name:   "server-grpc-start",
				Source: grpcSvrStartT,
				Data: map[string]interface{}{
					"Services": svcdata,
				},
			}, {
				Name:   "server-grpc-logger",
				Source: grpcSvrLoggerT,
			}, {
				Name:   "server-grpc-init",
				Source: grpcSvrInitT,
				Data: map[string]interface{}{
					"Services": svcdata,
				},
			}, {
				Name:   "server-grpc-register",
				Source: grpcRegisterSvrT,
				Data: map[string]interface{}{
					"Services": svcdata,
				},
				FuncMap: map[string]interface{}{
					"goify":      codegen.Goify,
					"needStream": needStream,
				},
			}, {
				Name:   "server-grpc-end",
				Source: grpcSvrEndT,
				Data: map[string]interface{}{
					"Services": svcdata,
				},
			},
		}
	}
	return &codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}
}

// needStream returns true if at least one method in the defined services
// uses stream for sending payload/result.
func needStream(data []*ServiceData) bool {
	for _, svc := range data {
		for _, e := range svc.Endpoints {
			if e.ServerStream != nil || e.ClientStream != nil {
				return true
			}
		}
	}
	return false
}

const (
	// input: map[string]interface{}{"Services":[]*ServiceData}
	grpcSvrStartT = `{{ comment "handleGRPCServer starts configures and starts a gRPC server on the given URL. It shuts down the server if any error is received in the error channel." }}
func handleGRPCServer(ctx context.Context, u *url.URL{{ range $.Services }}{{ if .Service.Methods }}, {{ .Service.VarName }}Endpoints *{{ .Service.PkgName }}.Endpoints{{ end }}{{ end }}, wg *sync.WaitGroup, errc chan error, logger *log.Logger, debug bool) {
`

	grpcSvrLoggerT = `
	// Setup goa log adapter.
  var (
    adapter middleware.Logger
  )
  {
    adapter = middleware.NewLogger(logger)
  }
	`

	// input: map[string]interface{}{"Services":[]*ServiceData}
	grpcSvrInitT = `
	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
	{{- range .Services }}
		{{ .Service.VarName }}Server *{{.Service.PkgName}}svr.Server
	{{- end }}
	)
	{
	{{- range .Services }}
		{{- if .Endpoints }}
		{{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New({{ .Service.VarName }}Endpoints{{ if .HasUnaryEndpoint }}, nil{{ end }}{{ if .HasStreamingEndpoint }}, nil{{ end }})
		{{-  else }}
		{{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New(nil{{ if .HasUnaryEndpoint }}, nil{{ end }}{{ if .HasStreamingEndpoint }}, nil{{ end }})
		{{-  end }}
	{{- end }}
	}
`

	// input: map[string]interface{}{"Services":[]*ServiceData}
	grpcRegisterSvrT = `
	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			grpcmdlwr.UnaryRequestID(),
			grpcmdlwr.UnaryServerLog(adapter),
		),
	{{- if needStream .Services }}
		grpcmiddleware.WithStreamServerChain(
			grpcmdlwr.StreamRequestID(),
			grpcmdlwr.StreamServerLog(adapter),
		),
	{{- end }}
	)

	// Register the servers.
	{{- range .Services }}
	{{ .PkgName }}.Register{{ goify .Service.VarName true }}Server(srv, {{ .Service.VarName }}Server)
	{{- end }}

	for svc, info := range srv.GetServiceInfo() {
		for _, m := range info.Methods {
			logger.Printf("serving gRPC method %s", svc + "/" + m.Name)
		}
	}

	// Register the server reflection service on the server.
	// See https://grpc.github.io/grpc/core/md_doc_server-reflection.html.
	reflection.Register(srv)
`

	// input: map[string]interface{}{"Services":[]*ServiceData}
	grpcSvrEndT = `
	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		{{ comment "Start gRPC server in a separate goroutine." }}
		go func() {
			lis, err := net.Listen("tcp", u.Host)
			if err != nil {
				errc <- err
			}
			logger.Printf("gRPC server listening on %q", u.Host)
			errc <- srv.Serve(lis)
		}()

		<-ctx.Done()
		logger.Printf("shutting down gRPC server at %q", u.Host)
		srv.Stop()
  }()
}
`
)
