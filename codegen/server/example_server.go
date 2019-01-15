package server

import (
	"os"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service"
	"goa.design/goa/expr"
)

// ExampleServerFiles returns an example server main implementation for every
// server expression in the service design.
func ExampleServerFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range root.API.Servers {
		if m := exampleSvrMain(genpkg, root, svr); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// exampleSvrMain returns the default main function for the given server
// expression.
func exampleSvrMain(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr) *codegen.File {
	pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	mainPath := filepath.Join("cmd", pkg, "main.go")
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
		{Path: "context"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "log"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "os/signal"},
		{Path: "strings"},
		{Path: "sync"},
		{Path: "time"},
		{Path: rootPath, Name: apiPkg},
	}

	svrData := Servers.Get(svr)

	// Iterate through services listed in the server expression.
	svcData := make([]*service.Data, len(svr.Services))
	for i, svc := range svr.Services {
		sd := service.Services.Get(svc)
		svcData[i] = sd
		specs = append(specs, &codegen.ImportSpec{
			Path: filepath.Join(genpkg, codegen.SnakeCase(svc)),
			Name: sd.PkgName,
		})
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		&codegen.SectionTemplate{
			Name:   "server-main-start",
			Source: mainStartT,
			Data: map[string]interface{}{
				"Server": svrData,
			},
			FuncMap: map[string]interface{}{
				"join": strings.Join,
			},
		},
		&codegen.SectionTemplate{
			Name:   "server-main-logger",
			Source: mainLoggerT,
			Data: map[string]interface{}{
				"APIPkg": apiPkg,
			},
		},
		&codegen.SectionTemplate{
			Name:   "server-main-services",
			Source: mainSvcsT,
			Data: map[string]interface{}{
				"APIPkg":   apiPkg,
				"Services": svcData,
			},
		},
		&codegen.SectionTemplate{
			Name:   "server-main-endpoints",
			Source: mainEndpointsT,
			Data: map[string]interface{}{
				"Services": svcData,
			},
		},
		&codegen.SectionTemplate{Name: "server-main-interrupts", Source: mainInterruptsT},
		&codegen.SectionTemplate{
			Name:   "server-main-handler",
			Source: mainServerHndlrT,
			Data: map[string]interface{}{
				"Server":   svrData,
				"Services": svcData,
			},
			FuncMap: map[string]interface{}{
				"goify":   codegen.Goify,
				"join":    strings.Join,
				"toUpper": strings.ToUpper,
			},
		},
		&codegen.SectionTemplate{Name: "server-main-end", Source: mainEndT},
	}

	return &codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}
}

const (
	// input: map[string]interface{"Server": *ServerData}
	mainStartT = `func main() {
	{{ comment "Define command line flags, add any other flag required to configure the service." }}
	var(
		hostF = flag.String("host", {{ printf "%q" .Server.DefaultHost.Name }}, "Server host (valid values: {{ (join .Server.AvailableHosts ", ") }})")
		domainF = flag.String("domain", "", "Host domain name (overrides host domain specified in service design)")
	{{- range .Server.Transports }}
	{{ .Type }}PortF = flag.String("{{ .Type }}-port", "", "{{ .Name }} port (overrides host {{ .Name }} port specified in service design)")
	{{- end }}
	{{- range .Server.Variables }}
	{{ .VarName }}F = flag.String({{ printf "%q" .Name }}, {{ printf "%q" .DefaultValue }}, "{{ .Description }}{{ if .Values }} (valid values: {{ join .Values ", " }}){{ end }}")
	{{- end }}
		secureF = flag.Bool("secure", false, "Use secure scheme (https or grpcs)")
		dbgF  = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()
`

	// input: map[string]interface{"APIPkg": string}
	mainLoggerT = `// Setup logger and goa log adapter. Replace logger with your own using
	// your log package of choice. The goa.design/middleware/logging/...
	// packages define log adapters for common log packages.
	var (
		logger *log.Logger
	)
	{
		logger = log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)
	}
`

	// input: map[string]interface{"APIPkg": string, "Services": []*service.Data}
	mainSvcsT = `{{ comment "Initialize the services." }}
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
`

	// input: map[string]interface{"Services": []*service.Data}
	mainEndpointsT = `{{ comment "Wrap the services in endpoints that can be invoked from other services potentially running in different processes." }}
	var (
	{{- range .Services }}
		{{- if .Methods }}
		{{ .VarName }}Endpoints *{{ .PkgName }}.Endpoints
		{{- end }}
	{{- end }}
	)
	{
	{{- range .Services }}
		{{- if .Methods }}
			{{ .VarName }}Endpoints = {{ .PkgName }}.NewEndpoints({{ .VarName }}Svc)
		{{- end }}
	{{- end }}
	}
`

	mainInterruptsT = `// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		errc <- fmt.Errorf("%s", <-c)
	}()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
`

	// input: map[string]interface{"Server": *Data, "Services": []*service.Data}
	mainServerHndlrT = `{{ comment "Start the servers and send errors (if any) to the error channel." }}
	switch *hostF {
{{- range $h := .Server.Hosts }}
	case {{ printf "%q" $h.Name }}:
	{{- range $u := $h.URIs }}
		{{- if $.Server.HasTransport $u.Transport.Type }}
		{
			addr := {{ printf "%q" $u.URL }}
			{{- range $h.Variables }}
				{{- if .Values }}
					var {{ .VarName }}Seen bool
					{
						for _, v := range []string{ {{ range $v := .Values }}"{{ $v }}",{{ end }} } {
							if v == *{{ .VarName }}F {
								{{ .VarName }}Seen = true
								break
							}
						}
					}
					if !{{ .VarName }}Seen {
						fmt.Fprintf(os.Stderr, "invalid value for URL '{{ .Name }}' variable: %q (valid values: {{ join .Values "," }})", *{{ .VarName }}F)
						os.Exit(1)
					}
				{{- end }}
				addr = strings.Replace(addr, {{ printf "\"{%s}\"" .Name }}, *{{ .VarName }}F, -1)
			{{- end }}
			u, err := url.Parse(addr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid URL %#v: %s", addr, err)
				os.Exit(1)
			}
			if *secureF {
				u.Scheme = "{{ $u.Transport.Type }}s"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *{{ $u.Transport.Type }}PortF != "" {
				h := strings.Split(u.Host, ":")[0]
				u.Host = h + ":" + *{{ $u.Transport.Type }}PortF
			} else if u.Port() == "" {
				u.Host += ":{{ $u.Port }}"
			}
			handle{{ toUpper $u.Transport.Name }}Server(ctx, u, {{ range $.Services }}{{ if .Methods }}{{ .VarName }}Endpoints, {{ end }}{{ end }}&wg, errc, logger, *dbgF)
		}
	{{- end }}
	{{ end }}
{{- end }}
	default:
		fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: {{ join .Server.AvailableHosts "|" }})", *hostF)
	}
`

	mainEndT = `{{ comment "Wait for signal." }}
	logger.Printf("exiting (%v)", <-errc)

	{{ comment "Send cancellation signal to the goroutines." }}
	cancel()

	wg.Wait()
	logger.Println("exited")
}
`
)
