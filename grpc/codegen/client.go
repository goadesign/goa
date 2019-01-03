package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service"
	"goa.design/goa/expr"
)

// ClientFiles returns all the client gRPC transport files.
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
	path := filepath.Join(codegen.Gendir, "grpc", codegen.SnakeCase(svc.Name()), "client", "client.go")
	data := GRPCServices.Get(svc.Name())
	title := fmt.Sprintf("%s GRPC client", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "google.golang.org/grpc"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: "goa.design/goa/grpc", Name: "goagrpc"},
			{Path: filepath.Join(genpkg, codegen.SnakeCase(svc.Name())), Name: data.Service.PkgName},
			{Path: filepath.Join(genpkg, codegen.SnakeCase(svc.Name()), "views"), Name: data.Service.ViewsPkg},
			{Path: filepath.Join(genpkg, "grpc", codegen.SnakeCase(svc.Name()), pbPkgName)},
		}),
	}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "client-struct",
		Source: clientStructT,
		Data:   data,
	})
	for _, e := range data.Endpoints {
		if e.ClientStream != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-stream-struct-type",
				Source: streamStructTypeT,
				Data:   e.ClientStream,
			})
		}
	}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "client-init",
		Source: clientInitT,
		Data:   data,
	})
	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-endpoint-init",
			Source: clientEndpointInitT,
			Data:   e,
		})
	}
	for _, e := range data.Endpoints {
		if e.ClientStream != nil {
			if e.ClientStream.RecvConvert != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-recv",
					Source: streamRecvT,
					Data:   e.ClientStream,
				})
			}
			if e.Method.StreamKind == expr.ClientStreamKind || e.Method.StreamKind == expr.BidirectionalStreamKind {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-send",
					Source: streamSendT,
					Data:   e.ClientStream,
				})
			}
			if e.ServerStream.MustClose {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-close",
					Source: streamCloseT,
					Data:   e.ClientStream,
				})
			}
			if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-set-view",
					Source: streamSetViewT,
					Data:   e.ClientStream,
				})
			}
		}
	}
	return &codegen.File{Path: path, SectionTemplates: sections}
}

func clientEncodeDecode(genpkg string, svc *expr.GRPCServiceExpr) *codegen.File {
	var (
		path     string
		sections []*codegen.SectionTemplate

		data = GRPCServices.Get(svc.Name())
	)
	{
		path = filepath.Join(codegen.Gendir, "grpc", codegen.SnakeCase(svc.Name()), "client", "encode_decode.go")
		sections = []*codegen.SectionTemplate{
			codegen.Header(svc.Name()+" gRPC client encoders and decoders", "client", []*codegen.ImportSpec{
				{Path: "context"},
				{Path: "strconv"},
				{Path: "google.golang.org/grpc"},
				{Path: "google.golang.org/grpc/metadata"},
				{Path: "goa.design/goa", Name: "goa"},
				{Path: "goa.design/goa/grpc", Name: "goagrpc"},
				{Path: filepath.Join(genpkg, codegen.SnakeCase(svc.Name())), Name: data.Service.PkgName},
				{Path: filepath.Join(genpkg, codegen.SnakeCase(svc.Name()), "views"), Name: data.Service.ViewsPkg},
				{Path: filepath.Join(genpkg, "grpc", codegen.SnakeCase(svc.Name()), pbPkgName)},
			}),
		}
		fm := transTmplFuncs(svc)
		fm["metadataEncodeDecodeData"] = metadataEncodeDecodeData
		fm["typeConversionData"] = typeConversionData
		fm["isBearer"] = isBearer
		for _, e := range data.Endpoints {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "remote-method-builder",
				Source: remoteMethodBuilderT,
				Data:   e,
			})
			if e.PayloadRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "request-encoder",
					Source:  requestEncoderT,
					Data:    e,
					FuncMap: fm,
				})
			}
			if e.ResultRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "response-decoder",
					Source:  responseDecoderT,
					Data:    e,
					FuncMap: fm,
				})
			}
		}
	}
	return &codegen.File{Path: path, SectionTemplates: sections}
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

// input: ServiceData
const clientStructT = `{{ printf "%s lists the service endpoint gRPC clients." .ClientStruct | comment }}
type {{ .ClientStruct }} struct {
	grpccli {{ .PkgName }}.{{ .ClientInterface }}
	opts []grpc.CallOption
}
`

