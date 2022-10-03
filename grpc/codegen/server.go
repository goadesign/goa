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
			{Name: "server-struct", Source: serverStructT, Data: data},
		}
		for _, e := range data.Endpoints {
			if e.ServerStream != nil {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-stream-struct-type",
					Source: streamStructTypeT,
					Data:   e.ServerStream,
				})
			}
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-init",
			Source: serverInitT,
			Data:   data,
		})
		for _, e := range data.Endpoints {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "grpc-handler-init",
				Source: handlerInitT,
				Data:   e,
			})
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-grpc-interface",
				Source: serverGRPCInterfaceT,
				Data:   e,
			})
		}
		for _, e := range data.Endpoints {
			if e.ServerStream != nil {
				if e.ServerStream.SendConvert != nil {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-send",
						Source: streamSendT,
						Data:   e.ServerStream,
					})
				}
				if e.Method.StreamKind == expr.ClientStreamKind || e.Method.StreamKind == expr.BidirectionalStreamKind {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-recv",
						Source: streamRecvT,
						Data:   e.ServerStream,
					})
				}
				if e.ServerStream.MustClose {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-close",
						Source: streamCloseT,
						Data:   e.ServerStream,
					})
				}
				if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
					sections = append(sections, &codegen.SectionTemplate{
						Name:   "server-stream-set-view",
						Source: streamSetViewT,
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
					Source: responseEncoderT,
					Data:   e,
					FuncMap: map[string]interface{}{
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
					Source:  requestDecoderT,
					Data:    e,
					FuncMap: fm,
				})
			}
		}
	}
	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

func transTmplFuncs(s *expr.GRPCServiceExpr) map[string]interface{} {
	return map[string]interface{}{
		"goTypeRef": func(dt expr.DataType) string {
			return service.Services.Get(s.Name()).Scope.GoTypeRef(&expr.AttributeExpr{Type: dt})
		},
	}
}

// typeConversionData produces the template data suitable for executing the
// "type_conversion" template.
func typeConversionData(dt expr.DataType, varName string, target string) map[string]interface{} {
	return map[string]interface{}{
		"Type":    dt,
		"VarName": varName,
		"Target":  target,
	}
}

// metadataEncodeDecodeData produces the template data suitable for executing the
// "metadata_decoder" and "metadata_encoder" template.
func metadataEncodeDecodeData(md *MetadataData, vname string) map[string]interface{} {
	return map[string]interface{}{
		"Metadata": md,
		"VarName":  vname,
	}
}

// input: ServiceData
const serverStructT = `{{ printf "%s implements the %s.%s interface." .ServerStruct .PkgName .ServerInterface | comment }}
type {{ .ServerStruct }} struct {
{{- range .Endpoints }}
	{{ .Method.VarName }}H {{ if .ServerStream }}goagrpc.StreamHandler{{ else }}goagrpc.UnaryHandler{{ end }}
{{- end }}
	{{ .PkgName }}.Unimplemented{{ .ServerInterface }}
}
`

// input: ServiceData
const serverInitT = `{{ printf "%s instantiates the server struct with the %s service endpoints." .ServerInit .Service.Name | comment }}
func {{ .ServerInit }}(e *{{ .Service.PkgName }}.Endpoints{{ if .HasUnaryEndpoint }}, uh goagrpc.UnaryHandler{{ end }}{{ if .HasStreamingEndpoint }}, sh goagrpc.StreamHandler{{ end }}) *{{ .ServerStruct }} {
	return &{{ .ServerStruct }}{
	{{- range .Endpoints }}
		{{ .Method.VarName }}H: New{{ .Method.VarName }}Handler(e.{{ .Method.VarName }}{{ if .ServerStream }}, sh{{ else }}, uh{{ end }}),
	{{- end }}
	}
}
`

// input: EndpointData
const handlerInitT = `{{ printf "New%sHandler creates a gRPC handler which serves the %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func New{{ .Method.VarName }}Handler(endpoint goa.Endpoint, h goagrpc.{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler) goagrpc.{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler {
	if h == nil {
		h = goagrpc.New{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler(endpoint, {{ if .Method.Payload }}Decode{{ .Method.VarName }}Request{{ else }}nil{{ end }}{{ if not .ServerStream }}, Encode{{ .Method.VarName }}Response{{ end }})
	}
	return h
}
`

