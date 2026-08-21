// This file declares application packages imported by generated examples so
// their qualifiers are selected with the same catalog as generated services.
package example

import (
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// Plan reserves the application and interceptor package aliases consumed by
// example server and client files before the generation catalog freezes.
func Plan(generation *codegen.Generation) error {
	rootPath := path.Dir(generation.GenPkg())
	for _, root := range generation.Roots() {
		design, ok := root.(*expr.RootExpr)
		if !ok {
			continue
		}
		scope := codegen.NewNameScope()
		for _, service := range design.Services {
			scope.Unique(strings.ToLower(codegen.Goify(service.Name, false)))
		}
		packageName := scope.Unique(strings.ToLower(codegen.Goify(design.API.Name, false)), "api")
		if err := generation.DeclareImport(codegen.NewImport(packageName, rootPath)); err != nil {
			return err
		}
		if hasInterceptors(design) {
			if err := generation.DeclareImport(codegen.NewImport("interceptors", rootPath+"/interceptors")); err != nil {
				return err
			}
		}
	}
	return nil
}

// hasInterceptors reports whether generated examples import the application
// interceptor package for at least one service.
func hasInterceptors(root *expr.RootExpr) bool {
	for _, service := range root.Services {
		if len(service.ServerInterceptors) > 0 || len(service.ClientInterceptors) > 0 {
			return true
		}
	}
	return false
}
