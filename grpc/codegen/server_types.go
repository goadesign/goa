package codegen

import (
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
)

// ServerTypeFiles returns the gRPC transport type files.
func ServerTypeFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, len(root.API.GRPC.Services))
	seen := make(map[string]struct{})
	for i, r := range root.API.GRPC.Services {
		fw[i] = serverType(genpkg, r, seen)
	}
	return fw
}

// serverType returns the file containing the constructor functions to
// transform the gRPC request types to the corresponding service payload types
// and service result types to the corresponding gRPC response types.
//
// seen keeps track of the constructor names that have already been generated
// to prevent duplicate code generation.
func serverType(genpkg string, svc *expr.GRPCServiceExpr, seen map[string]struct{}) *codegen.File {
	var (
		path      string
		initData  []*InitData
		validated []*ValidationData

		sd = GRPCServices.Get(svc.Name())
	)
	{
		collect := func(c *ConvertData) {
			if c.Init != nil {
				initData = append(initData, c.Init)
			}
		}

		path = filepath.Join(codegen.Gendir, "grpc", codegen.SnakeCase(svc.Name()), "server", "types.go")
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
		}

		for _, v := range sd.Validations {
			validated = append(validated, v)
		}
	}

	header := codegen.Header(svc.Name()+" gRPC server types", "server",
		[]*codegen.ImportSpec{
			{Path: "unicode/utf8"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: filepath.Join(genpkg, codegen.SnakeCase(svc.Name())), Name: sd.Service.PkgName},
			{Path: filepath.Join(genpkg, codegen.SnakeCase(svc.Name()), "views"), Name: sd.Service.ViewsPkg},
			{Path: filepath.Join(genpkg, "grpc", codegen.SnakeCase(svc.Name()), pbPkgName)},
		},
	)
	sections := []*codegen.SectionTemplate{header}
	for _, init := range initData {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-type-init",
			Source: typeInitT,
			Data:   init,
		})
	}
	for _, data := range validated {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-validate",
			Source: validateT,
			Data:   data,
		})
	}
	for _, h := range sd.TransformHelpers {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-transform-helper",
			Source: transformHelperT,
			Data:   h,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// input: TransformFunctionData
const transformHelperT = `{{ printf "%s builds a value of type %s from a value of type %s." .Name .ResultTypeRef .ParamTypeRef | comment }}
func {{ .Name }}(v {{ .ParamTypeRef }}) {{ .ResultTypeRef }} {
  {{ .Code }}
  return res
}
`
