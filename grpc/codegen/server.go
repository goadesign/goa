package codegen

import (
	"fmt"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// ServerFiles returns all the server files for every gRPC service. The files
// contain the server which implements the generated gRPC server interface and
// encoders and decoders to transform protocol buffer types and gRPC metadata
// into goa types and vice versa.
func ServerFiles(genpkg string, services *ServicesData) []*codegen.File {
	svcLen := len(services.Root.API.GRPC.Services)
	fw := make([]*codegen.File, 2*svcLen)
	for i, svc := range services.Root.API.GRPC.Services {
		fw[i] = serverFile(genpkg, svc, services)
	}
	for i, svc := range services.Root.API.GRPC.Services {
		fw[i+svcLen] = serverEncodeDecode(genpkg, svc, services)
	}
	return fw
}

// serverFile returns the files defining the gRPC server.
func serverFile(genpkg string, svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = services.Get(svc.Name())
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
		sections = []*codegen.SectionTemplate{
			codegen.Header(svc.Name()+" gRPC server", "server", imports),
			{
				Name:   "server-struct",
				Source: grpcTemplates.Read(grpcServerStructTypeT),
				Data:   data,
			},
		}
		for _, e := range data.Endpoints {
			if e.ServerStream != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-stream-struct-type",
					Source: grpcTemplates.Read(grpcStreamStructTypeT),
					Data:   e.ServerStream,
				})
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-init",
			Source: grpcTemplates.Read(grpcServerInitT),
			Data:   data,
		})
		for _, e := range data.Endpoints {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "grpc-handler-init",
				Source: grpcTemplates.Read(grpcHandlerInitT),
				Data:   e,
			}, &codegen.SectionTemplate{
				Name:   "server-grpc-interface",
				Source: grpcTemplates.Read(grpcServerGRPCInterfaceT),
				Data:   e,
			})
		}
		for _, e := range data.Endpoints {
			if e.ServerStream != nil {
				if e.ServerStream.SendConvert != nil {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-send",
						Source: grpcTemplates.Read(grpcStreamSendT),
						Data:   e.ServerStream,
					})
				}
				if e.Method.StreamKind == expr.ClientStreamKind || e.Method.StreamKind == expr.BidirectionalStreamKind {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-recv",
						Source: grpcTemplates.Read(grpcStreamRecvT),
						Data:   e.ServerStream,
					})
				}
				if e.ServerStream.MustClose {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-close",
						Source: grpcTemplates.Read(grpcStreamCloseT),
						Data:   e.ServerStream,
					})
				}
				if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-set-view",
						Source: grpcTemplates.Read(grpcStreamSetViewT),
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
func serverEncodeDecode(genpkg string, svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = services.Get(svc.Name())
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
		sections = []*codegen.SectionTemplate{codegen.Header(title, "server", imports)}

		for _, e := range data.Endpoints {
			if e.Response.ServerConvert != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "response-encoder",
					Source: grpcTemplates.Read(grpcResponseEncoderT, grpcConvertTypeToStringP, "string_conversion"),
					Data:   e,
					FuncMap: map[string]any{
						"typeConversionData":       typeConversionData,
						"metadataEncodeDecodeData": metadataEncodeDecodeData,
					},
				})
			}
			if e.PayloadRef != "" {
				fm := transTmplFuncs(svc, services)
				fm["isEmpty"] = isEmpty
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "request-decoder",
					Source:  grpcTemplates.Read(grpcRequestDecoderT, grpcConvertStringToTypeP, "type_conversion", "slice_conversion", "slice_item_conversion"),
					Data:    e,
					FuncMap: fm,
				})
			}
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

func transTmplFuncs(s *expr.GRPCServiceExpr, services *ServicesData) map[string]any {
	return map[string]any{
		"goTypeRef": func(dt expr.DataType) string {
			return services.ServicesData.Get(s.Name()).Scope.GoTypeRef(&expr.AttributeExpr{Type: dt})
		},
	}
}

// typeConversionData produces the template data suitable for executing the
// "type_conversion" template.
func typeConversionData(dt expr.DataType, varName, target string) map[string]any {
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
