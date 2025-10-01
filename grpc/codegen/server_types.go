package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// ServerTypeFiles returns the server types files containing all the server
// interfaces and types needed to implement gRPC server.
func ServerTypeFiles(genpkg string, services *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(services.Root.API.GRPC.Services))
	for i, svc := range services.Root.API.GRPC.Services {
		fw[i] = serverType(genpkg, svc, services)
	}
	return fw
}

// serverType returns the file defining the gRPC server types.
func serverType(genpkg string, svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		initData []*InitData

		sd         = services.Get(svc.Name())
		foundInits = make(map[string]struct{})
	)
	{
		collect := func(c *ConvertData) {
			if c.Init != nil {
				initData = append(initData, c.Init)
			}
		}
		for _, a := range svc.GRPCEndpoints {
			ed := sd.Endpoint(a.Name())
			if c := ed.Request.ServerConvert; c != nil {
				collect(c)
			}
			if c := ed.Response.ServerConvert; c != nil {
				collect(c)
			}
			if ed.ServerStream != nil {
				if c := ed.ServerStream.SendConvert; c != nil {
					collect(c)
				}
				if c := ed.ServerStream.RecvConvert; c != nil {
					collect(c)
				}
			}
			for _, e := range ed.Errors {
				if c := e.Response.ServerConvert; c != nil {
					collect(c)
				}
			}
		}
	}

	var (
		fpath    string
		sections []*codegen.SectionTemplate
	)
	{
		svcName := sd.Service.PathName
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "server", "types.go")
		imports := []*codegen.ImportSpec{
			{Path: "unicode/utf8"},
			codegen.GoaImport(""),
			{Path: path.Join(genpkg, svcName), Name: sd.Service.PkgName},
			{Path: path.Join(genpkg, svcName, "views"), Name: sd.Service.ViewsPkg},
			{Path: path.Join(genpkg, "grpc", svcName, pbPkgName), Name: sd.PkgName},
		}
		// Add imports if Any type is used
		needsAnyTypeImports := false
		for _, e := range svc.GRPCEndpoints {
			if hasAnyType(e.MethodExpr.Payload) || hasAnyType(e.MethodExpr.Result) {
				needsAnyTypeImports = true
				break
			}
			for _, er := range e.MethodExpr.Errors {
				if hasAnyType(er.AttributeExpr) {
					needsAnyTypeImports = true
					break
				}
			}
			if needsAnyTypeImports {
				break
			}
		}
		if needsAnyTypeImports {
			imports = append(imports, &codegen.ImportSpec{Path: "google.golang.org/protobuf/types/known/anypb", Name: "anypb"})
			imports = append(imports, &codegen.ImportSpec{Path: "encoding/json"})
			imports = append(imports, &codegen.ImportSpec{Path: "google.golang.org/protobuf/types/known/structpb", Name: "structpb"})
		}
		imports = append(imports, sd.Service.ProtoImports...)
		sections = []*codegen.SectionTemplate{codegen.Header(svc.Name()+" gRPC server types", "server", imports)}
		for _, init := range initData {
			if _, ok := foundInits[init.Name]; ok {
				continue
			}
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-type-init",
				Source: grpcTemplates.Read(grpcTypeInitT),
				Data:   init,
				FuncMap: map[string]any{
					"isAlias": expr.IsAlias,
					"fullName": func(dt expr.DataType) string {
						if loc := codegen.UserTypeLocation(dt); loc != nil {
							return loc.PackageName() + "." + dt.Name()
						}
						return dt.Name()
					},
				},
			})
			foundInits[init.Name] = struct{}{}
		}
		for _, data := range sd.validations {
			if data.Kind == validateClient {
				continue
			}
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-validate",
				Source: grpcTemplates.Read(grpcValidateT),
				Data:   data,
			})
		}
		for _, h := range sd.transformHelpers {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-transform-helper",
				Source: grpcTemplates.Read(grpcTransformHelperT),
				Data:   h,
			})
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}