// input: ServiceData
const clientInitT = `{{ printf "New%s instantiates gRPC client for all the %s service servers." .ClientStruct .Service.Name | comment }}
func New{{ .ClientStruct }}(cc *grpc.ClientConn, opts ...grpc.CallOption) *{{ .ClientStruct }} {
  return &{{ .ClientStruct }}{
		grpccli: {{ .ClientInterfaceInit }}(cc),
		opts: opts,
	}
}
`

// input: EndpointData
const clientEndpointInitT = `{{ printf "%s calls the %q function in %s.%s interface." .Method.VarName .Method.VarName .PkgName .ClientInterface | comment }}
func (c *{{ .ClientStruct }}) {{ .Method.VarName }}() goa.Endpoint {
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		inv := goagrpc.NewInvoker(
			Build{{ .Method.VarName }}Func(c.grpccli, c.opts...),
			{{ if .PayloadRef }}Encode{{ .Method.VarName }}Request{{ else }}nil{{ end }},
			{{ if .ResultRef }}Decode{{ .Method.VarName }}Response{{ else }}nil{{ end }})
		res, err := inv.Invoke(ctx, v)
		if err != nil {
			return nil, goagrpc.DecodeError(err)
		}
		return res, nil
	}
}
`

// input: EndpointData
const remoteMethodBuilderT = `{{ printf "Build%sFunc builds the remote method to invoke for %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Build{{ .Method.VarName }}Func(grpccli {{ .PkgName }}.{{ .ClientInterface }}, cliopts ...grpc.CallOption) goagrpc.RemoteFunc {
	return func(ctx context.Context, reqpb interface{}, opts ...grpc.CallOption) (interface{}, error) {
		for _, opt := range cliopts {
			opts = append(opts, opt)
		}
		return grpccli.{{ .Method.VarName }}(ctx{{ if not .Method.StreamingPayload }}, reqpb.({{ .Request.ClientConvert.TgtRef }}){{ end }}, opts...)
	}
}
`

// input: EndpointData
const responseDecoderT = `{{ printf "Decode%sResponse decodes responses from the %s %s endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Decode{{ .Method.VarName }}Response(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
{{- if or .Response.Headers .Response.Trailers }}
	var (
	{{- range .Response.Headers }}
		{{ .VarName }} {{ .TypeRef }}
	{{- end }}
	{{- range .Response.Trailers }}
		{{ .VarName }} {{ .TypeRef }}
	{{- end }}
		err error
	)
	{
		{{- range .Response.Headers }}
			{{ template "metadata_decoder" (metadataEncodeDecodeData . "hdr") }}
			{{- if .Validate }}
				{{ .Validate }}
			{{- end }}
		{{- end }}
		{{- range .Response.Trailers }}
			{{ template "metadata_decoder" (metadataEncodeDecodeData . "trlr") }}
			{{- if .Validate }}
				{{ .Validate }}
			{{- end }}
		{{- end }}
	}
	if err != nil {
		return nil, err
	}
{{- end }}
{{- if .ViewedResultRef }}
  var view string
  {
    if vals := hdr.Get("goa-view"); len(vals) > 0 {
      view = vals[0]
    }
  }
{{- end }}
{{- if .ServerStream }}
	return &{{ .ClientStream.VarName }}{
		stream: v.({{ .ClientStream.Interface }}),
	{{- if .ViewedResultRef }}
		view: view,
	{{- end }}
	}, nil
{{- else }}
	message, ok := v.({{ .Response.ClientConvert.SrcRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .Response.ClientConvert.SrcRef }}", v)
	}
	{{- if .Response.ClientConvert.Validation }}
		err = {{ .Response.ClientConvert.Validation.Name }}(message)
	{{- end }}
	res := {{ .Response.ClientConvert.Init.Name }}({{ range .Response.ClientConvert.Init.Args }}{{ .Name }}, {{ end }})
	{{- if .ViewedResultRef }}
		vres := {{ if not .Method.ViewedResult.IsCollection }}&{{ end }}{{ .Method.ViewedResult.FullName }}{Projected: res}
		vres.View = view
		return {{ .ServicePkgName }}.{{ .Method.ViewedResult.ResultInit.Name }}({{ range .Method.ViewedResult.ResultInit.Args}}{{ .Name }}, {{ end }}), nil
	{{- else }}
		return res, nil
	{{- end }}
{{- end }}
}

{{- define "metadata_decoder" }}
	{{- if or (eq .Metadata.Type.Name "string") (eq .Metadata.Type.Name "any") }}
		{{- if .Metadata.Required }}
			if vals := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len(vals) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Metadata.Name }}, "metadata"))
			} else {
				{{ .Metadata.VarName }} = vals[0]
			}
		{{- else }}
			if vals := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len(vals) > 0 {
				{{ .Metadata.VarName }} = vals[0]
			}
		{{- end }}
	{{- else if .Metadata.StringSlice }}
		{{- if .Metadata.Required }}
			if vals := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len(vals) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Metadata.Name }}, "metadata"))
			} else {
				{{ .Metadata.VarName }} = vals
			}
		{{- else }}
			{{ .Metadata.VarName }} = {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }})
		{{- end }}
	{{- else if .Metadata.Slice }}
		{{- if .Metadata.Required }}
			if {{ .Metadata.VarName }}Raw := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len({{ .Metadata.VarName }}Raw) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Metadata.Name }}, "metadata"))
			} else {
				{{- template "slice_conversion" .Metadata }}
			}
		{{- else }}
			if {{ .Metadata.VarName }}Raw := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len({{ .Metadata.VarName }}Raw) > 0 {
				{{- template "slice_conversion" .Metadata }}
			}
		{{- end }}
	{{- else }}
		{{- if .Metadata.Required }}
			if vals := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len(vals) == 0 {
				err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Metadata.Name }}, "metadata"))
			} else {
				{{ .Metadata.VarName }}Raw = vals[0]
				{{ template "type_conversion" .Metadata }}
			}
		{{- else }}
			if vals := {{ .VarName }}.Get({{ printf "%q" .Metadata.Name }}); len(vals) > 0 {
				{{ .Metadata.VarName }}Raw = vals[0]
				{{ template "type_conversion" .Metadata }}
			}
		{{- end }}
	{{- end }}
{{- end }}
` + convertStringToTypeT

