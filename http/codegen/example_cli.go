package codegen

import (
	"os"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

// ExampleCLIFiles returns an example client tool HTTP implementation for each
// server expression.
func ExampleCLIFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var files []*codegen.File
	for _, svr := range root.API.Servers {
		if f := exampleCLI(genpkg, root, svr); f != nil {
			files = append(files, f)
		}
	}
	return files
}

// exampleCLI returns an example client tool HTTP implementation for the given
// server expression.
func exampleCLI(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr) *codegen.File {
	svrdata := example.Servers.Get(svr)
	path := filepath.Join("cmd", svrdata.Dir+"-cli", "http.go")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	var (
		rootPath string
		apiPkg   string

		scope = codegen.NewNameScope()
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
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/http/cli/" + svrdata.Dir, Name: "cli"},
		{Path: rootPath, Name: apiPkg},
	}

	var svcData []*ServiceData
	for _, svc := range svr.Services {
		if data := HTTPServices.Get(svc); data != nil {
			svcData = append(svcData, data)
		}
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "cli-http-start",
			Source: httpCLIStartT,
		},
		{
			Name:   "cli-http-streaming",
			Source: httpCLIStreamingT,
			Data: map[string]interface{}{
				"Services": svcData,
			},
			FuncMap: map[string]interface{}{
				"needStream": needStream,
			},
		},
		{
			Name:   "cli-http-end",
			Source: httpCLIEndT,
			Data: map[string]interface{}{
				"Services": svcData,
				"APIPkg":   apiPkg,
			},
			FuncMap: map[string]interface{}{
				"needStream":   needStream,
				"hasWebSocket": hasWebSocket,
			},
		},
		{
			Name:   "cli-http-usage",
			Source: httpCLIUsageT,
		},
	}
	return &codegen.File{
		Path:             path,
		SectionTemplates: sections,
		SkipExist:        true,
	}
}

const (
	httpCLIStartT = `func doHTTP(scheme, host string, timeout int, debug bool) (goa.Endpoint, interface{}, error) {
	var (
		doer goahttp.Doer
	)
	{
		doer = &http.Client{Timeout: time.Duration(timeout) * time.Second}
		if debug {
			doer = goahttp.NewDebugDoer(doer)
		}
	}
`

	// input: map[string]interface{}{"Services": []*ServiceData}
	httpCLIStreamingT = `{{- if needStream .Services }}
	var (
    dialer *websocket.Dialer
  )
  {
    dialer = websocket.DefaultDialer
  }
	{{ end }}
`

	// input: map[string]interface{}{"Services": []*ServiceData}
	httpCLIEndT = `return cli.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
		{{- if needStream .Services }}
		dialer,
			{{- range $svc := .Services }}
				{{- if hasWebSocket $svc }}
				nil,
				{{- end }}
			{{- end }}
		{{- end }}
		{{- range .Services }}
			{{- range .Endpoints }}
			  {{- if .MultipartRequestDecoder }}
		{{ $.APIPkg }}.{{ .MultipartRequestEncoder.FuncName }},
				{{- end }}
			{{- end }}
		{{- end }}
	)
}
`

	httpCLIUsageT = `
func httpUsageCommands() string {
  return cli.UsageCommands()
}

func httpUsageExamples() string {
  return cli.UsageExamples()
}
`
)
