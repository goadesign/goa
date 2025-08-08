package codegen

import (
	"strings"

	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// PathFiles returns the service path files.
func PathFiles(data *httpcodegen.ServicesData) []*codegen.File {
	res := httpcodegen.PathFiles(data)
	for _, f := range res {
		updateHeader(f)
		f.Path = strings.Replace(f.Path, "/http/", "/jsonrpc/", 1)
	}
	return res
}
