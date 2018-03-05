package codegen

import (
	"os"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	httpdesign "goa.design/goa/http/design"
)

// ExampleCLI returns an example client tool main implementation.
func ExampleCLI(genpkg string, root *httpdesign.RootExpr) *codegen.File {
	path := filepath.Join("cmd", codegen.SnakeCase(root.Design.API.Name)+"cli", "main.go")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
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
		{Path: "encoding/json"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "net/http"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "strings"},
		{Path: "time"},
		{Path: "goa.design/goa/http", Name: "goahttp"},
		{Path: rootPath, Name: apiPkg},
		{Path: genpkg + "/http/cli"},
	}
	svcdata := make([]*ServiceData, 0, len(root.HTTPServices))
	for _, svc := range root.HTTPServices {
		svcdata = append(svcdata, HTTPServices.Get(svc.Name()))
	}
	data := map[string]interface{}{
		"Services": svcdata,
		"APIPkg":   apiPkg,
		"APIName":  root.Design.API.Name,
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		&codegen.SectionTemplate{
			Name:   "cli-main",
			Source: mainCLIT,
			Data:   data,
		},
	}
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// input: map[string]interface{}{"Services":[]ServiceData, "APIPkg": string, "APIName": string}
const mainCLIT = `func main() {
	var (
		addr    = flag.String("url", "http://localhost:8080", "` + "`" + `URL` + "`" + ` to service host")
		verbose = flag.Bool("verbose", false, "Print request and response details")
		v       = flag.Bool("v", false, "Print request and response details")
		timeout = flag.Int("timeout", 30, "Maximum number of ` + "`" + `seconds` + "`" + ` to wait for response")
	)
	flag.Usage = usage
	flag.Parse()

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
		if scheme == "" {
			scheme = "http"
		}
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

	endpoint, payload, err := cli.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
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
	fmt.Fprintf(os.Stderr, ` + "`" + `%s is a command line client for the {{ .APIName }} API.

Usage:
    %s [-url URL][-timeout SECONDS][-verbose|-v] SERVICE ENDPOINT [flags]

    -url URL:    specify service URL (http://localhost:8080)
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
