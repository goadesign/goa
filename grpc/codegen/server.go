// This file renders gRPC servers and codecs per service; each returned file
// receives imports from the complete endpoint set it renders.
package codegen

import (
	"fmt"
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// serverFiles returns the planned server interfaces, encoders, and decoders.
func serverFiles(services *ServicesData) []*codegen.File {
	svcLen := len(services.servicePlans)
	fw := make([]*codegen.File, 2*svcLen)
	for i, servicePlan := range services.servicePlans {
		fw[i] = addEndpointImports(serverFile(servicePlan.expression, services), services, servicePlan)
	}
	for i, servicePlan := range services.servicePlans {
		fw[i+svcLen] = addEndpointImports(serverEncodeDecode(servicePlan.expression, services), services, servicePlan)
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
		outputPackage := path.Join(services.GenPkg(), "grpc", svcName, "server")
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "server", "server.go")
		imports := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "errors"},
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			{Path: "google.golang.org/grpc/codes"},
			services.ServiceImport(outputPackage, svc.Name()),
			services.PackageImport(outputPackage, path.Join(services.GenPkg(), "grpc", svcName, pbPkgName)),
		}
		for _, e := range data.Endpoints {
			if e.Request.StreamEnvelope != nil {
				imports = append(imports, &codegen.ImportSpec{Path: "io"})
				break
			}
		}
		if serviceHasCallerSelectedViewedServerStream(data) {
			imports = append(imports, &codegen.ImportSpec{Path: "google.golang.org/grpc/metadata"})
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
func serverEncodeDecode(svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = services.Get(svc.Name())
	)
	{
		svcName := data.Service.PathName
		outputPackage := path.Join(services.GenPkg(), "grpc", svcName, "server")
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
			services.ServiceImport(outputPackage, svc.Name()),
			services.PackageImport(outputPackage, path.Join(services.GenPkg(), "grpc", svcName, pbPkgName)),
		}
		if serviceHasViewedResult(data) {
			imports = append(imports, services.ViewImport(outputPackage, svc.Name()))
		}
		if responseMetadataNeedsFormat(data) {
			imports = append(imports, &codegen.ImportSpec{Path: "fmt"})
		}
		sections = []*codegen.SectionTemplate{codegen.Header(title, "server", imports)}

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

// requestMetadataNeedsFormat reports whether request metadata can contain a Go
// value whose concrete type is unknown until the client runs.
func requestMetadataNeedsFormat(service *ServiceData) bool {
	for _, endpoint := range service.Endpoints {
		if metadataNeedsFormat(endpoint.Request.Metadata) {
			return true
		}
	}
	return false
}

// responseMetadataNeedsFormat reports whether response metadata can contain a
// Go value whose concrete type is unknown until the server runs.
func responseMetadataNeedsFormat(service *ServiceData) bool {
	for _, endpoint := range service.Endpoints {
		for _, group := range [][]*MetadataData{endpoint.Response.Headers, endpoint.Response.Trailers} {
			if metadataNeedsFormat(group) {
				return true
			}
		}
	}
	return false
}

// metadataNeedsFormat reports whether one metadata field uses Goa's Any type.
// All other supported metadata types have an exact string conversion.
func metadataNeedsFormat(fields []*MetadataData) bool {
	for _, field := range fields {
		typeKind := field.Type.Kind()
		if array := expr.AsArray(field.Type); array != nil {
			typeKind = array.ElemType.Type.Kind()
		}
		if typeKind == expr.AnyKind {
			return true
		}
	}
	return false
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
