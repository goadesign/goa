package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// ClientTypeFiles returns the types file for every gRPC service that contain
// constructors to transform:
//
//   * service payload types into protocol buffer request message types
//   * protocol buffer response message types into service result types
func ClientTypeFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, len(root.API.GRPC.Services))
	seen := make(map[string]struct{})
	for i, r := range root.API.GRPC.Services {
		fw[i] = clientType(genpkg, r, seen)
	}
	return fw
}

// clientType returns the file containing the constructor functions to
// transform the service payload types to the corresponding gRPC request types
// and gRPC response types to the corresponding service result types.
//
// seen keeps track of the constructor names that have already been generated
// to prevent duplicate code generation.
func clientType(genpkg string, svc *expr.GRPCServiceExpr, seen map[string]struct{}) *codegen.File {
	var (
		initData []*InitData

		sd = GRPCServices.Get(svc.Name())
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
		imports = append(imports, sd.Service.UserTypeImports...)
		imports = append(imports, sd.Service.ProtoImports...)
		sections = []*codegen.SectionTemplate{codegen.Header(svc.Name()+" gRPC client types", "client", imports)}
		for _, init := range initData {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-type-init",
				Source: typeInitT,
				Data:   init,
				FuncMap: map[string]interface{}{
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
				Source: validateT,
				Data:   data,
			})
		}
		for _, h := range sd.transformHelpers {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-transform-helper",
				Source: transformHelperT,
				Data:   h,
			})
		}
	}

	return &codegen.File{Path: fpath, SectionTemplates: sections}
}
