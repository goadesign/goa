// This file adds the imports that the HTTP plan prepared for each JSON-RPC
// output file.
package codegen

import (
	"goa.design/goa/v3/codegen"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// addFileImports adds the service-type imports prepared for file's path and
// returns file so callers can append it directly to their result.
func addFileImports(file *codegen.File, service httpcodegen.JSONRPCServiceSnapshot) *codegen.File {
	codegen.AddImport(file.SectionTemplates[0], service.FileImports(file.Path)...)
	return file
}
