package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// ClientFiles returns the client files that contain client methods to call the
// corresponding service methods along with the encoding and decoding logic.
func ClientFiles(genpkg string, services *ServicesData) []*codegen.File {
	svcLen := len(services.Root.API.GRPC.Services)
	fw := make([]*codegen.File, 2*svcLen)
	for i, svc := range services.Root.API.GRPC.Services {
		fw[i] = clientFile(genpkg, svc, services)
	}
	for i, svc := range services.Root.API.GRPC.Services {
		fw[i+svcLen] = clientEncodeDecode(genpkg, svc, services)
	}
	return fw
}

// clientFile returns the file implementing the gRPC client.
func clientFile(genpkg string, svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = services.Get(svc.Name())
	)
	{
		svcName := data.Service.PathName
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "client", "client.go")
		imports := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "google.golang.org/grpc"},
			codegen.GoaImport(""),
			codegen.GoaNamedImport("grpc", "goagrpc"),
			codegen.GoaNamedImport("grpc/pb", "goapb"),
			{Path: path.Join(genpkg, svcName), Name: data.Service.PkgName},
			{Path: path.Join(genpkg, svcName, "views"), Name: data.Service.ViewsPkg},
			{Path: path.Join(genpkg, "grpc", svcName, pbPkgName), Name: data.PkgName},
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
func clientEncodeDecode(genpkg string, svc *expr.GRPCServiceExpr, services *ServicesData) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = services.Get(svc.Name())
	)
	{
		svcName := data.Service.PathName
		fpath = filepath.Join(codegen.Gendir, "grpc", svcName, "client", "encode_decode.go")
		imports := []*codegen.ImportSpec{
			{Path: "fmt"},
			{Path: "context"},
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
		sections = []*codegen.SectionTemplate{codegen.Header(svc.Name()+" gRPC client encoders and decoders", "client", imports)}
		fm := transTmplFuncs(svc, services)
		fm["metadataEncodeDecodeData"] = metadataEncodeDecodeData
		fm["typeConversionData"] = typeConversionData
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
					Source:  grpcTemplates.Read(grpcRequestEncoderT, grpcConvertTypeToStringP, "string_conversion"),
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

// isBearer returns true if the security scheme uses a Bearer scheme.
func isBearer(schemes []*service.SchemeData) bool {
	for _, s := range schemes {
		if s.Name != "Authorization" {
			continue
		}
		if s.Type == "JWT" || s.Type == "OAuth2" {
			return true
		}
	}
	return false
}