// input: EndpointData
const serverGRPCInterfaceT = `{{ printf "%s implements the %q method in %s.%s interface." .Method.VarName .Method.VarName .PkgName .ServerInterface | comment }}
func (s *{{ .ServerStruct }}) {{ .Method.VarName }}(
	{{- if not .ServerStream }}ctx context.Context, {{ end }}
	{{- if not .Method.StreamingPayload }}message {{ .Request.Message.Ref }}{{ if .ServerStream }}, {{ end }}{{ end }}
	{{- if .ServerStream }}stream {{ .ServerStream.Interface }}{{ end }}) {{ if .ServerStream }}error{{ else if .Response.Message }}({{ .Response.Message.Ref }},	error{{ if .Response.Message }}){{ end }}{{ end }} {
{{- if .ServerStream }}
	ctx := stream.Context()
{{- end }}
	ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
	ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})

{{- if .ServerStream }}
	{{if .PayloadRef }}p{{ else }}_{{ end }}, err := s.{{ .Method.VarName }}H.Decode(ctx, {{ if .Method.StreamingPayload }}nil{{ else }}message{{ end }})
	{{- template "handle_error" . }}
	ep := &{{ .ServicePkgName }}.{{ .Method.VarName }}EndpointInput{
		Stream: &{{ .ServerStream.VarName }}{stream: stream},
	{{- if .PayloadRef }}
		Payload: p.({{ .PayloadRef }}),
	{{- end }}
	}
	err = s.{{ .Method.VarName }}H.Handle(ctx, ep)
{{- else }}
	resp, err := s.{{ .Method.VarName }}H.Handle(ctx, message)
{{- end }}
	{{- template "handle_error" . }}
	return {{ if not $.ServerStream }}resp.({{ .Response.ServerConvert.TgtRef }}), {{ end }}nil
}

{{- define "handle_error" }}
	if err != nil {
	{{- if .Errors }}
		var en goa.GoaErrorNamer
		if errors.As(err, &en) {
			switch en.GoaErrorName() {
		{{- range .Errors }}
			case {{ printf "%q" .Name }}:
				{{- if .Response.ServerConvert }}
					var er {{ .Response.ServerConvert.SrcRef }}
					errors.As(err, &er)
				{{- end }}
				return {{ if not $.ServerStream }}nil, {{ end }}goagrpc.NewStatusError({{ .Response.StatusCode }}, err, {{ if .Response.ServerConvert }}{{ .Response.ServerConvert.Init.Name }}({{ range .Response.ServerConvert.Init.Args }}{{ .Name }}, {{ end }}){{ else }}goagrpc.NewErrorResponse(err){{ end }})
		{{- end }}
			}
		}
	{{- end }}
		return {{ if not $.ServerStream }}nil, {{ end }}goagrpc.EncodeError(err)
	}
{{- end }}
`

