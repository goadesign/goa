package codegen

import (
	"os"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
	httpdesign "goa.design/goa/http/design"
)

// ExampleCLI returns an example client tool main implementation.
func ExampleCLI(genpkg string, root *httpdesign.RootExpr) []*codegen.File {
	files := make([]*codegen.File, len(design.Root.API.Servers))
	for i, svr := range design.Root.API.Servers {
		pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
		apiPkg := strings.ToLower(codegen.Goify(root.Design.API.Name, false))
		path := filepath.Join("cmd", pkg+"-cli", "main.go")
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return nil // file already exists, skip it.
		}
		idx := strings.LastIndex(genpkg, string(os.PathSeparator))
		rootPath := "."
		if idx > 0 {
			rootPath = genpkg[:idx]
		}
		specs := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "encoding/json"},
			{Path: "flag"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "os"},
			{Path: "strings"},
			{Path: "time"},
			{Path: "github.com/gorilla/websocket"},
			{Path: "goa.design/goa/http", Name: "goahttp"},
			{Path: genpkg + "/http/cli/" + pkg, Name: "cli"},
			{Path: rootPath, Name: apiPkg},
		}
		svcdata := make([]*ServiceData, len(svr.Services))
		for i, svc := range svr.Services {
			svcdata[i] = HTTPServices.Get(svc)
		}
		vars := design.AsObject(svr.Hosts[0].Variables.Type)
		var variables []map[string]interface{}
		if len(*vars) > 0 {
			variables = make([]map[string]interface{}, len(*vars))
			for i, v := range *vars {
				def := v.Attribute.DefaultValue
				if def == nil {
					// DSL ensures v.Attribute has either a
					// default value or an enum validation
					def = v.Attribute.Validation.Values[0]
				}
				variables[i] = map[string]interface{}{
					"Name":         v.Name,
					"Description":  v.Attribute.Description,
					"VarName":      codegen.Goify(v.Name, false),
					"DefaultValue": def,
				}
			}
		}
		data := map[string]interface{}{
			"Services":   svcdata,
			"APIPkg":     apiPkg,
			"ServerName": svr.Name,
			"DefaultURL": svr.Hosts[0].URIs[0],
			"Variables":  variables,
		}
		sections := []*codegen.SectionTemplate{
			codegen.Header("", "main", specs),
			{
				Name:   "cli-main",
				Source: mainCLIT,
				Data:   data,
				FuncMap: map[string]interface{}{
					"needStreaming": needStreaming,
				},
			},
		}
		files[i] = &codegen.File{
			Path:             path,
			SectionTemplates: sections,
			SkipExist:        true,
		}
	}
	return files
}

// needStreaming returns true if at least one endpoint in the service
// uses stream for sending payload/result.
func needStreaming(data []*ServiceData) bool {
	for _, s := range data {
		if streamingEndpointExists(s) {
			return true
		}
	}
	return false
}

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "ServerName": string}
const mainCLIT = `func main() {
	var (
		addr    = flag.String("url", "{{ .DefaultURL }}", "` + "`" + `URL` + "`" + ` to service host")
{{ range .Variables }}
		{{ .VarName }} = flag.String("{{ .Name }}", {{ printf "%q" .DefaultValue }}, {{ printf "%q" .Description }})
{{- end }}
		verbose = flag.Bool("verbose", false, "Print request and response details")
		v       = flag.Bool("v", false, "Print request and response details")
		timeout = flag.Int("timeout", 30, "Maximum number of ` + "`" + `seconds` + "`" + ` to wait for response")
	)
	flag.Usage = usage
	flag.Parse()
{{ if .Variables }}

	{{ range .Variables }}
	addr = strings.Replace(addr, {{ printf "\"{%s}\"" .Name }}, {{ .VarName }}, -1)
	{{- end }}
{{- end }}

	var (
		scheme string
		host   string
		debug  bool
	)
	{
		u, err := url.Parse(*addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %#v: %s", *addr, err)
			os.Exit(1)
		}
		scheme = u.Scheme
		host = u.Host
		debug = *verbose || *v
	}

	var (
		doer goahttp.Doer
	)
	{
		doer = &http.Client{Timeout: time.Duration(*timeout) * time.Second}
		if debug {
			doer = goahttp.NewDebugDoer(doer)
		}
	}

	{{ if needStreaming .Services }}
	var (
    dialer *websocket.Dialer
		connConfigFn goahttp.ConnConfigureFunc
  )
  {
    dialer = websocket.DefaultDialer
  }
	{{ end }}

	endpoint, payload, err := cli.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
		{{- if needStreaming .Services }}
		dialer,
		connConfigFn,
		{{- end }}
		{{- range .Services }}
			{{- range .Endpoints }}
			  {{- if .MultipartRequestDecoder }}
		{{ $.APIPkg }}.{{ .MultipartRequestEncoder.FuncName }},
				{{- end }}
			{{- end }}
		{{- end }}
	)
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}

	data, err := endpoint(context.Background(), payload)

	if debug {
		doer.(goahttp.DebugDoer).Fprint(os.Stderr)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if data != nil && !debug {
		m, _ := json.MarshalIndent(data, "", "    ")
		fmt.Println(string(m))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, ` + "`" + `%s is a command line client for the {{ .ServerName }} server.

Usage:
    %s [-url URL][-timeout SECONDS][-verbose|-v] SERVICE ENDPOINT [flags]

    -url URL:    specify service URL ({{ .DefaultURL }})
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)

Commands:
%s
Additional help:
    %s SERVICE [ENDPOINT] --help

Example:
%s
` + "`" + `, os.Args[0], os.Args[0], indent(cli.UsageCommands()), os.Args[0], indent(cli.UsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
`
