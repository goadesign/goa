package codegen

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
)

// ExampleServerFiles returns and example main and dummy service
// implementations.
func ExampleServerFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, svc := range root.API.HTTP.Services {
		f := dummyServiceFile(genpkg, root, svc)
		if f != nil {
			fw = append(fw, f)
		}
	}
	for _, svr := range root.API.Servers {
		if m := exampleMain(genpkg, root, svr); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// dummyServiceFile returns a dummy implementation of the given service.
func dummyServiceFile(genpkg string, root *expr.RootExpr, svc *expr.HTTPServiceExpr) *codegen.File {
	path := codegen.SnakeCase(svc.Name()) + ".go"
	data := HTTPServices.Get(svc.Name())
	apiPkg := strings.ToLower(codegen.Goify(root.API.Name, false))
	sections := []*codegen.SectionTemplate{
		codegen.Header("", apiPkg, []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "log"},
			{Path: "mime/multipart"},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: data.Service.PkgName},
		}),
		{
			Name:   "dummy-service",
			Source: dummyServiceStructT,
			Data:   data,
		},
	}
	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "dummy-endpoint",
			Source: dummyEndpointImplT,
			Data:   e,
		})
		if e.MultipartRequestDecoder != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "dummy-multipart-request-decoder",
				Source: dummyMultipartRequestDecoderImplT,
				Data:   e.MultipartRequestDecoder,
			})
		}
		if e.MultipartRequestEncoder != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "dummy-multipart-request-encoder",
				Source: dummyMultipartRequestEncoderImplT,
				Data:   e.MultipartRequestEncoder,
			})
		}
	}

	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
		SkipExist:        true,
	}
}

func exampleMain(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr) *codegen.File {
	pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	mainPath := filepath.Join("cmd", pkg, "main.go")
	idx := strings.LastIndex(genpkg, string(os.PathSeparator))
	rootPath := "."
	if idx > 0 {
		rootPath = genpkg[:idx]
	}
	apiPkg := strings.ToLower(codegen.Goify(root.API.Name, false))
	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "log"},
		{Path: "net/http"},
		{Path: "os"},
		{Path: "os/signal"},
		{Path: "time"},
		{Path: "goa.design/goa", Name: "goa"},
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

	sections := []*codegen.SectionTemplate{codegen.Header("", "main", specs)}
	svcdata := make([]*ServiceData, len(svr.Services))
	for i, svc := range svr.Services {
		svcdata[i] = HTTPServices.Get(svc)
	}
	if needStream(svcdata) {
		specs = append(specs, &codegen.ImportSpec{Path: "github.com/gorilla/websocket"})
	}
	// URIs have been validated by DSL.
	u, _ := url.Parse(string(root.API.Servers[0].Hosts[0].URIs[0]))
	data := map[string]interface{}{
		"Services":    svcdata,
		"APIPkg":      apiPkg,
		"DefaultHost": u.Host,
	}

	// Service Main sections
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main-start",
		Source: mainStartT,
		Data:   data,
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main-logger",
		Source: mainLoggerT,
		Data:   data,
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main-struct",
		Source: mainStructT,
		Data:   data,
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main-endpoints",
		Source: mainEndpointsT,
		Data:   data,
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main-encoding",
		Source: mainEncodingT,
		Data:   data,
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main-mux",
		Source: mainMuxT,
		Data:   data,
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:    "service-main-server-init",
		Source:  mainServerInitT,
		Data:    data,
		FuncMap: map[string]interface{}{"needStream": needStream},
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main-middleware",
		Source: mainMiddlewareT,
		Data:   data,
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main-http",
		Source: mainHTTPT,
		Data:   data,
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main-end",
		Source: mainEndT,
		Data:   data,
	})

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main-errorhandler",
		Source: mainErrorHandlerT,
		Data:   data,
	})

	return &codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}
}