// input: EndpointData
const requestDecoderT = `{{ printf "Decode%sRequest decodes requests sent to %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Decode{{ .Method.VarName }}Request(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
{{- if .Request.Metadata }}
	var (
	{{- range .Request.Metadata }}
		{{ .VarName }} {{ .TypeRef }}
	{{- end }}
		err error
	)
	{
	{{- range .Request.Metadata }}
		{{- if or (eq .TypeName "string") (eq .Type.Name "any") }}
			{{- if .Required }}
				if vals := md.Get({{ printf "%q" .Name }}); len(vals) == 0 {
					err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, "metadata"))
				} else {
					{{ .VarName }} = vals[0]
				}
			{{- else }}
				if vals := md.Get({{ printf "%q" .Name }}); len(vals) > 0 {
					{{ .VarName }} = {{ if .Pointer }}&{{ end }}vals[0]
				}
			{{- end }}
		{{- else if .StringSlice }}
			{{- if .Required }}
				if vals := md.Get({{ printf "%q" .Name }}); len(vals) == 0 {
					err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, "metadata"))
				} else {
					{{ .VarName }} = vals
				}
			{{- else }}
				{{ .VarName }} = md.Get({{ printf "%q" .Name }})
			{{- end }}
		{{- else if .Slice }}
			{{- if .Required }}
				if {{ .VarName }}Raw := md.Get({{ printf "%q" .Name }}); len({{ .VarName }}Raw) == 0 {
					err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, "metadata"))
				} else {
					{{- template "slice_conversion" . }}
				}
			{{- else }}
				if {{ .VarName }}Raw := md.Get({{ printf "%q" .Name }}); len({{ .VarName }}Raw) > 0 {
					{{- template "slice_conversion" . }}
				}
			{{- end }}
		{{- else }}
			{{- if .Required }}
				if vals := md.Get({{ printf "%q" .Name }}); len(vals) == 0 {
					err = goa.MergeErrors(err, goa.MissingFieldError({{ printf "%q" .Name }}, "metadata"))
				} else {
					{{ .VarName }}Raw := vals[0]
					{{ template "type_conversion" . }}
				}
			{{- else }}
				if vals := md.Get({{ printf "%q" .Name }}); len(vals) > 0 {
					{{ .VarName }}Raw := vals[0]
					{{ template "type_conversion" . }}
				}
			{{- end }}
		{{- end }}
		{{- if .Validate }}
			{{ .Validate }}
		{{- end }}
	{{- end }}
	}
	if err != nil {
		return nil, err
	}
{{- end }}
{{- if and (not .Method.StreamingPayload) (not (isEmpty .Request.Message.Type)) }}
	var (
		message {{ .Request.ServerConvert.SrcRef }}
		ok bool
	)
	{
		if message, ok = v.({{ .Request.ServerConvert.SrcRef }}); !ok {
			return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .Request.Message.Ref }}", v)
		}
	{{- if .Request.ServerConvert.Validation }}
		if err {{ if .Request.Metadata }}={{ else }}:={{ end }} {{ .Request.ServerConvert.Validation.Name }}(message); err != nil {
			return nil, err
		}
	{{- end }}
	}
{{- end }}
	var payload {{ .PayloadRef }}
	{
		{{- if .Request.ServerConvert }}
			payload = {{ .Request.ServerConvert.Init.Name }}({{ range .Request.ServerConvert.Init.Args }}{{ .Name }}, {{ end }})
		{{- else }}
			payload = {{ (index .Request.Metadata 0).VarName }}
		{{- end }}
{{- range .MetadataSchemes }}
	{{- if ne .Type "Basic" }}
		{{- if not .CredRequired }}
			if payload.{{ .CredField }} != nil {
		{{- end }}
		if strings.Contains({{ if .CredPointer }}*{{ end }}payload.{{ .CredField }}, " ") {
			// Remove authorization scheme prefix (e.g. "Bearer")
			cred := strings.SplitN({{ if .CredPointer }}*{{ end }}payload.{{ .CredField }}, " ", 2)[1]
			payload.{{ .CredField }} = {{ if .CredPointer }}&{{ end }}cred
		}
		{{- if not .CredRequired }}
			}
		{{- end }}
	{{- end }}
{{- end }}
	}
	return payload, nil
}
` + convertStringToTypeT

// input: EndpointData
const responseEncoderT = `{{ printf "Encode%sResponse encodes responses from the %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Encode{{ .Method.VarName }}Response(ctx context.Context, v interface{}, hdr, trlr *metadata.MD) (interface{}, error) {
{{- if .ViewedResultRef }}
	vres, ok := v.({{ .ViewedResultRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .ViewedResultRef }}", v)
	}
	result := vres.Projected
	(*hdr).Append("goa-view", vres.View)
{{- else if .ResultRef }}
	result, ok := v.({{ .ResultRef }})
	if !ok {
		return nil, goagrpc.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .ResultRef }}", v)
	}
{{- end }}
	resp := {{ .Response.ServerConvert.Init.Name }}({{ range .Response.ServerConvert.Init.Args }}{{ .Name }}, {{ end }})
{{- range .Response.Headers }}
	{{ template "metadata_encoder" (metadataEncodeDecodeData . "(*hdr)") }}
{{- end }}
{{- range .Response.Trailers }}
	{{ template "metadata_encoder" (metadataEncodeDecodeData . "(*trlr)") }}
{{- end }}
	return resp, nil
}

{{- define "metadata_encoder" }}
	{{- if .Metadata.StringSlice }}
	{{ .VarName }}.Append({{ printf "%q" .Metadata.Name }}, res.{{ .Metadata.FieldName }}...)
	{{- else if .Metadata.Slice }}
		for _, value := range res.{{ .Metadata.FieldName }} {
			{{ template "string_conversion" (typeConversionData .Metadata.Type.ElemType.Type "valueStr" "value") }}
			{{ .VarName }}.Append({{ printf "%q" .Metadata.Name }}, valueStr)
		}
	{{- else }}
		{{- if .Metadata.Pointer }}
			if res.{{ .Metadata.FieldName }} != nil {
		{{- end }}
		{{ .VarName }}.Append({{ printf "%q" .Metadata.Name }},
			{{- if eq .Metadata.Type.Name "bytes" }} string(
			{{- else if not (eq .Metadata.TypeName "string") }} fmt.Sprintf("%v",
			{{- end }}
			{{- if .Metadata.Pointer }}*{{ end }}p.{{ .Metadata.FieldName }}
			{{- if or (eq .Metadata.Type.Name "bytes") (not (eq .Metadata.TypeName "string")) }})
			{{- end }})
		{{- if .Metadata.Pointer }}
			}
		{{- end }}
	{{- end }}
{{- end }}
` + convertTypeToStringT

