package codegen

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// ClientFiles returns the client implementation for every gRPC service. The
// files include the client which invokes the protoc-generated gRPC client
// and encoders and decoders to transform protocol buffer types and gRPC
// metadata into goa types and vice versa.
func ClientFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	svcLen := len(root.API.GRPC.Services)
	fw := make([]*codegen.File, 2*svcLen)
	for i, svc := range root.API.GRPC.Services {
		fw[i] = client(genpkg, svc)
	}
	for i, svc := range root.API.GRPC.Services {
		fw[i+svcLen] = clientEncodeDecode(genpkg, svc)
	}
	return fw
}

// client returns the files defining the gRPC client.
func client(genpkg string, svc *expr.GRPCServiceExpr) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = GRPCServices.Get(svc.Name())
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
		imports = append(imports, data.Service.UserTypeImports...)
		sections = []*codegen.SectionTemplate{
			codegen.Header(svc.Name()+" gRPC client", "client", imports),
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-struct",
			Source: readTemplate("client_struct"),
			Data:   data,
		})
		for _, e := range data.Endpoints {
			if e.ClientStream != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-struct-type",
					Source: readTemplate("stream_struct_type"),
					Data:   e.ClientStream,
				})
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-init",
			Source: readTemplate("client_init"),
			Data:   data,
		})
		for _, e := range data.Endpoints {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-endpoint-init",
				Source: readTemplate("client_endpoint_init"),
				Data:   e,
			})
		}
		for _, e := range data.Endpoints {
			if e.ClientStream != nil {
				if e.ClientStream.RecvConvert != nil {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-stream-recv",
						Source: readTemplate("stream_recv"),
						Data:   e.ClientStream,
					})
				}
				if e.Method.StreamKind == expr.ClientStreamKind || e.Method.StreamKind == expr.BidirectionalStreamKind {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-stream-send",
						Source: readTemplate("stream_send"),
						Data:   e.ClientStream,
					})
				}
				if e.ClientStream.MustClose {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-stream-close",
						Source: readTemplate("stream_close"),
						Data:   e.ClientStream,
					})
				}
				if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "client-stream-set-view",
						Source: readTemplate("stream_set_view"),
						Data:   e.ClientStream,
					})
				}
			}
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

func clientEncodeDecode(genpkg string, svc *expr.GRPCServiceExpr) *codegen.File {
	var (
		fpath    string
		sections []*codegen.SectionTemplate

		data = GRPCServices.Get(svc.Name())
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
		imports = append(imports, data.Service.UserTypeImports...)
		sections = []*codegen.SectionTemplate{codegen.Header(svc.Name()+" gRPC client encoders and decoders", "client", imports)}
		fm := transTmplFuncs(svc)
		fm["metadataEncodeDecodeData"] = metadataEncodeDecodeData
		fm["typeConversionData"] = typeConversionData
		fm["isBearer"] = isBearer
		for _, e := range data.Endpoints {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "remote-method-builder",
				Source: readTemplate("remote_method_builder"),
				Data:   e,
			})
			if e.PayloadRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "request-encoder",
					Source:  readTemplate("request_encoder", "convert_type_to_string"),
					Data:    e,
					FuncMap: fm,
				})
			}
			if e.ResultRef != "" || e.ClientStream != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "response-decoder",
					Source:  readTemplate("response_decoder", "convert_string_to_type"),
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
