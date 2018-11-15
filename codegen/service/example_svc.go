package service

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

type (
	// transport is a type for supported transports.
	transport int

	// TransportData contains the data about a transport. It is used only by the
	// example codegen functions.
	TransportData struct {
		// Type is the transport type.
		Type transport
		// Services is the list of services that support the transport.
		Services []string
		// Addr is the default address of the service implementing the transport.
		Host string
		// Port is the default listening port for the transport.
		Port string
		// IsDefault indicates the default transport.
		IsDefault bool
	}

	// dummyEndpointData contains the data needed to render dummy endpoint
	// implementation in the dummy service file.
	dummyEndpointData struct {
		*MethodData
		// ServiceVarName is the service variable name.
		ServiceVarName string
		// PayloadFullRef is the fully qualified reference to the payload.
		PayloadFullRef string
		// ResultFullRef is the fully qualified reference to the result.
		ResultFullRef string
		// ResultIsStruct indicates that the result type is a struct.
		ResultIsStruct bool
		// ResultView is the view to render the result. It is set only if the
		// result type uses views.
		ResultView string
	}
)

const (
	// TransportHTTP is the HTTP transport.
	TransportHTTP transport = iota + 1
)

// Name returns the name of the transport.
func (t *TransportData) Name() string {
	switch t.Type {
	case TransportHTTP:
		return "http"
	}
	return ""
}

// DisplayName returns the display name for the transport.
func (t *TransportData) DisplayName() string {
	switch t.Type {
	case TransportHTTP:
		return "HTTP"
	}
	return ""
}

// URL returns the URL for the transport. It panics if the URL is invalid.
func (t *TransportData) URL() string {
	u, err := url.Parse(t.Host + ":" + t.Port)
	if err != nil {
		panic("invalid URL: " + err.Error())
	}
	return u.String()
}

// ExampleServiceFiles returns a dummy service implementation and
// example service main.go.
func ExampleServiceFiles(genpkg string, root *design.RootExpr, transports []*TransportData) []*codegen.File {
	var fw []*codegen.File
	for _, svc := range root.Services {
		if f := dummyServiceFile(genpkg, root, svc); f != nil {
			fw = append(fw, f)
		}
	}
	if m := exampleMain(genpkg, root, transports); m != nil {
		fw = append(fw, m)
	}
	return fw
}

// dummyServiceFile returns a dummy implementation of the given service.
func dummyServiceFile(genpkg string, root *design.RootExpr, svc *design.ServiceExpr) *codegen.File {
	path := codegen.SnakeCase(svc.Name) + ".go"
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	data := Services.Get(svc.Name)
	apiPkg := strings.ToLower(codegen.Goify(root.API.Name, false))
	sections := []*codegen.SectionTemplate{
		codegen.Header("", apiPkg, []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "log"},
			{Path: filepath.Join(genpkg, codegen.SnakeCase(svc.Name)), Name: data.PkgName},
		}),
		{
			Name:   "dummy-service",
			Source: dummyServiceStructT,
			Data:   data,
		},
	}
	for _, m := range svc.Methods {
		md := data.Method(m.Name)
		ed := &dummyEndpointData{
			MethodData:     md,
			ServiceVarName: data.VarName,
		}
		if m.Payload.Type != design.Empty {
			ed.PayloadFullRef = data.Scope.GoFullTypeRef(m.Payload, data.PkgName)
		}
		if m.Result.Type != design.Empty {
			ed.ResultFullRef = data.Scope.GoFullTypeRef(m.Result, data.PkgName)
			ed.ResultIsStruct = design.IsObject(m.Result.Type)
			if md.ViewedResult != nil {
				view := "default"
				if m.Result.Meta != nil {
					if v, ok := m.Result.Meta["view"]; ok {
						view = v[0]
					}
				}
				ed.ResultView = view
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "dummy-endpoint",
			Source: dummyEndpointImplT,
			Data:   ed,
		})
	}

	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
	}
}

func exampleMain(genpkg string, root *design.RootExpr, transports []*TransportData) *codegen.File {
	mainPath := filepath.Join("cmd", codegen.SnakeCase(codegen.Goify(root.API.Name, true))+"_svc", "main.go")
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	idx := strings.LastIndex(genpkg, string(os.PathSeparator))
	rootPath := "."
	if idx > 0 {
		rootPath = genpkg[:idx]
	}
	apiPkg := strings.ToLower(codegen.Goify(root.API.Name, false))
	specs := []*codegen.ImportSpec{
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "log"},
		{Path: "os"},
		{Path: "os/signal"},
		{Path: rootPath, Name: apiPkg},
	}
	svcdata := make([]*Data, 0, len(root.Services))
	for _, svc := range root.Services {
		sd := Services.Get(svc.Name)
		svcdata = append(svcdata, sd)
		specs = append(specs, &codegen.ImportSpec{
			Path: filepath.Join(genpkg, codegen.SnakeCase(svc.Name)),
			Name: sd.PkgName,
		})
	}
	data := map[string]interface{}{
		"Services":   svcdata,
		"APIPkg":     apiPkg,
		"Transports": transports,
	}
	sections := []*codegen.SectionTemplate{codegen.Header("", "main", specs)}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "service-main",
		Source: mainT,
		Data:   data,
		FuncMap: map[string]interface{}{
			"transportSupported": transportSupported,
		},
	})

	return &codegen.File{Path: mainPath, SectionTemplates: sections}
}

