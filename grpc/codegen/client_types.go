package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// ClientTypeFiles returns the client types files containing all the client
// interfaces and types needed to implement gRPC client.
func ClientTypeFiles(genpkg string, services *ServicesData) []*codegen.File {
	fw := make([]*codegen.File, len(services.Root.API.GRPC.Services))
	for i, svc := range services.Root.API.GRPC.Services {
		fw[i] = clientType(genpkg, svc, services)
	}
	return fw
}

// clientType returns the file defining the gRPC client types.
func clientType(genpkg string, svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		initData []*InitData

		sd = services.Get(svc.Name())
	)
	{
		seen := make(map[string]struct{})
		collect := func(c *ConvertData) {
			if c.Init == nil {
				return
			}
			if _, ok := seen[c.Init.Name]; ok {
				return
			}
			seen[c.Init.Name] = struct{}{}
			initData = append(initData, c.Init)
		}
		for _, a := range svc.GRPCEndpoints {
			ed := sd.Endpoint(a.Name())
			if c := ed.Request.ClientConvert; c != nil {
				collect(c)
			}
			if c := ed.Response.ClientConvert; c != nil {
				collect(c)
			}
			if ed.ClientStream != nil {
				if c := ed.ClientStream.RecvConvert; c != nil {
					collect(c)
				}
				if c := ed.ClientStream.SendConvert; c != nil {
					collect(c)
				}
			}
			for _, e := range ed.Errors {
				if c := e.Response.ClientConvert; c != nil {
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
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "client", "types.go")
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
		sections = []*codegen.SectionTemplate{codegen.Header(svc.Name()+" gRPC client types", "client", imports)}
		for _, init := range initData {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-type-init",
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
		}
		for _, data := range sd.validations {
			if data.Kind == validateServer {
				continue
			}
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-validate",
				Source: grpcTemplates.Read(grpcValidateT),
				Data:   data,
			})
		}
		for _, h := range sd.transformHelpers {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-transform-helper",
				Source: grpcTemplates.Read(grpcTransformHelperT),
				Data:   h,
			})
		}
	}

	return &codegen.File{Path: fpath, SectionTemplates: sections}
}
