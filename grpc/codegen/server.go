package codegen

import (
	"fmt"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// ServerFiles returns all the server files for every gRPC service. The files
// contain the server which implements the generated gRPC server interface and
// encoders and decoders to transform protocol buffer types and gRPC metadata
// into goa types and vice versa.
func ServerFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	svcLen := len(root.API.GRPC.Services)
	fw := make([]*codegen.File, 2*svcLen)
	for i, svc := range root.API.GRPC.Services {
		fw[i] = serverFile(genpkg, svc)
	}
	for i, svc := range root.API.GRPC.Services {
		fw[i+svcLen] = serverEncodeDecode(genpkg, svc)
	}
	return fw
}

// serverFile returns the files defining the gRPC server.
func serverFile(genpkg string, svc *expr.GRPCServiceExpr) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = GRPCServices.Get(svc.Name())
	)
	{
		svcName := data.Service.PathName
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "server", "server.go")
		imports := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "errors"},
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			{Path: "google.golang.org/grpc/codes"},
			{Path: path.Join(genpkg, svcName), Name: data.Service.PkgName},
			{Path: path.Join(genpkg, svcName, "views"), Name: data.Service.ViewsPkg},
			{Path: path.Join(genpkg, "grpc", svcName, pbPkgName), Name: data.PkgName},
		}
		imports = append(imports, data.Service.UserTypeImports...)
		sections = []*codegen.SectionTemplate{
			codegen.Header(svc.Name()+" gRPC server", "server", imports),
			{
				Name:   "server-struct",
				Source: readTemplate("server_struct_type"),
				Data:   data,
			},
		}
		for _, e := range data.Endpoints {
			if e.ServerStream != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-stream-struct-type",
					Source: readTemplate("stream_struct_type"),
					Data:   e.ServerStream,
				})
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-init",
			Source: readTemplate("server_init"),
			Data:   data,
		})
		for _, e := range data.Endpoints {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "grpc-handler-init",
				Source: readTemplate("grpc_handler_init"),
				Data:   e,
			})
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-grpc-interface",
				Source: readTemplate("server_grpc_interface"),
				Data:   e,
			})
		}
		for _, e := range data.Endpoints {
			if e.ServerStream != nil {
				if e.ServerStream.SendConvert != nil {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-send",
						Source: readTemplate("stream_send"),
						Data:   e.ServerStream,
					})
				}
				if e.Method.StreamKind == expr.ClientStreamKind || e.Method.StreamKind == expr.BidirectionalStreamKind {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-recv",
						Source: readTemplate("stream_recv"),
						Data:   e.ServerStream,
					})
				}
				if e.ServerStream.MustClose {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-close",
						Source: readTemplate("stream_close"),
						Data:   e.ServerStream,
					})
				}
				if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-set-view",
						Source: readTemplate("stream_set_view"),
						Data:   e.ServerStream,
					})
				}
			}
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

// serverEncodeDecode returns the file defining the gRPC server encoding and
// decoding logic.
func serverEncodeDecode(genpkg string, svc *expr.GRPCServiceExpr) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = GRPCServices.Get(svc.Name())
	)
	{
		svcName := data.Service.PathName
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "server", "encode_decode.go")
		title := fmt.Sprintf("%s gRPC server encoders and decoders", svc.Name())
		imports := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "strings"},
			{Path: "strconv"},
			{Path: "unicode/utf8"},
			{Path: "google.golang.org/grpc"},
			{Path: "google.golang.org/grpc/metadata"},
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			{Path: path.Join(genpkg, svcName), Name: data.Service.PkgName},
			{Path: path.Join(genpkg, svcName, "views"), Name: data.Service.ViewsPkg},
			{Path: path.Join(genpkg, "grpc", svcName, pbPkgName), Name: data.PkgName},
		}
		imports = append(imports, data.Service.UserTypeImports...)
		sections = []*codegen.SectionTemplate{codegen.Header(title, "server", imports)}

		for _, e := range data.Endpoints {
			if e.Response.ServerConvert != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "response-encoder",
					Source: readTemplate("response_encoder", "convert_type_to_string"),
					Data:   e,
					FuncMap: map[string]any{
						"typeConversionData":       typeConversionData,
						"metadataEncodeDecodeData": metadataEncodeDecodeData,
					},
				})
			}
			if e.PayloadRef != "" {
				fm := transTmplFuncs(svc)
				fm["isEmpty"] = isEmpty
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "request-decoder",
					Source:  readTemplate("request_decoder", "convert_string_to_type"),
					Data:    e,
					FuncMap: fm,
				})
			}
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

func transTmplFuncs(s *expr.GRPCServiceExpr) map[string]any {
	return map[string]any{
		"goTypeRef": func(dt expr.DataType) string {
			return service.Services.Get(s.Name()).Scope.GoTypeRef(&expr.AttributeExpr{Type: dt})
		},
	}
}

// typeConversionData produces the template data suitable for executing the
// "type_conversion" template.
func typeConversionData(dt expr.DataType, varName string, target string) map[string]any {
	return map[string]any{
		"Type":    dt,
		"VarName": varName,
		"Target":  target,
	}
}

// metadataEncodeDecodeData produces the template data suitable for executing the
// "metadata_decoder" and "metadata_encoder" template.
func metadataEncodeDecodeData(md *MetadataData, vname string) map[string]any {
	return map[string]any{
		"Metadata": md,
		"VarName":  vname,
	}
}
