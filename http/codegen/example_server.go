package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
	httpdesign "goa.design/goa.v2/http/design"
)

// ExampleServerFiles returns and example main and dummy service
// implementations.
func ExampleServerFiles(root *httpdesign.RootExpr) []codegen.File {
	fw := make([]codegen.File, len(root.HTTPServices)+1)
	for i, svc := range root.HTTPServices {
		fw[i] = dummyServiceFile(svc)
	}
	fw[len(root.HTTPServices)] = exampleMain(root)
	return fw
}

// dummyServiceFile returns a dummy implementation of the given service.
func dummyServiceFile(svc *httpdesign.ServiceExpr) codegen.File {
	path := codegen.SnakeCase(svc.Name()) + ".go"
	data := HTTPServices.Get(svc.Name())
	sections := func(genPkg string) []*codegen.Section {
		s := []*codegen.Section{
			codegen.Header("", codegen.KebabCase(design.Root.API.Name), []*codegen.ImportSpec{
				{Path: "context"},
				{Path: "log"},
				{Path: genPkg + "/" + codegen.Goify(svc.Name(), false)},
			}),
			{Template: dummyServiceStructTmpl(svc), Data: data},
		}
		for _, e := range data.Endpoints {
			s = append(s, &codegen.Section{Template: dummyEndpointImplTmpl(svc), Data: e})
		}

		return s
	}

	return codegen.NewSource(path, sections)
}

func exampleMain(root *httpdesign.RootExpr) codegen.File {
	path := filepath.Join("cmd", codegen.SnakeCase(root.Design.API.Name)+"svc", "main.go")
	sections := func(genPkg string) []*codegen.Section {
		idx := strings.LastIndex(genPkg, string(os.PathSeparator))
		rootPath := "."
		if idx > 0 {
			rootPath = genPkg[:idx]
		}
		specs := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "flag"},
			{Path: "fmt"},
			{Path: "log"},
			{Path: "net/http"},
			{Path: "os"},
			{Path: "os/signal"},
			{Path: "time"},
			{Path: "goa.design/goa.v2", Name: "goa"},
			{Path: "goa.design/goa.v2/http", Name: "goahttp"},
			{Path: "goa.design/goa.v2/http/middleware/debugging"},
			{Path: "goa.design/goa.v2/http/middleware/logging"},
			{Path: rootPath, Name: codegen.KebabCase(root.Design.API.Name)},
		}
		for _, svc := range root.HTTPServices {
			pkgName := HTTPServices.Get(svc.Name()).Service.PkgName
			specs = append(specs, &codegen.ImportSpec{
				Path: filepath.Join(genPkg, "http", pkgName, "server"),
				Name: pkgName + "svr",
			})
			specs = append(specs, &codegen.ImportSpec{
				Path: filepath.Join(genPkg, pkgName),
			})
		}
		s := []*codegen.Section{
			codegen.Header("", "main", specs),
		}
		var svcdata []*ServiceData
		for _, svc := range root.HTTPServices {
			svcdata = append(svcdata, HTTPServices.Get(svc.Name()))
		}
		data := map[string]interface{}{
			"Services": svcdata,
			"APIPkg":   codegen.KebabCase(root.Design.API.Name),
		}
		s = append(s, &codegen.Section{Template: mainTmpl, Data: data})

		return s
	}

	return codegen.NewSource(path, sections)
}

func dummyServiceStructTmpl(r *httpdesign.ServiceExpr) *template.Template {
	return template.Must(transTmpl(r).New("dummy-service").Parse(dummyServiceStructT))
}

func dummyEndpointImplTmpl(r *httpdesign.ServiceExpr) *template.Template {
	return template.Must(transTmpl(r).New("dummy-endpoint").Parse(dummyEndpointImplT))
}

var mainTmpl = template.Must(template.New("server-main").Parse(mainT))

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
{{- if .Result.Ref }}
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
	// your log package of choice. The goa.design/logging package defines
	// log adapters for common log packages. Writing adapters for other log
	// packages is very simple as well.
	var (
		logger  *log.Logger
		adapter goa.LogAdapter
	)
	{
		logger = log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)
		adapter = goa.AdaptStdLogger(logger)
	}

	// Create the structs that implement the services.
	var (
	{{- range .Services }}
		{{ .Service.PkgName }}Svc {{.Service.PkgName}}.Service
	{{- end }}
	)
	{
	{{- range .Services }}
		{{ .Service.PkgName }}Svc = {{ $.APIPkg }}.New{{ .Service.VarName }}(logger)
	{{- end }}
	}

	// Wrap the services in endpoints that can be invoked from other
	// services potentially running in different processes.
	var (
	{{- range .Services }}
		{{ .Service.PkgName }}Endpoints *{{.Service.PkgName}}.Endpoints
	{{- end }}
	)
	{
	{{- range .Services }}
		{{ .Service.PkgName }}Endpoints = {{ .Service.PkgName }}.NewEndpoints({{ .Service.PkgName }}Svc)
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
		{{ .Service.PkgName }}Server = {{ .Service.PkgName }}svr.New({{ .Service.PkgName }}Endpoints, mux, dec, enc)
	{{- end }}
	}

	// Configure the mux.
	{{- range .Services }}
	{{ .Service.PkgName }}svr.Mount(mux, {{ .Service.PkgName }}Server)
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
