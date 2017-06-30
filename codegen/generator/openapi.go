package generator

import (
	"fmt"

	"goa.design/goa.v2/codegen"
	restfiles "goa.design/goa.v2/codegen/files/rest"
	"goa.design/goa.v2/design/rest"
	"goa.design/goa.v2/eval"
)

// OpenAPI iterates through the roots and returns the file needed to render
// the service OpenAPI spec. It returns an error if the roots slice does not
// include a rest root.
func OpenAPI(roots ...eval.Root) ([]codegen.File, error) {
	var (
		file codegen.File
		err  error
	)
	for _, root := range roots {
		switch r := root.(type) {
		case *rest.RootExpr:
			file, err = restfiles.OpenAPI(r)
		}
	}
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("could not find rest design in DSL roots")
	}
	return []codegen.File{file}, nil
}
