// This file renders gRPC servers and codecs per service; each returned file
// receives imports from the complete endpoint set it renders.
package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// serverFiles returns the planned server interfaces, encoders, and decoders.
func serverFiles(services *ServicesData) []*codegen.File {
	svcLen := len(services.servicePlans)
	fw := make([]*codegen.File, 2*svcLen)
	for i, servicePlan := range services.servicePlans {
		fw[i] = addEndpointImports(serverFile(servicePlan.expression, services), services)
	}
	for i, servicePlan := range services.servicePlans {
		fw[i+svcLen] = addEndpointImports(serverEncodeDecode(servicePlan.expression, services), services)
	}
	return fw
}

// serverFile returns the files defining the gRPC server.
func serverFile(svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = services.Get(svc.Name())
	)
	{
		svcName := data.Service.PathName
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "server", "server.go")
		sections = []*codegen.SectionTemplate{
			codegen.Header(svc.Name()+" gRPC server", "server", nil),
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
func serverEncodeDecode(svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = services.Get(svc.Name())
	)
	{
		svcName := data.Service.PathName
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "server", "encode_decode.go")
		title := fmt.Sprintf("%s gRPC server encoders and decoders", svc.Name())
		sections = []*codegen.SectionTemplate{codegen.Header(title, "server", nil)}

		for _, e := range data.Endpoints {
			if e.Response.ServerConvert != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "response-encoder",
					Source: grpcTemplates.Read(grpcResponseEncoderT, grpcTypeToStringExpressionP),
					Data:   e,
					FuncMap: map[string]any{
						"typeStringExpressionData": typeStringExpressionData,
						"metadataEncodeDecodeData": metadataEncodeDecodeData,
					},
				})
			}
			if e.PayloadRef != "" {
				fm := transTmplFuncs(data)
				fm["isEmpty"] = isEmpty
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "request-decoder",
					Source:  grpcTemplates.Read(grpcRequestDecoderT, grpcConvertStringToTypeP, "type_conversion", "slice_conversion", "slice_item_conversion", "metadata_decode"),
					Data:    e,
					FuncMap: fm,
				})
			}
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

// transTmplFuncs returns the type formatter used by metadata templates for one
// saved service.
func transTmplFuncs(service *ServiceData) map[string]any {
	return map[string]any{
		"goTypeRef": func(dt expr.DataType) string {
			return service.Scope.GoTypeRef(&expr.AttributeExpr{Type: dt})
		},
	}
}

// typeStringExpressionData describes one primitive value that generated code
// converts to a metadata string.
func typeStringExpressionData(dt expr.DataType, target string) map[string]any {
	return map[string]any{
		"Type":   dt,
		"Target": target,
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
