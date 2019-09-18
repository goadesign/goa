package example

import (
	"os"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// CLIFiles returns example client tool main implementation for each server
// expression in the design.
func CLIFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, svr := range root.API.Servers {
		if m := exampleCLIMain(genpkg, root, svr); m != nil {
			fw = append(fw, m)
		}
	}
	return fw
}

// exampleCLIMain returns an example client tool main implementation for the
// given server expression.
func exampleCLIMain(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr) *codegen.File {
	svrdata := Servers.Get(svr)

	path := filepath.Join("cmd", svrdata.Dir+"-cli", "main.go")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "encoding/json"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "strings"},
		codegen.GoaImport(""),
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		&codegen.SectionTemplate{
			Name:   "cli-main-start",
			Source: cliMainStartT,
			Data: map[string]interface{}{
				"Server": svrdata,
			},
			FuncMap: map[string]interface{}{
				"join": strings.Join,
			},
		},
		&codegen.SectionTemplate{
			Name:   "cli-main-var-init",
			Source: cliMainVarInitT,
			Data: map[string]interface{}{
				"Server": svrdata,
			},
			FuncMap: map[string]interface{}{
				"join": strings.Join,
			},
		},
		&codegen.SectionTemplate{
			Name:   "cli-main-endpoint-init",
			Source: cliMainEndpointInitT,
			Data: map[string]interface{}{
				"Server": svrdata,
			},
			FuncMap: map[string]interface{}{
				"join":    strings.Join,
				"toUpper": strings.ToUpper,
			},
		},
		&codegen.SectionTemplate{Name: "cli-main-end", Source: cliMainEndT},
		&codegen.SectionTemplate{
			Name:   "cli-main-usage",
			Source: cliMainUsageT,
			Data: map[string]interface{}{
				"APIName": root.API.Name,
				"Server":  svrdata,
			},
			FuncMap: map[string]interface{}{
				"toUpper": strings.ToUpper,
				"join":    strings.Join,
			},
		},
	}
	return &codegen.File{Path: path, SectionTemplates: sections, SkipExist: true}
}

const (
	// input: map[string]interface{}{"Server": *Data}
	cliMainStartT = `func main() {
	var (
		hostF = flag.String("host", {{ printf "%q" .Server.DefaultHost.Name }}, "Server host (valid values: {{ (join .Server.AvailableHosts ", ") }})")
		addrF = flag.String("url", "", "URL to service host")
	{{ range .Server.Variables }}
		{{ .VarName }}F = flag.String({{ printf "%q" .Name }}, {{ printf "%q" .DefaultValue }}, {{ printf "%q" .Description }})
	{{- end }}
		verboseF = flag.Bool("verbose", false, "Print request and response details")
		vF = flag.Bool("v", false, "Print request and response details")
		timeoutF = flag.Int("timeout", 30, "Maximum number of seconds to wait for response")
	)
	flag.Usage = usage
	flag.Parse()
`

	// input: map[string]interface{}{"Server": *Data}
	cliMainVarInitT = `var (
		addr string
		timeout int
		debug bool
	)
	{
		addr = *addrF
		if addr == "" {
			switch *hostF {
		{{- range $h := .Server.Hosts }}
			case {{ printf "%q" $h.Name }}:
				addr = {{ printf "%q" ($h.DefaultURL $.Server.DefaultTransport.Type) }}
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
						fmt.Fprintf(os.Stderr, "invalid value for URL '{{ .Name }}' variable: %q (valid values: {{ join .Values "," }})\n", *{{ .VarName }}F)
						os.Exit(1)
					}
				{{- end }}
				addr = strings.Replace(addr, {{ printf "\"{%s}\"" .Name }}, *{{ .VarName }}F, -1)
			{{- end }}
		{{- end }}
			default:
				fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: {{ join .Server.AvailableHosts "|" }})\n", *hostF)
				os.Exit(1)
			}
		}
		timeout = *timeoutF
		debug = *verboseF || *vF
	}

	var (
		scheme string
		host string
	)
	{
		u, err := url.Parse(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", addr, err)
			os.Exit(1)
		}
		scheme = u.Scheme
		host = u.Host
	}
`

	// input: map[string]interface{}{"Server": *Data}
	cliMainEndpointInitT = `var(
		endpoint goa.Endpoint
		payload interface{}
		err error
	)
	{
		switch scheme {
	{{- range $t := .Server.Transports }}
		case "{{ $t.Type }}", "{{ $t.Type }}s":
			endpoint, payload, err = do{{ toUpper $t.Name }}(scheme, host, timeout, debug)
	{{- end }}
		default:
			fmt.Fprintf(os.Stderr, "invalid scheme: %q (valid schemes: {{ join .Server.Schemes "|" }})\n", scheme)
			os.Exit(1)
		}
	}
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}
`

	cliMainEndT = `
	data, err := endpoint(context.Background(), payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if data != nil {
		m, _ := json.MarshalIndent(data, "", "    ")
		fmt.Println(string(m))
	}
}
`

	// input: map[string]interface{}{"APIName": string, "Server": *Data}
	cliMainUsageT = `
func usage() {
  fmt.Fprintf(os.Stderr, ` + "`" + `%s is a command line client for the {{ .APIName }} API.

Usage:
    %s [-host HOST][-url URL][-timeout SECONDS][-verbose|-v]{{ range .Server.Variables }}[-{{ .Name }} {{ toUpper .Name }}]{{ end }} SERVICE ENDPOINT [flags]

    -host HOST:  server host ({{ .Server.DefaultHost.Name }}). valid values: {{ (join .Server.AvailableHosts ", ") }}
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)
	{{- range .Server.Variables }}
    -{{ .Name }}:    {{ .Description }} ({{ .DefaultValue }})
	{{- end }}

Commands:
%s
Additional help:
    %s SERVICE [ENDPOINT] --help

Example:
%s
` + "`" + `, os.Args[0], os.Args[0], indent({{ .Server.DefaultTransport.Type }}UsageCommands()), os.Args[0], indent({{ .Server.DefaultTransport.Type }}UsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
`
)
