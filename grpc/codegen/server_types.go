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
		imports = append(imports, sd.Service.ProtoImports...)
		sections = []*codegen.SectionTemplate{codegen.Header(svc.Name()+" gRPC server types", "server", imports)}
		for _, init := range initData {
			if _, ok := foundInits[init.Name]; ok {
				continue
			}
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-type-init",
				Source: readTemplate("type_init"),
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
				Source: readTemplate("validate"),
				Data:   data,
			})
		}
		for _, h := range sd.transformHelpers {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-transform-helper",
				Source: readTemplate("transform_helper"),
				Data:   h,
			})
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}
