package codegen

import (
	"os"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
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
		apiPkg   string
		pkg      string
	)
	{
		apiPkg = strings.ToLower(codegen.Goify(root.API.Name, false))
		pkg = codegen.SnakeCase(codegen.Goify(svr.Name, true))
		mainPath = filepath.Join("cmd", pkg+"-cli", "grpc.go")
		if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
			return nil // file already exists, skip it.
		}
	}

	var (
		specs []*codegen.ImportSpec
	)
	{
		idx := strings.LastIndex(genpkg, string(os.PathSeparator))
		rootPath := "."
		if idx > 0 {
			rootPath = genpkg[:idx]
		}
		specs = []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "encoding/json"},
			{Path: "flag"},
			{Path: "fmt"},
			{Path: "google.golang.org/grpc"},
			{Path: "os"},
			{Path: "time"},
			{Path: "goa.design/goa"},
			{Path: "goa.design/goa/grpc", Name: "goagrpc"},
			{Path: rootPath, Name: apiPkg},
			{Path: filepath.Join(genpkg, "grpc", "cli", pkg), Name: "cli"},
		}
	}

	var (
		sections []*codegen.SectionTemplate
	)
	{
		sections = []*codegen.SectionTemplate{
			codegen.Header("", "main", specs),
			&codegen.SectionTemplate{Name: "do-grpc-cli", Source: grpcCLIDoT},
		}
	}

	return &codegen.File{Path: mainPath, SectionTemplates: sections, SkipExist: true}
}

const (
	grpcCLIDoT = `func doGRPC(scheme, host string, timeout int, debug bool) (goa.Endpoint, interface{}, error) {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
    fmt.Fprintln(os.Stderr, fmt.Sprintf("could not connect to gRPC server at %s: %v", host, err))
  }
	return cli.ParseEndpoint(conn)
}

func grpcUsageCommands() string {
	return cli.UsageCommands()
}

func grpcUsageExamples() string {
	return cli.UsageExamples()
}
`
)
