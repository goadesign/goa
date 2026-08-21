// This file declares gRPC runtime imports and generated protobuf package
// aliases before the shared generation catalog is frozen.
package codegen

import (
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// Plan reserves every literal gRPC import qualifier and each generated
// protobuf package alias used by gRPC render templates.
func Plan(generation *codegen.Generation) error {
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("context"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("errors"),
		codegen.SimpleImport("flag"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("net"),
		codegen.SimpleImport("net/url"),
		codegen.SimpleImport("os"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("sync"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.SimpleImport("goa.design/clue/debug"),
		codegen.SimpleImport("goa.design/clue/log"),
		codegen.GoaImport(""),
		codegen.GoaNamedImport("grpc", "goagrpc"),
		codegen.GoaNamedImport("grpc/pb", "goapb"),
		codegen.SimpleImport("google.golang.org/grpc"),
		codegen.SimpleImport("google.golang.org/grpc/codes"),
		codegen.SimpleImport("google.golang.org/grpc/credentials/insecure"),
		codegen.SimpleImport("google.golang.org/grpc/metadata"),
		codegen.SimpleImport("google.golang.org/grpc/reflection"),
		codegen.SimpleImport("google.golang.org/protobuf/types/known/structpb"),
	}
	for _, spec := range imports {
		if err := generation.RequireImport(spec); err != nil {
			return err
		}
	}
	for _, root := range generation.Roots() {
		design, ok := root.(*expr.RootExpr)
		if !ok {
			continue
		}
		for _, service := range design.API.GRPC.Services {
			pathName := codegen.SnakeCase(codegen.Goify(service.Name(), false))
			packageName := strings.ToLower(codegen.Goify(service.Name(), false))
			if err := generation.ReserveGeneratedImport(codegen.NewImport(packageName+"c", path.Join(generation.GenPkg(), "grpc", pathName, "client"))); err != nil {
				return err
			}
			if err := generation.ReserveGeneratedImport(codegen.NewImport(packageName+"svr", path.Join(generation.GenPkg(), "grpc", pathName, "server"))); err != nil {
				return err
			}
			if err := generation.ReserveGeneratedImport(codegen.NewImport(pathName+"pb", path.Join(generation.GenPkg(), "grpc", pathName, pbPkgName))); err != nil {
				return err
			}
		}
		if len(design.API.GRPC.Services) > 0 {
			for _, server := range design.API.Servers {
				serverName := codegen.SnakeCase(codegen.Goify(server.Name, true))
				if err := generation.ReserveGeneratedImport(codegen.NewImport("cli", path.Join(generation.GenPkg(), "grpc", "cli", serverName))); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
