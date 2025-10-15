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
		{Path: "sort"},
		{Path: "slices"},
		{Path: "strings"},
		codegen.GoaImport(""),
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header("", "main", specs),
		{
			Name:   "cli-main-start",
			Source: exampleTemplates.Read(clientStartT),
			Data: map[string]any{
				"Server":     svrdata,
				"HasJSONRPC": hasJSONRPC(root, svr),
				"HasHTTP":    hasHTTP(root, svr),
			},
			FuncMap: map[string]any{
				"join": strings.Join,
			},
		}, {
			Name:   "cli-main-var-init",
			Source: exampleTemplates.Read(clientVarInitT),
			Data: map[string]any{
				"Server": svrdata,
			},
			FuncMap: map[string]any{
				"join": strings.Join,
			},
		}, {
			Name:   "cli-main-endpoint-init",
			Source: exampleTemplates.Read(clientEndpointInitT),
			Data: map[string]any{
				"Server":     svrdata,
				"Root":       root,
				"HasJSONRPC": hasJSONRPC(root, svr),
				"HasHTTP":    hasHTTP(root, svr),
			},
			FuncMap: map[string]any{
				"join":    strings.Join,
				"toUpper": strings.ToUpper,
			},
		}, {
			Name:   "cli-main-end",
			Source: exampleTemplates.Read(clientEndT),
		}, {
			Name:   "cli-main-usage",
			Source: exampleTemplates.Read(clientUsageT),
			Data: map[string]any{
				"APIName":    root.API.Name,
				"Server":     svrdata,
				"HasJSONRPC": hasJSONRPC(root, svr),
				"HasHTTP":    hasHTTP(root, svr),
			},
			FuncMap: map[string]any{
				"toUpper": strings.ToUpper,
				"join":    strings.Join,
			},
		},
	}
	return &codegen.File{Path: path, SectionTemplates: sections, SkipExist: true}
}

// hasJSONRPC returns true if the server expression has a JSON-RPC server.
func hasJSONRPC(root *expr.RootExpr, svr *expr.ServerExpr) bool {
	for _, s := range svr.Services {
		if root.API.JSONRPC.Service(s) != nil {
			return true
		}
	}
	return false
}

// hasHTTP returns true if the server expression has an HTTP server.
func hasHTTP(root *expr.RootExpr, svr *expr.ServerExpr) bool {
	for _, s := range svr.Services {
		if root.API.HTTP.Service(s) != nil {
			return true
		}
	}
	return false
}
