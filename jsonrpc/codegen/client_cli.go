package codegen

import (
	"strings"

	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// ClientCLIFiles returns the JSON-RPC transport type files.
func ClientCLIFiles(genpkg string, services *httpcodegen.ServicesData) []*codegen.File {
	res := httpcodegen.ClientCLIFiles(genpkg, services)
	for _, f := range res {
		updateHeader(f)
		f.Path = strings.Replace(f.Path, "/http/", "/jsonrpc/", 1)
	}
	return res
}
