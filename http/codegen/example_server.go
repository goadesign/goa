package codegen

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	httpdesign "goa.design/goa/http/design"
)

// ExampleServerFiles returns and example main and dummy service
// implementations.
func ExampleServerFiles(genpkg string, root *httpdesign.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, svc := range root.HTTPServices {
		f := dummyServiceFile(genpkg, root, svc)
		if f != nil {
			fw = append(fw, f)
		}
	}
	if m := exampleMain(genpkg, root); m != nil {
		fw = append(fw, m)
	}
	return fw
}

// dummyServiceFile returns a dummy implementation of the given service.
func dummyServiceFile(genpkg string, root *httpdesign.RootExpr, svc *httpdesign.ServiceExpr) *codegen.File {
	path := codegen.SnakeCase(svc.Name()) + ".go"
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	data := HTTPServices.Get(svc.Name())
	apiPkg := strings.ToLower(codegen.Goify(root.Design.API.Name, false))
	sections := []*codegen.SectionTemplate{
		codegen.Header("", apiPkg, []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "log"},
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
	}

	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
	}
}

func exampleMain(genpkg string, root *httpdesign.RootExpr) *codegen.File {
	mainPath := filepath.Join("cmd", codegen.SnakeCase(root.Design.API.Name)+"svc", "main.go")
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	idx := strings.LastIndex(genpkg, string(os.PathSeparator))
	rootPath := "."
	if idx > 0 {
		rootPath = genpkg[:idx]
	}
	apiPkg := strings.ToLower(codegen.Goify(root.Design.API.Name, false))
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
		{Path: "goa.design/goa/http/middleware/debugging"},
		{Path: "goa.design/goa/http/middleware/logging"},
		{Path: rootPath, Name: apiPkg},
	}
	for _, svc := range root.HTTPServices {
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
	var svcdata []*ServiceData
	for _, svc := range root.HTTPServices {
		svcdata = append(svcdata, HTTPServices.Get(svc.Name()))
	}
	data := map[string]interface{}{
		"Services": svcdata,
		"APIPkg":   apiPkg,
	}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main",
		Source: mainT,
		Data:   data,
	})

	return &codegen.File{Path: mainPath, SectionTemplates: sections}
}

// input: ServiceData
const dummyServiceStructT = `{{ printf "%s service example implementation.\nThe example methods log the requests and return zero values." .Service.Name | comment }}
type {{ .Service.PkgName }}Svc struct {
	logger *log.Logger
}

{{ printf "New%s returns the %s service implementation." .Service.VarName .Service.Name | comment }}
func New{{ .Service.VarName }}(logger *log.Logger) {{ .Service.PkgName }}.Service {
	return &{{ .Service.PkgName }}Svc{logger}
}
`

// input: EndpointData
const dummyEndpointImplT = `{{ comment .Method.Description }}
func (s *{{ .ServicePkgName }}Svc) {{ .Method.VarName }}(ctx context.Context{{ if .Payload.Ref }}, p {{ .Payload.Ref }}{{ end }}) ({{ if .Result.Ref }}{{ .Result.Ref }}, {{ end }}error) {
{{- if and .Result.Ref .Result.IsStruct }}
	res := &{{ .Result.Name }}{}
{{- else if .Result.Ref }}
	var res {{ .Result.Ref }}
{{- end }}
	s.logger.Print("{{ .ServiceName }}.{{ .Method.Name }}")
	return {{ if .Result.Ref }}res, {{ end }}nil
}
`

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string}
const mainT = `func main() {
	// Define command line flags, add any other flag required to configure
	// the service.
	var (
		addr = flag.String("listen", ":8080", "HTTP listen ` + "`" + `address` + "`" + `")
		dbg  = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()

	// Setup logger and goa log adapter. Replace logger with your own using
	// your log package of choice. The goa.design/middleware/logging/...
	// packages define log adapters for common log packages.
	var (
		logger  *log.Logger
		adapter logging.Adapter
	)
	{
		logger = log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)
		adapter = logging.Adapt(logger)
	}

	// Create the structs that implement the services.
	var (
	{{- range .Services }}
		{{-  if .Endpoints }}
		{{ .Service.PkgName }}Svc {{.Service.PkgName}}.Service
		{{-  end }}
	{{- end }}
	)
	{
	{{- range .Services }}
		{{-  if .Endpoints }}
		{{ .Service.PkgName }}Svc = {{ $.APIPkg }}.New{{ .Service.VarName }}(logger)
		{{-  end }}
	{{- end }}
	}

	// Wrap the services in endpoints that can be invoked from other
	// services potentially running in different processes.
	var (
	{{- range .Services }}
		{{-  if .Endpoints }}
		{{ .Service.PkgName }}Endpoints *{{.Service.PkgName}}.Endpoints
		{{-  end }}
	{{- end }}
	)
	{
	{{- range .Services }}
		{{-  if .Endpoints }}
		{{ .Service.PkgName }}Endpoints = {{ .Service.PkgName }}.NewEndpoints({{ .Service.PkgName }}Svc)
		{{-  end }}
	{{- end }}
	}

	// Provide the transport specific request decoder and response encoder.
	// The goa http package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions,
	// see goa.design/encoding.
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)

	// Build the service HTTP request multiplexer and configure it to serve
	// HTTP requests to the service endpoints.
	var mux goahttp.Muxer
	{
		mux = goahttp.NewMuxer()
	}

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	var (
	{{- range .Services }}
		{{ .Service.PkgName }}Server *{{.Service.PkgName}}svr.Server
	{{- end }}
	)
	{
	{{- range .Services }}
		{{-  if .Endpoints }}
		{{ .Service.PkgName }}Server = {{ .Service.PkgName }}svr.New({{ .Service.PkgName }}Endpoints, mux, dec, enc)
		{{-  else }}
		{{ .Service.PkgName }}Server = {{ .Service.PkgName }}svr.New(nil, mux, dec, enc)
		{{-  end }}
	{{- end }}
	}

	// Configure the mux.
	{{- range .Services }}
	{{ .Service.PkgName }}svr.Mount(mux{{ if .Endpoints }}, {{ .Service.PkgName }}Server{{ end }})
	{{- end }}

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		if *dbg {
			handler = debugging.New(mux, adapter)(handler)
		}
		handler = logging.New(adapter)(handler)
	}

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
		for _, m := range {{ .Service.PkgName }}Server.Mounts {
			{{- if .FileServers }}
		  logger.Printf("[INFO] service %q file %q mounted on %s %s", {{ .Service.PkgName }}Server.Service(), m.Method, m.Verb, m.Pattern)
			{{- else }}
			logger.Printf("[INFO] service %q method %q mounted on %s %s", {{ .Service.PkgName }}Server.Service(), m.Method, m.Verb, m.Pattern)
			{{- end }}
		}
		{{- end }}
		logger.Printf("[INFO] listening on %s", *addr)
		errc <- srv.ListenAndServe()
	}()

	// Wait for signal.
	logger.Printf("exiting (%v)", <-errc)

	// Shutdown gracefully with a 30s timeout.
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	srv.Shutdown(ctx)

	logger.Println("exited")
}
`
