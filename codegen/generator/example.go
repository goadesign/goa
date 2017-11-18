package generator

import (
	"goa.design/goa/codegen"
	"goa.design/goa/eval"
	httpcodegen "goa.design/goa/http/codegen"
	httpdesign "goa.design/goa/http/design"
)

// Example iterates through the roots and returns files that implement an
// example service and client.
func Example(genpkg string, roots []eval.Root) ([]*codegen.File, error) {
	var files []*codegen.File
	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			files = httpcodegen.ExampleServerFiles(genpkg, r)
			if cli := httpcodegen.ExampleCLI(genpkg, r); cli != nil {
				files = append(files, cli)
			}
			break
		}
	}
	return files, nil
}
