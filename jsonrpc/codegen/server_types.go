package codegen

import (
	"strings"

	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// ServerTypeFiles returns the JSON-RPC transport type files.
func ServerTypeFiles(genpkg string, services *httpcodegen.ServicesData) []*codegen.File {
	res := httpcodegen.ServerTypeFiles(genpkg, services)
	for _, f := range res {
		updateHeader(f)
		f.Path = strings.Replace(f.Path, "/http/", "/jsonrpc/", 1)
	}
	return res
}
