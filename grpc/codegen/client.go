// This file renders gRPC clients and codecs per service; each returned file
// owns the generated-type imports used by its conversions.
package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// clientFiles returns the planned client methods and their encoders and decoders.
func clientFiles(services *ServicesData) []*codegen.File {
	svcLen := len(services.servicePlans)
	fw := make([]*codegen.File, 2*svcLen)
	for i, servicePlan := range services.servicePlans {
		fw[i] = addEndpointImports(clientFile(servicePlan.expression, services), services, servicePlan)
	}
	for i, servicePlan := range services.servicePlans {
		fw[i+svcLen] = addEndpointImports(clientEncodeDecode(servicePlan.expression, services), services, servicePlan)
	}
	return fw
}

// clientFile returns the file implementing the gRPC client.
func clientFile(svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = services.Get(svc.Name())
	)
	{
		svcName := data.Service.PathName
		outputPackage := path.Join(services.GenPkg(), "grpc", svcName, "client")
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "client", "client.go")
		imports := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "google.golang.org/grpc"},
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			codegen.GoaNamedImport("grpc/pb", "goapb"),
			services.ServiceImport(outputPackage, svc.Name()),
			services.PackageImport(outputPackage, path.Join(services.GenPkg(), "grpc", svcName, pbPkgName)),
		}
		if serviceHasViewedClientStream(data) {
			imports = append(imports, services.ViewImport(outputPackage, svc.Name()))
		}
		sections = []*codegen.SectionTemplate{
			codegen.Header(svc.Name()+" gRPC client", "client", imports),
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-struct",
			Source: grpcTemplates.Read(grpcClientStructT),
			Data:   data,
		})
		for _, e := range data.Endpoints {
			if e.ClientStream != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-struct-type",
					Source: grpcTemplates.Read(grpcStreamStructTypeT),
					Data:   e.ClientStream,
				})
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "grpc-client-init",
			Source: grpcTemplates.Read(grpcClientInitT),
			Data:   data,
		})
		for _, e := range data.Endpoints {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-endpoint-init",
				Source: grpcTemplates.Read(grpcClientEndpointInitT),
				Data:   e,
			})
		}
		for _, e := range data.Endpoints {
			if e.ClientStream != nil {
				if e.ClientStream.RecvConvert != nil {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-stream-recv",
						Source: grpcTemplates.Read(grpcStreamRecvT),
						Data:   e.ClientStream,
					})
				}
				if e.Method.StreamKind == expr.ClientStreamKind || e.Method.StreamKind == expr.BidirectionalStreamKind {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-stream-send",
						Source: grpcTemplates.Read(grpcStreamSendT),
						Data:   e.ClientStream,
					})
				}
				if e.ClientStream.MustClose {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-stream-close",
						Source: grpcTemplates.Read(grpcStreamCloseT),
						Data:   e.ClientStream,
					})
				}
				if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-stream-set-view",
						Source: grpcTemplates.Read(grpcStreamSetViewT),
						Data:   e.ClientStream,
					})
				}
			}
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

// clientEncodeDecode returns the file containing the gRPC client encoding and
// decoding logic.
func clientEncodeDecode(svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = services.Get(svc.Name())
	)
	{
		svcName := data.Service.PathName
		outputPackage := path.Join(services.GenPkg(), "grpc", svcName, "client")
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "client", "encode_decode.go")
		imports := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "strconv"},
			{Path: "unicode/utf8"},
			{Path: "google.golang.org/grpc"},
			{Path: "google.golang.org/grpc/metadata"},
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			services.ServiceImport(outputPackage, svc.Name()),
			services.PackageImport(outputPackage, path.Join(services.GenPkg(), "grpc", svcName, pbPkgName)),
		}
		if requestMetadataNeedsFormat(data) {
			imports = append(imports, &codegen.ImportSpec{Path: "fmt"})
		}
		if serviceHasUnaryViewedResult(data) {
			imports = append(imports, services.ViewImport(outputPackage, svc.Name()))
		}
		sections = []*codegen.SectionTemplate{codegen.Header(svc.Name()+" gRPC client encoders and decoders", "client", imports)}
		fm := transTmplFuncs(data)
		fm["hasInitArg"] = hasInitArg
		fm["metadataEncodeDecodeData"] = metadataEncodeDecodeData
		fm["typeStringExpressionData"] = typeStringExpressionData
		fm["isBearer"] = isBearer
		for _, e := range data.Endpoints {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "remote-method-builder",
				Source: grpcTemplates.Read(grpcRemoteMethodBuilderT),
				Data:   e,
			})
			if e.PayloadRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "request-encoder",
					Source:  grpcTemplates.Read(grpcRequestEncoderT, grpcTypeToStringExpressionP),
					Data:    e,
					FuncMap: fm,
				})
			}
			if e.ResultRef != "" || e.ClientStream != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "response-decoder",
					Source:  grpcTemplates.Read(grpcResponseDecoderT, grpcConvertStringToTypeP, "type_conversion", "slice_conversion", "slice_item_conversion"),
					Data:    e,
					FuncMap: fm,
				})
			}
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

// hasInitArg reports whether a generated constructor consumes the named
// source variable. Templates use it to avoid declaring an unused variable for
// an empty protobuf message that only carries response metadata.
func hasInitArg(args []*InitArgData, name string) bool {
	for _, arg := range args {
		if arg.Name == name {
			return true
		}
	}
	return false
}

// isBearer returns true if the security scheme uses a Bearer scheme.
func isBearer(schemes []*service.SchemeData) bool {
	for _, s := range schemes {
		if s.Name != "Authorization" {
			continue
		}
		if s.Type == "Bearer" || s.Type == "JWT" || s.Type == "OAuth2" {
			return true
		}
	}
	return false
}