// transportSupported returns true if the service name supports the given
// transport.
func transportSupported(svcName string, t *TransportData) bool {
	for _, s := range t.Services {
		if s == svcName {
			return true
		}
	}
	return false
}

// input: Data
const dummyServiceStructT = `{{ printf "%s service example implementation.\nThe example methods log the requests and return zero values." .Name | comment }}
type {{ .VarName }}Svc struct {
	logger *log.Logger
}

{{ printf "New%s returns the %s service implementation." .StructName .Name | comment }}
func New{{ .StructName }}(logger *log.Logger) {{ .PkgName }}.Service {
	return &{{ .VarName }}Svc{logger}
}
`

// input: endpointData
const dummyEndpointImplT = `{{ comment .Description }}
{{- if .ServerStream }}
func (s *{{ .ServiceVarName }}Svc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}, stream {{ .ServerStream.Interface }}) (err error) {
{{- else }}
func (s *{{ .ServiceVarName }}Svc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}) ({{ if .ResultFullRef }}res {{ .ResultFullRef }}, {{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }} {{ end }}err error) {
{{- end }}
{{- if and (and .ResultFullRef .ResultIsStruct) (not .ServerStream) }}
	res = &{{ .Result }}{}
{{- end }}
{{- if .ResultView }}
	{{- if .ServerStream }}
	stream.SetView({{ printf "%q" .Result.View }})
	{{- else }}
	view = {{ printf "%q" .ResultView }}
	{{- end }}
{{- end }}
	s.logger.Print("{{ .ServiceVarName }}.{{ .Name }}")
	return
}
`

// input: map[string]interface{}{"Services":[]*Data, "APIPkg": string, "Transports" []*TransportData}
const mainT = `func main() {
  // Define command line flags, add any other flag required to configure
  // the service.
  var (
  {{- range .Transports }}
    {{ .Name }}AddrF = flag.String("{{ .Name }}-listen", "{{ printf ":%s" .Port }}", "{{ .DisplayName }} listen ` + "`" + `address` + "`" + `")
  {{- end }}
    dbgF  = flag.Bool("debug", false, "Log request and response bodies")
  )
  flag.Parse()

  // Setup logger and goa log adapter. Replace logger with your own using
  // your log package of choice. The goa.design/middleware/logging/...
  // packages define log adapters for common log packages.
  var (
    logger *log.Logger
  )
  {
    logger = log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)
  }

	// Create the structs that implement the services.
	var (
	{{- range .Services }}
		{{- if .Methods }}
		{{ .VarName }}Svc {{ .PkgName }}.Service
		{{- end }}
	{{- end }}
	)
	{
	{{- range .Services }}
		{{- if .Methods }}
		{{ .VarName }}Svc = {{ $.APIPkg }}.New{{ .StructName }}(logger)
		{{- end }}
	{{- end }}
	}

	// Wrap the services in endpoints that can be invoked from other
	// services potentially running in different processes.
	var (
	{{- range .Services }}
		{{- if .Methods }}
		{{ .VarName }}Endpoints *{{ .PkgName }}.Endpoints
		{{- end }}
	{{- end }}
	)
	{
	{{- range .Services }}{{ $svc := . }}
		{{- if .Methods }}
		{{ .VarName }}Endpoints = {{ .PkgName }}.NewEndpoints({{ .VarName }}Svc{{ range .Schemes }}, {{ $.APIPkg }}.{{ $svc.StructName }}{{ .Type }}Auth{{ end }})
		{{- end }}
	{{- end }}
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

{{- range $t := .Transports }}
  {{ $t.Name }}Srvr := {{ $t.Name }}Serve(*{{ $t.Name }}AddrF, {{ range $.Services }}{{ if and .Methods (transportSupported .Name $t) }}{{ .VarName }}Endpoints, {{ end }}{{ end }}errc, logger, *dbgF)
{{- end }}

	// Wait for signal.
	logger.Printf("exiting (%v)", <-errc)
{{- range .Transports }}
	logger.Println("Shutting down {{ .DisplayName }} server at " + *{{ .Name }}AddrF)
  {{ .Name }}Stop({{ .Name }}Srvr)
{{- end }}
	logger.Println("exited")
}
`
