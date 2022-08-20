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
		{Path: "log"},
		{Path: "net/http"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "sync"},
		{Path: "time"},
		codegen.GoaNamedImport("http", "goahttp"),
		codegen.GoaNamedImport("http/middleware", "httpmdlwr"),
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
			Source: httpSvrStartT,
			Data: map[string]interface{}{
				"Services": svcdata,
			},
		},
		{Name: "server-http-logger", Source: httpSvrLoggerT},
		{Name: "server-http-encoding", Source: httpSvrEncodingT},
		{Name: "server-http-mux", Source: httpSvrMuxT},
		{
			Name:   "server-http-init",
			Source: httpSvrInitT,
			Data: map[string]interface{}{
				"Services": svcdata,
				"APIPkg":   apiPkg,
			},
			FuncMap: map[string]interface{}{"needStream": needStream, "hasWebSocket": hasWebSocket},
		},
		{Name: "server-http-middleware", Source: httpSvrMiddlewareT},
		{
			Name:   "server-http-end",
			Source: httpSvrEndT,
			Data: map[string]interface{}{
				"Services": svcdata,
			},
		},
		{Name: "server-http-errorhandler", Source: httpSvrErrorHandlerT},
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
					Source: dummyMultipartRequestDecoderImplT,
					Data:   e.MultipartRequestDecoder,
				})
			}
			if e.MultipartRequestEncoder != nil {
				mustGen = true
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "dummy-multipart-request-encoder",
					Source: dummyMultipartRequestEncoderImplT,
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

const (
	// input: MultipartData
	dummyMultipartRequestDecoderImplT = `{{ printf "%s implements the multipart decoder for service %q endpoint %q. The decoder must populate the argument p after encoding." .FuncName .ServiceName .MethodName | comment }}
func {{ .FuncName }}(mr *multipart.Reader, p *{{ .Payload.Ref }}) error {
	// Add multipart request decoder logic here
	return nil
}
`

	// input: MultipartData
	dummyMultipartRequestEncoderImplT = `{{ printf "%s implements the multipart encoder for service %q endpoint %q." .FuncName .ServiceName .MethodName | comment }}
func {{ .FuncName }}(mw *multipart.Writer, p {{ .Payload.Ref }}) error {
	// Add multipart request encoder logic here
	return nil
}
`

	// input: map[string]interface{}{"Services":[]*ServiceData}
	httpSvrStartT = `{{ comment "handleHTTPServer starts configures and starts a HTTP server on the given URL. It shuts down the server if any error is received in the error channel." }}
func handleHTTPServer(ctx context.Context, u *url.URL{{ range $.Services }}{{ if .Service.Methods }}, {{ .Service.VarName }}Endpoints *{{ .Service.PkgName }}.Endpoints{{ end }}{{ end }}, wg *sync.WaitGroup, errc chan error, logger *log.Logger, debug bool) {
`

	httpSvrLoggerT = `
	// Setup goa log adapter.
	var (
		adapter middleware.Logger
	)
	{
		adapter = middleware.NewLogger(logger)
	}
	`

	httpSvrEncodingT = `
	// Provide the transport specific request decoder and response encoder.
	// The goa http package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions,
	// see goa.design/implement/encoding.
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)
`

	httpSvrMuxT = `
	// Build the service HTTP request multiplexer and configure it to serve
	// HTTP requests to the service endpoints.
	var mux goahttp.Muxer
	{
		mux = goahttp.NewMuxer()
	}
`

	// input: map[string]interface{}{"APIPkg":string, "Services":[]*ServiceData}
	httpSvrInitT = `
	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	var (
	{{- range .Services }}
		{{ .Service.VarName }}Server *{{.Service.PkgName}}svr.Server
	{{- end }}
	)
	{
		eh := errorHandler(logger)
	{{- if needStream .Services }}
		upgrader := &websocket.Upgrader{}
	{{- end }}
	{{- range $svc := .Services }}
		{{-  if .Endpoints }}
		{{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New({{ .Service.VarName }}Endpoints, mux, dec, enc, eh, nil{{ if hasWebSocket $svc }}, upgrader, nil{{ end }}{{ range .Endpoints }}{{ if .MultipartRequestDecoder }}, {{ $.APIPkg }}.{{ .MultipartRequestDecoder.FuncName }}{{ end }}{{ end }}{{ range .FileServers }}, nil{{ end }})
		{{-  else }}
		{{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New(nil, mux, dec, enc, eh, nil{{ range .FileServers }}, nil{{ end }})
		{{-  end }}
	{{- end }}
	{{- if .Services }}
		if debug {
			servers := goahttp.Servers{
				{{- range $svc := .Services }}
				{{ .Service.VarName }}Server,
				{{- end }}
			}
			servers.Use(httpmdlwr.Debug(mux, os.Stdout))
		}
	{{- end }}
	}
	// Configure the mux.
	{{- range .Services }}
		{{ .Service.PkgName }}svr.Mount(mux, {{ .Service.VarName }}Server)
	{{- end }}
`

	httpSvrMiddlewareT = `
	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		handler = httpmdlwr.Log(adapter)(handler)
		handler = httpmdlwr.RequestID()(handler)
	}
`

	// input: map[string]interface{}{"Services":[]*ServiceData}
	httpSvrEndT = `
	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: u.Host, Handler: handler, ReadHeaderTimeout: time.Second * 60}

	{{- range .Services }}
		for _, m := range {{ .Service.VarName }}Server.Mounts {
			logger.Printf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
		}
	{{- end }}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		{{ comment "Start HTTP server in a separate goroutine." }}
		go func() {
			logger.Printf("HTTP server listening on %q", u.Host)
			errc <- srv.ListenAndServe()
		}()

		<-ctx.Done()
		logger.Printf("shutting down HTTP server at %q", u.Host)

		{{ comment "Shutdown gracefully with a 30s timeout." }}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			logger.Printf("failed to shutdown: %v", err)
		}
	}()
}
`

	httpSvrErrorHandlerT = `
// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logger *log.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(middleware.RequestIDKey).(string)
		_, _ = w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Printf("[%s] ERROR: %s", id, err.Error())
	}
}
`
)
