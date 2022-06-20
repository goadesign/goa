package example

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// ServerFiles returns an example server main implementation for every server
// expression in the service design.
func ServerFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
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
	svrdata := Servers.Get(svr)
	mainPath := filepath.Join("cmd", svrdata.Dir, "main.go")
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "log"},
		{Path: "net"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "os/signal"},
		{Path: "strings"},
		{Path: "sync"},
		{Path: "syscall"},
		{Path: "time"},
		codegen.GoaImport("middleware"),
	}

	// Iterate through services listed in the server expression.
	svcData := make([]*service.Data, len(svr.Services))
	scope := codegen.NewNameScope()
	for i, svc := range svr.Services {
		sd := service.Services.Get(svc)
		svcData[i] = sd
		specs = append(specs, &codegen.ImportSpec{
			Path: path.Join(genpkg, sd.PathName),
			Name: scope.Unique(sd.PkgName),
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

	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "server-main-start",
			Source: mainStartT,
			Data: map[string]interface{}{
				"Server": svrdata,
			},
			FuncMap: map[string]interface{}{
				"join": strings.Join,
			},
		}, {
			Name:   "server-main-logger",
			Source: mainLoggerT,
			Data: map[string]interface{}{
				"APIPkg": apiPkg,
			},
		}, {
			Name:   "server-main-services",
			Source: mainSvcsT,
			Data: map[string]interface{}{
				"APIPkg":   apiPkg,
				"Services": svcData,
			},
			FuncMap: map[string]interface{}{
				"mustInitServices": mustInitServices,
			},
		}, {
			Name:   "server-main-endpoints",
			Source: mainEndpointsT,
			Data: map[string]interface{}{
				"Services": svcData,
			},
			FuncMap: map[string]interface{}{
				"mustInitServices": mustInitServices,
			},
		}, {
			Name:   "server-main-interrupts",
			Source: mainInterruptsT,
		}, {
			Name:   "server-main-handler",
			Source: mainServerHndlrT,
			Data: map[string]interface{}{
				"Server":   svrdata,
				"Services": svcData,
			},
			FuncMap: map[string]interface{}{
				"goify":   codegen.Goify,
				"join":    strings.Join,
				"toUpper": strings.ToUpper,
			},
		},
		{
			Name:   "server-main-end",
			Source: mainEndT,
		},
	}

	return &codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}
}

// mustInitServices returns true if at least one of the services defines methods.
// It is used by the template to initialize service variables.
func mustInitServices(data []*service.Data) bool {
	for _, svc := range data {
		if len(svc.Methods) > 0 {
			return true
		}
	}
	return false
}

const (
	// input: map[string]interface{"Server": *ServerData}
	mainStartT = `
func main() {
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
	mainLoggerT = `
	{{ comment "Setup logger. Replace logger with your own log package of choice." }}
	var (
		logger *log.Logger
	)
	{
		logger = log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)
	}
`

	// input: map[string]interface{"APIPkg": string, "Services": []*service.Data}
	mainSvcsT = `
{{- if mustInitServices .Services }}
	{{ comment "Initialize the services." }}
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
{{- end }}
`

	// input: map[string]interface{"Services": []*service.Data}
	mainEndpointsT = `
{{- if mustInitServices .Services }}
	{{ comment "Wrap the services in endpoints that can be invoked from other services potentially running in different processes." }}
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
{{- end }}
`

	mainInterruptsT = `
	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
`

	// input: map[string]interface{"Server": *Data, "Services": []*service.Data}
	mainServerHndlrT = `
	{{ comment "Start the servers and send errors (if any) to the error channel." }}
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
						logger.Fatalf("invalid value for URL '{{ .Name }}' variable: %q (valid values: {{ join .Values "," }})\n", *{{ .VarName }}F)
					}
				{{- end }}
				addr = strings.Replace(addr, {{ printf "\"{%s}\"" .Name }}, *{{ .VarName }}F, -1)
			{{- end }}
			u, err := url.Parse(addr)
			if err != nil {
				logger.Fatalf("invalid URL %#v: %s\n", addr, err)
			}
			if *secureF {
				u.Scheme = "{{ $u.Transport.Type }}s"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *{{ $u.Transport.Type }}PortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					logger.Fatalf("invalid URL %#v: %s\n", u.Host, err)
				}
				u.Host = net.JoinHostPort(h, *{{ $u.Transport.Type }}PortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "{{ $u.Port }}")
			}
			handle{{ toUpper $u.Transport.Name }}Server(ctx, u, {{ range $t := $.Server.Transports }}{{ if eq $t.Type $u.Transport.Type }}{{ range $s := $t.Services }}{{ range $.Services }}{{ if eq $s .Name }}{{ if .Methods }}{{ .VarName }}Endpoints, {{ end }}{{ end }}{{ end }}{{ end }}{{ end }}{{ end }}&wg, errc, logger, *dbgF)
		}
	{{- end }}
	{{ end }}
{{- end }}
	default:
		logger.Fatalf("invalid host argument: %q (valid hosts: {{ join .Server.AvailableHosts "|" }})\n", *hostF)
	}
`

	mainEndT = `
	{{ comment "Wait for signal." }}
	logger.Printf("exiting (%v)", <-errc)

	{{ comment "Send cancellation signal to the goroutines." }}
	cancel()

	wg.Wait()
	logger.Println("exited")
}
`
)
