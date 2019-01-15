package codegen

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
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
	pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	fpath := filepath.Join("cmd", pkg, "http.go")
	idx := strings.LastIndex(genpkg, string(os.PathSeparator))
	rootPath := "."
	if idx > 0 {
		rootPath = genpkg[:idx]
	}
	apiPkg := strings.ToLower(codegen.Goify(root.API.Name, false))
	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "log"},
		{Path: "net/http"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "sync"},
		{Path: "time"},
		{Path: "goa.design/goa/http", Name: "goahttp"},
		{Path: "goa.design/goa/http/middleware"},
		{Path: "github.com/gorilla/websocket"},
		{Path: rootPath, Name: apiPkg},
	}

	for _, svc := range root.API.HTTP.Services {
		pkgName := HTTPServices.Get(svc.Name()).Service.PkgName
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, "http", codegen.SnakeCase(svc.Name()), "server"),
			Name: pkgName + "svr",
		})
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, codegen.SnakeCase(svc.Name())),
			Name: pkgName,
		})
	}

	svcdata := make([]*ServiceData, len(svr.Services))
	for i, svc := range svr.Services {
		svcdata[i] = HTTPServices.Get(svc)
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		&codegen.SectionTemplate{
			Name:   "server-http-start",
			Source: httpSvrStartT,
			Data: map[string]interface{}{
				"Services": svcdata,
			},
		},
		&codegen.SectionTemplate{Name: "server-http-logger", Source: httpSvrLoggerT},
		&codegen.SectionTemplate{Name: "server-http-encoding", Source: httpSvrEncodingT},
		&codegen.SectionTemplate{Name: "server-http-mux", Source: httpSvrMuxT},
		&codegen.SectionTemplate{
			Name:   "server-http-init",
			Source: httpSvrInitT,
			Data: map[string]interface{}{
				"Services": svcdata,
				"APIPkg":   apiPkg,
			},
			FuncMap: map[string]interface{}{"needStream": needStream},
		},
		&codegen.SectionTemplate{Name: "server-http-middleware", Source: httpSvrMiddlewareT},
		&codegen.SectionTemplate{
			Name:   "server-http-end",
			Source: httpSvrEndT,
			Data: map[string]interface{}{
				"Services": svcdata,
			},
		},
		&codegen.SectionTemplate{Name: "server-http-errorhandler", Source: httpSvrErrorHandlerT},
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

		apiPkg = strings.ToLower(codegen.Goify(root.API.Name, false))
	)
	{
		specs := []*codegen.ImportSpec{
			{Path: "mime/multipart"},
		}
		data := HTTPServices.Get(svc.Name())
		pkgName := data.Service.PkgName
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, codegen.SnakeCase(svc.Name())),
			Name: pkgName,
		})
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

// needStream returns true if at least one method in the defined services
// uses stream for sending payload/result.
func needStream(data []*ServiceData) bool {
	for _, svc := range data {
		if streamingEndpointExists(svc) {
			return true
		}
	}
	return false
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
	// Setup logger and goa log adapter. Replace logger with your own using
	// your log package of choice.
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
	// see goa.design/encoding.
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
	{{- range .Services }}
		{{-  if .Endpoints }}
		{{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New({{ .Service.VarName }}Endpoints, mux, dec, enc, eh{{ if needStream $.Services }}, upgrader, nil{{ end }}{{ range .Endpoints }}{{ if .MultipartRequestDecoder }}, {{ $.APIPkg }}.{{ .MultipartRequestDecoder.FuncName }}{{ end }}{{ end }})
		{{-  else }}
		{{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New(nil, mux, dec, enc, eh)
		{{-  end }}
	{{- end }}
	}
	// Configure the mux.
	{{- range .Services }}
		{{ .Service.PkgName }}svr.Mount(mux{{ if .Endpoints }}, {{ .Service.VarName }}Server{{ end }})
	{{- end }}
`

	httpSvrMiddlewareT = `
	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		if debug {
			handler = middleware.Debug(mux, os.Stdout)(handler)
		}
		handler = middleware.Log(adapter)(handler)
		handler = middleware.RequestID()(handler)
	}
`

	// input: map[string]interface{}{"Services":[]*ServiceData}
	httpSvrEndT = `
	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: u.Host, Handler: handler}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		{{ comment "Start HTTP server in a separate goroutine." }}
		go func() {
		{{- range .Services }}
			for _, m := range {{ .Service.VarName }}Server.Mounts {
				{{- if .FileServers }}
				logger.Printf("file %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
				{{- else }}
				logger.Printf("method %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
				{{- end }}
			}
		{{- end }}

			logger.Printf("HTTP server listening on %q", u.Host)
			errc <- srv.ListenAndServe()
		}()

		select {
		case <-ctx.Done():
			logger.Printf("shutting down HTTP server at %q", u.Host)

			{{ comment "Shutdown gracefully with a 30s timeout." }}
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			srv.Shutdown(ctx)
			return
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
		w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Printf("[%s] ERROR: %s", id, err.Error())
	}
}
`
)