// input: TypeData
const convertStringToTypeT = `{{- define "slice_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for i, rv := range {{ .VarName }}Raw {
		{{- template "slice_item_conversion" . }}
	}
{{- end }}

{{- define "slice_item_conversion" }}
	{{- if eq .Type.ElemType.Type.Name "string" }}
		{{ .VarName }}[i] = rv
	{{- else if eq .Type.ElemType.Type.Name "bytes" }}
		{{ .VarName }}[i] = []byte(rv)
	{{- else if eq .Type.ElemType.Type.Name "int" }}
		v, err2 := strconv.ParseInt(rv, 10, strconv.IntSize)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of integers"))
		}
		{{ .VarName }}[i] = int(v)
	{{- else if eq .Type.ElemType.Type.Name "int32" }}
		v, err2 := strconv.ParseInt(rv, 10, 32)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of integers"))
		}
		{{ .VarName }}[i] = int32(v)
	{{- else if eq .Type.ElemType.Type.Name "int64" }}
		v, err2 := strconv.ParseInt(rv, 10, 64)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of integers"))
		}
		{{ .VarName }}[i] = v
	{{- else if eq .Type.ElemType.Type.Name "uint" }}
		v, err2 := strconv.ParseUint(rv, 10, strconv.IntSize)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of unsigned integers"))
		}
		{{ .VarName }}[i] = uint(v)
	{{- else if eq .Type.ElemType.Type.Name "uint32" }}
		v, err2 := strconv.ParseUint(rv, 10, 32)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of unsigned integers"))
		}
		{{ .VarName }}[i] = int32(v)
	{{- else if eq .Type.ElemType.Type.Name "uint64" }}
		v, err2 := strconv.ParseUint(rv, 10, 64)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of unsigned integers"))
		}
		{{ .VarName }}[i] = v
	{{- else if eq .Type.ElemType.Type.Name "float32" }}
		v, err2 := strconv.ParseFloat(rv, 32)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of floats"))
		}
		{{ .VarName }}[i] = float32(v)
	{{- else if eq .Type.ElemType.Type.Name "float64" }}
		v, err2 := strconv.ParseFloat(rv, 64)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of floats"))
		}
		{{ .VarName }}[i] = v
	{{- else if eq .Type.ElemType.Type.Name "boolean" }}
		v, err2 := strconv.ParseBool(rv)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "array of booleans"))
		}
		{{ .VarName }}[i] = v
	{{- else if eq .Type.ElemType.Type.Name "any" }}
		{{ .VarName }}[i] = rv
	{{- else }}
		// unsupported slice type {{ .Type.ElemType.Type.Name }} for var {{ .VarName }}
	{{- end }}
{{- end }}

{{- define "type_conversion" }}
	{{- if eq .Type.Name "bytes" }}
		{{ .VarName }} = []byte({{ .VarName }}Raw)
	{{- else if eq .Type.Name "int" }}
		v, err2 := strconv.ParseInt({{ .VarName }}Raw, 10, strconv.IntSize)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "integer"))
		}
		{{- if .Pointer }}
		pv := int(v)
		{{ .VarName }} = &pv
		{{- else }}
		{{ .VarName }} = int(v)
		{{- end }}
	{{- else if eq .Type.Name "int32" }}
		v, err2 := strconv.ParseInt({{ .VarName }}Raw, 10, 32)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "integer"))
		}
		{{- if .Pointer }}
		pv := int32(v)
		{{ .VarName }} = &pv
		{{- else }}
		{{ .VarName }} = int32(v)
		{{- end }}
	{{- else if eq .Type.Name "int64" }}
		v, err2 := strconv.ParseInt({{ .VarName }}Raw, 10, 64)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "integer"))
		}
		{{ .VarName }} = {{ if .Pointer}}&{{ end }}v
	{{- else if eq .Type.Name "uint" }}
		v, err2 := strconv.ParseUint({{ .VarName }}Raw, 10, strconv.IntSize)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "unsigned integer"))
		}
		{{- if .Pointer }}
		pv := uint(v)
		{{ .VarName }} = &pv
		{{- else }}
		{{ .VarName }} = uint(v)
		{{- end }}
	{{- else if eq .Type.Name "uint32" }}
		v, err2 := strconv.ParseUint({{ .VarName }}Raw, 10, 32)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "unsigned integer"))
		}
		{{- if .Pointer }}
		pv := uint32(v)
		{{ .VarName }} = &pv
		{{- else }}
		{{ .VarName }} = uint32(v)
		{{- end }}
	{{- else if eq .Type.Name "uint64" }}
		v, err2 := strconv.ParseUint({{ .VarName }}Raw, 10, 64)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "unsigned integer"))
		}
		{{ .VarName }} = {{ if .Pointer }}&{{ end }}v
	{{- else if eq .Type.Name "float32" }}
		v, err2 := strconv.ParseFloat({{ .VarName }}Raw, 32)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "float"))
		}
		{{- if .Pointer }}
		pv := float32(v)
		{{ .VarName }} = &pv
		{{- else }}
		{{ .VarName }} = float32(v)
		{{- end }}
	{{- else if eq .Type.Name "float64" }}
		v, err2 := strconv.ParseFloat({{ .VarName }}Raw, 64)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "float"))
		}
		{{ .VarName }} = {{ if .Pointer }}&{{ end }}v
	{{- else if eq .Type.Name "boolean" }}
		v, err2 := strconv.ParseBool({{ .VarName }}Raw)
		if err2 != nil {
			err = goa.MergeErrors(err, goa.InvalidFieldTypeError({{ printf "%q" .VarName }}, {{ .VarName}}Raw, "boolean"))
		}
		{{ .VarName }} = {{ if .Pointer }}&{{ end }}v
	{{- else }}
		// unsupported type {{ .Type.Name }} for var {{ .VarName }}
	{{- end }}
{{- end }}
`