// dummyMultipart returns a dummy implementation of the multipart decoders
// and encoders.
func dummyMultipart(genpkg string, root *expr.RootExpr) *codegen.File {
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
		for _, svc := range root.API.HTTP.Services {
			pkgName := HTTPServices.Get(svc.Name()).Service.PkgName
			specs = append(specs, &codegen.ImportSpec{
				Path: path.Join(genpkg, codegen.SnakeCase(svc.Name())),
				Name: pkgName,
			})
		}
		header := codegen.Header("", apiPkg, specs)
		sections = []*codegen.SectionTemplate{header}
		for _, svc := range root.API.HTTP.Services {
			data := HTTPServices.Get(svc.Name())
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

// needStream returns true if at least one method in the list of services
// uses stream for sending payload/result.
func needStream(data []*ServiceData) bool {
	for _, svc := range data {
		if streamingEndpointExists(svc) {
			return true
		}
	}
	return false
}

// input: ServiceData
const dummyServiceStructT = `{{ printf "%s service example implementation.\nThe example methods log the requests and return zero values." .Service.Name | comment }}
type {{ .Service.VarName }}Svc struct {
	logger *log.Logger
}
{{ printf "New%s returns the %s service implementation." .Service.StructName .Service.Name | comment }}
func New{{ .Service.StructName }}(logger *log.Logger) {{ .Service.PkgName }}.Service {
	return &{{ .Service.VarName }}Svc{logger}
}
`

// input: EndpointData
const dummyEndpointImplT = `{{ comment .Method.Description }}
{{- if .ServerStream }}
func (s *{{ .ServiceVarName }}Svc) {{ .Method.VarName }}(ctx context.Context{{ if .Payload.Ref }}, p {{ .Payload.Ref }}{{ end }}, stream {{ .ServerStream.Interface }}) (err error) {
{{- else }}
func (s *{{ .ServiceVarName }}Svc) {{ .Method.VarName }}(ctx context.Context{{ if .Payload.Ref }}, p {{ .Payload.Ref }}{{ end }}) ({{ if .Result.Ref }}res {{ .Result.Ref }}, {{ if .Method.ViewedResult }}{{ if not .Method.ViewedResult.ViewName }}view string, {{ end }}{{ end }} {{ end }}err error) {
{{- end }}
{{- if and (and .Result.Ref .Result.IsStruct) (not .ServerStream) }}
	res = &{{ .Result.Name }}{}
{{- end }}
{{- if .Method.ViewedResult }}
	{{- if .ServerStream }}
	stream.SetView({{ printf "%q" .Result.View }})
	{{- else if not .Method.ViewedResult.ViewName }}
	view = {{ printf "%q" .Result.View }}
	{{- end }}
{{- end }}
	s.logger.Print("{{ .ServiceVarName }}.{{ .Method.Name }}")
	return
}
`

// input: MultipartData
const dummyMultipartRequestDecoderImplT = `{{ printf "%s implements the multipart decoder for service %q endpoint %q. The decoder must populate the argument p after encoding." .FuncName .ServiceName .MethodName | comment }}
func {{ .FuncName }}(mr *multipart.Reader, p *{{ .Payload.Ref }}) error {
	// Add multipart request decoder logic here
	return nil
}
`

// input: MultipartData
const dummyMultipartRequestEncoderImplT = `{{ printf "%s implements the multipart encoder for service %q endpoint %q." .FuncName .ServiceName .MethodName | comment }}
func {{ .FuncName }}(mw *multipart.Writer, p {{ .Payload.Ref }}) error {
	// Add multipart request encoder logic here
	return nil
}
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainStartT = `func main() {
	// Define command line flags, add any other flag required to configure
	// the service.
	var (
		addr = flag.String("listen", "{{ .DefaultHost }}", "HTTP listen ` + "`" + `address` + "`" + `")
		dbg  = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainLoggerT = `
	// Setup logger and goa log adapter. Replace logger with your own using
	// your log package of choice.
	var (
		adapter middleware.Logger
		logger *log.Logger
	)
	{
		logger = log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)
		adapter = middleware.NewLogger(logger)
	}
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainStructT = `
	// Create the structs that implement the services.
	var (
	{{- range .Services }}
		{{-  if .Endpoints }}
		{{ .Service.VarName }}Svc {{.Service.PkgName}}.Service
		{{-  end }}
	{{- end }}
	)
	{
	{{- range .Services }}
		{{-  if .Endpoints }}
		{{ .Service.VarName }}Svc = {{ $.APIPkg }}.New{{ .Service.StructName }}(logger)
		{{-  end }}
	{{- end }}
	}
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainEndpointsT = `
	// Wrap the services in endpoints that can be invoked from other
	// services potentially running in different processes.
	var (
	{{- range .Services }}
		{{-  if .Endpoints }}
		{{ .Service.VarName }}Endpoints *{{.Service.PkgName}}.Endpoints
		{{-  end }}
	{{- end }}
	)
	{
	{{- range .Services }}{{ $svc := . }}
		{{-  if .Endpoints }}
		{{ .Service.VarName }}Endpoints = {{ .Service.PkgName }}.NewEndpoints({{ .Service.VarName }}Svc)
		{{-  end }}
	{{- end }}
	}
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainEncodingT = `
	// Provide the transport specific request decoder and response encoder.
	// The goa http package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions,
	// see goa.design/encoding.
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainMuxT = `
	// Build the service HTTP request multiplexer and configure it to serve
	// HTTP requests to the service endpoints.
	var mux goahttp.Muxer
	{
		mux = goahttp.NewMuxer()
	}
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainServerInitT = `
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
		eh := ErrorHandler(logger)
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

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainMiddlewareT = `
	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		if *dbg {
			handler = middleware.Debug(mux, os.Stdout)(handler)
		}
		handler = middleware.Log(adapter)(handler)
		handler = middleware.RequestID()(handler)
	}
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainHTTPT = `
	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)
	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the service to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		errc <- fmt.Errorf("%s", <-c)
	}()
	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: *addr, Handler: handler}
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
		logger.Printf("listening on %s", *addr)
		errc <- srv.ListenAndServe()
	}()
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainEndT = `
	// Wait for signal.
	logger.Printf("exiting (%v)", <-errc)
	// Shutdown gracefully with a 30s timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	logger.Println("exited")
}
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "DefaultHost": string}
const mainErrorHandlerT = `
// ErrorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func ErrorHandler(logger *log.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(middleware.RequestIDKey).(string)
		w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Printf("[%s] ERROR: %s", id, err.Error())
	}
}
`
