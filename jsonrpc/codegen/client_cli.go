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
		// Fix JSON-RPC specific template sections
		for _, section := range f.SectionTemplates {
			if section.Name == "parse-endpoint" {
				// Update the template source to use goahttp.ConnConfigureFunc instead of *ConnConfigurer
				section.Source = strings.ReplaceAll(section.Source,
					"{{ .VarName }}Configurer *{{ .PkgName }}.ConnConfigurer,",
					"{{ .VarName }}ConfigFn goahttp.ConnConfigureFunc,")
				section.Source = strings.ReplaceAll(section.Source,
					", {{ .VarName }}Configurer{{ end }}",
					", {{ .VarName }}ConfigFn{{ end }}")
			}
		}
	}
	return res
}