// input: EndpointData
const requestEncoderT = `{{ printf "Encode%sRequest encodes requests sent to %s %s endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Encode{{ .Method.VarName }}Request(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.({{ .PayloadRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .PayloadRef }}", v)
	}
{{- range .Request.Metadata }}
	{{- if .StringSlice }}
		for _, value := range payload{{ if .FieldName }}.{{ .FieldName }}{{ end }} {
			(*md).Append({{ printf "%q" .Name }}, value)
		}
	{{- else if .Slice }}
		for _, value := range payload{{ if .FieldName }}.{{ .FieldName }}{{ end }} {
			{{ template "string_conversion" (typeConversionData .Type.ElemType.Type "valueStr" "value") }}
			(*md).Append({{ printf "%q" .Name }}, valueStr)
		}
	{{- else }}
		{{- if .Pointer }}
			if payload{{ if .FieldName }}.{{ .FieldName }}{{ end }} != nil {
		{{- end }}
			{{- if (and (eq .Name "Authorization") (isBearer $.MetadataSchemes)) }}
				if !strings.Contains({{ if .Pointer }}*{{ end }}payload{{ if .FieldName }}.{{ .FieldName }}{{ end }}, " ") {
					(*md).Append(ctx, {{ printf "%q" .Name }}, "Bearer "+{{ if .Pointer }}*{{ end }}payload{{ if .FieldName }}.{{ .FieldName }}{{ end }})
				} else {
			{{- end }}
				(*md).Append({{ printf "%q" .Name }},
					{{- if eq .Type.Name "bytes" }} string(
					{{- else if not (eq .Type.Name "string") }} fmt.Sprintf("%v",
					{{- end }}
					{{- if .Pointer }}*{{ end }}payload{{ if .FieldName }}.{{ .FieldName }}{{ end }}
					{{- if or (eq .Type.Name "bytes") (not (eq .Type.Name "string")) }})
					{{- end }})
			{{- if (and (eq .Name "Authorization") (isBearer $.MetadataSchemes)) }}
				}
			{{- end }}
		{{- if .Pointer }}
			}
		{{- end }}
	{{- end }}
{{- end }}
{{- if .Request.ClientConvert }}
	return {{ .Request.ClientConvert.Init.Name }}({{ range .Request.ClientConvert.Init.Args }}{{ .Name }}, {{ end }}), nil
{{- else }}
	return nil, nil
{{- end }}
}
` + convertTypeToStringT
