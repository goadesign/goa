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
func exampleCLIMain(_ string, root *expr.RootExpr, svr *expr.ServerExpr) *codegen.File {
	svrdata := Servers.Get(svr, root)

	path := filepath.Join("cmd", svrdata.Dir+"-cli", "main.go")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	specs := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "encoding/json"},
		{Path: "errors"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "net/url"},
		{Path: "os"},
		{Path: "strings"},
		codegen.GoaImport(""),
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "cli-main-start",
			Source: readTemplate("client_start"),
			Data: map[string]any{
				"Server": svrdata,
			},
			FuncMap: map[string]any{
				"join": strings.Join,
			},
		}, {
			Name:   "cli-main-var-init",
			Source: readTemplate("client_var_init"),
			Data: map[string]any{
				"Server": svrdata,
			},
			FuncMap: map[string]any{
				"join": strings.Join,
			},
		}, {
			Name:   "cli-main-endpoint-init",
			Source: readTemplate("client_endpoint_init"),
			Data: map[string]any{
				"Server": svrdata,
			},
			FuncMap: map[string]any{
				"join":    strings.Join,
				"toUpper": strings.ToUpper,
			},
		}, {
			Name:   "cli-main-end",
			Source: readTemplate("client_end"),
		}, {
			Name:   "cli-main-usage",
			Source: readTemplate("client_usage"),
			Data: map[string]any{
				"APIName": root.API.Name,
				"Server":  svrdata,
			},
			FuncMap: map[string]any{
				"toUpper": strings.ToUpper,
				"join":    strings.Join,
			},
		},
	}
	return &codegen.File{Path: path, SectionTemplates: sections, SkipExist: true}
}
