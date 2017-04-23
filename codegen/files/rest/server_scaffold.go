package rest

import (
	"path/filepath"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design/rest"
)

// ServerScaffoldFiles returns all the server HTTP scaffold files.
func ServerScaffoldFiles(root *rest.RootExpr) []codegen.File {
	path := filepath.Join("cmd", root.Design.API.Name, "main.go")
	sections := func(genPkg string) []*codegen.Section {
		return nil
	}
	return []codegen.File{codegen.NewSource(path, sections)}
}
