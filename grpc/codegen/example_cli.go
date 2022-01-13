package codegen

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/expr"
)

// ExampleCLIFiles returns an example gRPC client tool implementation.
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
	var (
		mainPath string

		svrdata = example.Servers.Get(svr)
	)
	{
		mainPath = filepath.Join("cmd", svrdata.Dir+"-cli", "grpc.go")
		if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
			return nil // file already exists, skip it.
		}
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

	var (
		specs []*codegen.ImportSpec
	)
	{
		specs = []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "encoding/json"},
			{Path: "flag"},
			{Path: "fmt"},
			{Path: "google.golang.org/grpc"},
			{Path: "google.golang.org/grpc/credentials/insecure"},
			{Path: "os"},
			{Path: "time"},
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			{Path: rootPath, Name: apiPkg},
			{Path: path.Join(genpkg, "grpc", "cli", svrdata.Dir), Name: "cli"},
		}
	}

	var (
		sections []*codegen.SectionTemplate
	)
	{
		sections = []*codegen.SectionTemplate{
			codegen.Header("", "main", specs),
			{Name: "do-grpc-cli", Source: grpcCLIDoT, Data: svrdata},
		}
	}

	return &codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}
}

const (
	grpcCLIDoT = `func doGRPC(scheme, host string, timeout int, debug bool) (goa.Endpoint, interface{}, error) {
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
    fmt.Fprintf(os.Stderr, "could not connect to gRPC server at %s: %v\n", host, err)
  }
	return cli.ParseEndpoint(conn)
}

{{ if eq .DefaultTransport.Type "grpc" }}
func grpcUsageCommands() string {
	return cli.UsageCommands()
}

func grpcUsageExamples() string {
	return cli.UsageExamples()
}
{{- end }}
`
)