// input: TypeData
const convertTypeToStringT = `{{- define "string_conversion" }}
	{{- if eq .Type.Name "boolean" -}}
		{{ .VarName }} := strconv.FormatBool({{ .Target }})
	{{- else if eq .Type.Name "int" -}}
		{{ .VarName }} := strconv.Itoa({{ .Target }})
	{{- else if eq .Type.Name "int32" -}}
		{{ .VarName }} := strconv.FormatInt(int64({{ .Target }}), 10)
	{{- else if eq .Type.Name "int64" -}}
		{{ .VarName }} := strconv.FormatInt({{ .Target }}, 10)
	{{- else if eq .Type.Name "uint" -}}
		{{ .VarName }} := strconv.FormatUint(uint64({{ .Target }}), 10)
	{{- else if eq .Type.Name "uint32" -}}
		{{ .VarName }} := strconv.FormatUint(uint64({{ .Target }}), 10)
	{{- else if eq .Type.Name "uint64" -}}
		{{ .VarName }} := strconv.FormatUint({{ .Target }}, 10)
	{{- else if eq .Type.Name "float32" -}}
		{{ .VarName }} := strconv.FormatFloat(float64({{ .Target }}), 'f', -1, 32)
	{{- else if eq .Type.Name "float64" -}}
		{{ .VarName }} := strconv.FormatFloat({{ .Target }}, 'f', -1, 64)
	{{- else if eq .Type.Name "string" -}}
		{{ .VarName }} := {{ .Target }}
	{{- else if eq .Type.Name "bytes" -}}
		{{ .VarName }} := string({{ .Target }})
	{{- else if eq .Type.Name "any" -}}
		{{ .VarName }} := fmt.Sprintf("%v", {{ .Target }})
	{{- else }}
		// unsupported type {{ .Type.Name }} for field {{ .FieldName }}
	{{- end }}
{{- end }}
`
