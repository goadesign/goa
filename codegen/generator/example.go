package generator

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/eval"
	httpcodegen "goa.design/goa.v2/http/codegen"
	httpdesign "goa.design/goa.v2/http/design"
)

// Example iterates through the roots and returns files that implement an
// example service and client.
func Example(roots []eval.Root) ([]codegen.File, error) {
	var files []codegen.File
	for _, root := range roots {
		if r, ok := root.(*httpdesign.RootExpr); ok {
			files = httpcodegen.ExampleServerFiles(r)
			files = append(files, httpcodegen.ExampleCLI(r))
			break
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("example: no HTTP design found")
	}
	return files, nil
}
