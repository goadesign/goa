package codegen

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service"
	"goa.design/goa/design"
	httpdesign "goa.design/goa/http/design"
)

// ServerFiles returns all the server HTTP transport files.
func ServerFiles(genpkg string, root *httpdesign.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, 2*len(root.HTTPServices))
	for i, svc := range root.HTTPServices {
		fw[i] = server(genpkg, svc)
	}
	for i, r := range root.HTTPServices {
		fw[i+len(root.HTTPServices)] = serverEncodeDecode(genpkg, r)
	}
	return fw
}

// server returns the files defining the HTTP server.
func server(genpkg string, svc *httpdesign.ServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "server", "server.go")
	data := HTTPServices.Get(svc.Name())
	title := fmt.Sprintf("%s HTTP server", svc.Name())
	funcs := map[string]interface{}{"join": func(ss []string, s string) string { return strings.Join(ss, s) }}
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "mime/multipart"},
			{Path: "net/http"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: "goa.design/goa/http", Name: "goahttp"},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: data.Service.PkgName},
		}),
	}

	sections = append(sections, &codegen.SectionTemplate{Name: "server-struct", Source: serverStructT, Data: data})
	sections = append(sections, &codegen.SectionTemplate{Name: "server-mountpoint", Source: mountPointStructT, Data: data})

	for _, e := range data.Endpoints {
		if e.MultipartRequestDecoder != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "multipart-request-decoder-type",
				Source: multipartRequestDecoderTypeT,
				Data:   e.MultipartRequestDecoder,
			})
		}
	}

	sections = append(sections, &codegen.SectionTemplate{Name: "server-init", Source: serverInitT, Data: data})
	sections = append(sections, &codegen.SectionTemplate{Name: "server-service", Source: serverServiceT, Data: data})
	sections = append(sections, &codegen.SectionTemplate{Name: "server-use", Source: serverUseT, Data: data})
	sections = append(sections, &codegen.SectionTemplate{Name: "server-mount", Source: serverMountT, Data: data})

	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{Name: "server-handler", Source: serverHandlerT, Data: e})
		sections = append(sections, &codegen.SectionTemplate{Name: "server-handler-init", Source: serverHandlerInitT, Data: e})
	}
	for _, s := range data.FileServers {
		sections = append(sections, &codegen.SectionTemplate{Name: "server-files", Source: fileServerT, FuncMap: funcs, Data: s})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// serverEncodeDecode returns the file defining the HTTP server encoding and
// decoding logic.
func serverEncodeDecode(genpkg string, svc *httpdesign.ServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "server", "encode_decode.go")
	data := HTTPServices.Get(svc.Name())
	title := fmt.Sprintf("%s HTTP server encoders and decoders", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "net/http"},
			{Path: "strconv"},
			{Path: "strings"},
			{Path: "encoding/json"},
			{Path: "mime/multipart"},
			{Path: "unicode/utf8"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: "goa.design/goa/http", Name: "goahttp"},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: data.Service.PkgName},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()) + "/" + "views", Name: data.Service.ViewsPkg},
		}),
	}

	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "response-encoder",
			FuncMap: transTmplFuncs(svc),
			Source:  responseEncoderT,
			Data:    e,
		})
		if e.Payload.Ref != "" {
			sections = append(sections, &codegen.SectionTemplate{
				Name:    "request-decoder",
				Source:  requestDecoderT,
				FuncMap: transTmplFuncs(svc),
				Data:    e,
			})
		}
		if e.MultipartRequestDecoder != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:    "multipart-request-decoder",
				Source:  multipartRequestDecoderT,
				FuncMap: transTmplFuncs(svc),
				Data:    e.MultipartRequestDecoder,
			})
		}

		if len(e.Errors) > 0 {
			sections = append(sections, &codegen.SectionTemplate{
				Name:    "error-encoder",
				Source:  errorEncoderT,
				FuncMap: transTmplFuncs(svc),
				Data:    e,
			})
		}
	}
	for _, h := range data.ServerTransformHelpers {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-transform-helper",
			Source: transformHelperT,
			Data:   h,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func transTmplFuncs(s *httpdesign.ServiceExpr) map[string]interface{} {
	return map[string]interface{}{
		"goTypeRef": func(dt design.DataType) string {
			return service.Services.Get(s.Name()).Scope.GoTypeRef(&design.AttributeExpr{Type: dt})
		},
		"conversionData":       conversionData,
		"headerConversionData": headerConversionData,
		"printValue":           printValue,
	}
}

// conversionData creates a template context suitable for executing the
// "type_conversion" template.
func conversionData(varName, name string, dt design.DataType) map[string]interface{} {
	return map[string]interface{}{
		"VarName": varName,
		"Name":    name,
		"Type":    dt,
	}
}

// headerConversionData produces the template data suitable for executing the
// "header_conversion" template.
func headerConversionData(dt design.DataType, varName string, target string) map[string]interface{} {
	return map[string]interface{}{
		"Type":    dt,
		"VarName": varName,
		"Target":  target,
	}
}

// printValue generates the Go code for a literal string containing the given
// value. printValue panics if the data type is not a primitive or an array.
func printValue(dt design.DataType, v interface{}) string {
	switch actual := dt.(type) {
	case *design.Array:
		val := reflect.ValueOf(v)
		elems := make([]string, val.Len())
		for i := 0; i < val.Len(); i++ {
			elems[i] = printValue(actual.ElemType.Type, val.Index(i).Interface())
		}
		return strings.Join(elems, ", ")
	case design.Primitive:
		return fmt.Sprintf("%v", v)
	default:
		panic("unsupported type value " + dt.Name()) // bug
	}
}

// input: ServiceData
const serverStructT = `{{ printf "%s lists the %s service endpoint HTTP handlers." .ServerStruct .Service.Name | comment }}
type {{ .ServerStruct }} struct {
	Mounts []*{{ .MountPointStruct }}
	{{- range .Endpoints }}
	{{ .Method.VarName }} http.Handler
	{{- end }}
}

// ErrorNamer is an interface implemented by generated error structs that
// exposes the name of the error as defined in the design.
type ErrorNamer interface {
	ErrorName() string
}
`

// input: ServiceData
const mountPointStructT = `{{ printf "%s holds information about the mounted endpoints." .MountPointStruct | comment }}
type {{ .MountPointStruct }} struct {
	{{ printf "Method is the name of the service method served by the mounted HTTP handler." | comment }}
	Method string
	{{ printf "Verb is the HTTP method used to match requests to the mounted handler." | comment }}
	Verb string
	{{ printf "Pattern is the HTTP request path pattern used to match requests to the mounted handler." | comment }}
	Pattern string
}
`

// input: ServiceData
const serverInitT = `{{ printf "%s instantiates HTTP handlers for all the %s service endpoints." .ServerInit .Service.Name | comment }}
func {{ .ServerInit }}(
	e *{{ .Service.PkgName }}.Endpoints,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
	{{- range .Endpoints }}
		{{- if .MultipartRequestDecoder }}
	{{ .MultipartRequestDecoder.VarName }} {{ .MultipartRequestDecoder.FuncName }},
		{{- end }}
	{{- end }}
) *{{ .ServerStruct }} {
	return &{{ .ServerStruct }}{
		Mounts: []*{{ .MountPointStruct }}{
			{{- range $e := .Endpoints }}
				{{- range $e.Routes }}
			{"{{ $e.Method.VarName }}", "{{ .Verb }}", "{{ .Path }}"},
				{{- end }}
			{{- end }}
			{{- range .FileServers }}
				{{- $filepath := .FilePath }}
				{{- range .RequestPaths }}
			{"{{ $filepath }}", "GET", "{{ . }}"},
				{{- end }}
			{{- end }}
		},
		{{- range .Endpoints }}
		{{ .Method.VarName }}: {{ .HandlerInit }}(e.{{ .Method.VarName }}, mux, {{ if .MultipartRequestDecoder }}{{ .MultipartRequestDecoder.InitName }}(mux, {{ .MultipartRequestDecoder.VarName }}){{ else }}dec{{ end }}, enc, eh),
		{{- end }}
	}
}
`

// input: ServiceData
const serverServiceT = `{{ printf "%s returns the name of the service served." .ServerService | comment }}
func (s *{{ .ServerStruct }}) {{ .ServerService }}() string { return "{{ .Service.Name }}" }
`

// input: ServiceData
const serverUseT = `{{ printf "Use wraps the server handlers with the given middleware." | comment }}
func (s *{{ .ServerStruct }}) Use(m func(http.Handler) http.Handler) { 
{{- range .Endpoints }}
	s.{{ .Method.VarName }} = m(s.{{ .Method.VarName }})
{{- end }}
}
`

// input: ServiceData
const serverMountT = `{{ printf "%s configures the mux to serve the %s endpoints." .MountServer .Service.Name | comment }}
func {{ .MountServer }}(mux goahttp.Muxer{{ if .Endpoints }}, h *{{ .ServerStruct }}{{ end }}) {
	{{- range .Endpoints }}
	{{ .MountHandler }}(mux, h.{{ .Method.VarName }})
	{{- end }}
	{{- range .FileServers }}
		{{- if .IsDir }}
	{{ .MountHandler }}(mux, http.FileServer(http.Dir({{ printf "%q" .FilePath }})))
		{{- else }}
	{{ .MountHandler }}(mux, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, {{ printf "%q" .FilePath }})
		}))
		{{- end }}
	{{- end }}
}
`

// input: EndpointData
const serverHandlerT = `{{ printf "%s configures the mux to serve the %q service %q endpoint." .MountHandler .ServiceName .Method.Name | comment }}
func {{ .MountHandler }}(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	{{- range .Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", f)
	{{- end }}
}
`

// input: FileServerData
const fileServerT = `{{ printf "%s configures the mux to serve GET request made to %q." .MountHandler (join .RequestPaths ", ") | comment }}
func {{ .MountHandler }}(mux goahttp.Muxer, h http.Handler) {
	{{- range .RequestPaths }}
	mux.Handle("GET", "{{ . }}", h.ServeHTTP)
	{{- end }}
}
`

// input: EndpointData
const serverHandlerInitT = `{{ printf "%s creates a HTTP handler which loads the HTTP request and calls the %q service %q endpoint." .HandlerInit .ServiceName .Method.Name | comment }}
func {{ .HandlerInit }}(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	dec func(*http.Request) goahttp.Decoder,
	enc func(context.Context, http.ResponseWriter) goahttp.Encoder,
	eh func(context.Context, http.ResponseWriter, error),
) http.Handler {
	var (
		{{- if .Payload.Ref }}
		decodeRequest  = {{ .RequestDecoder }}(mux, dec)
		{{- end }}
		encodeResponse = {{ .ResponseEncoder }}(enc)
		encodeError    = {{ if .Errors }}{{ .ErrorEncoder }}{{ else }}goahttp.ErrorEncoder{{ end }}(enc)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, {{ printf "%q" .Method.Name }})
		ctx = context.WithValue(ctx, goa.ServiceKey, {{ printf "%q" .ServiceName }})

		{{- if .Payload.Ref }}
		payload, err := decodeRequest(r)
		if err != nil {
			eh(ctx, w, err)
			return
		}

		res, err := endpoint(ctx, payload)
		{{- else }}
		res, err := endpoint(ctx, nil)
		{{- end }}

		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				eh(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			eh(ctx, w, err)
		}
	})
}
`

// input: TransformFunctionData
const transformHelperT = `{{ printf "%s builds a value of type %s from a value of type %s." .Name .ResultTypeRef .ParamTypeRef | comment }}
func {{ .Name }}(v {{ .ParamTypeRef }}) {{ .ResultTypeRef }} {
	{{ .Code }}
	return res
}
`

// input: EndpointData
const requestDecoderT = `{{ printf "%s returns a decoder for requests sent to the %s %s endpoint." .RequestDecoder .ServiceName .Method.Name | comment }}
func {{ .RequestDecoder }}(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
{{- if .MultipartRequestDecoder }}
		var payload {{ .Payload.Ref }}
		if err := decoder(r).Decode(&payload); err != nil {
			return nil, goa.DecodePayloadError(err.Error())
		}
{{- else if .Payload.Request.ServerBody }}
		var (
			body {{ .Payload.Request.ServerBody.VarName }}
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		{{- if .Payload.Request.ServerBody.ValidateRef }}
		{{ .Payload.Request.ServerBody.ValidateRef }}
		if err != nil {
			return nil, err
		}
		{{- end }}
{{- end }}
{{- if not .MultipartRequestDecoder }}
	{{- template "request_params_headers" .Payload.Request }}
	{{- if .Payload.Request.MustValidate }}
		if err != nil {
			return nil, err
		}
	{{- end }}
	{{- if .Payload.Request.PayloadInit }}
	payload := {{ .Payload.Request.PayloadInit.Name }}({{ range .Payload.Request.PayloadInit.ServerArgs }}{{ .Ref }}, {{ end }})
	{{- else if .Payload.DecoderReturnValue }}
	payload := {{ .Payload.DecoderReturnValue }}
	{{- else }}
	payload := body
	{{- end }}
{{- end }}
{{- if .BasicScheme }}{{ with .BasicScheme }}
	user, pass, {{ if or .UsernameRequired .PasswordRequired }}ok{{ else }}_{{ end }} := r.BasicAuth()
		{{- if or .UsernameRequired .PasswordRequired}}
	if !ok {
		return nil, goa.MissingFieldError("Authorization", "header")
	}
		{{- end }}
	payload.{{ .UsernameField }} = {{ if .UsernamePointer }}&{{ end }}user
	payload.{{ .PasswordField }} = {{ if .PasswordPointer }}&{{ end }}pass
{{- end }}{{ end }}
{{- range .HeaderSchemes }}
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

	return payload, nil
	}
}
` + requestParamsHeadersT

// input: RequestData
const requestParamsHeadersT = `{{- define "request_params_headers" }}
{{- if or .PathParams .QueryParams .Headers }}
{{- if .ServerBody }}{{/* we want a newline only if there was code before */}}
{{ end }}
		var (
		{{- range .PathParams }}
			{{ .VarName }} {{ .TypeRef }}
		{{- end }}
		{{- range .QueryParams }}
			{{ .VarName }} {{ .TypeRef }}
		{{- end }}
		{{- range .Headers }}
			{{ .VarName }} {{ .TypeRef }}
		{{- end }}
		{{- if not .ServerBody }}
		{{- if .MustValidate }}
			err error
		{{- end }}
		{{- end }}
		{{- if .PathParams }}

			params = mux.Vars(r)
		{{- end }}
		)

{{- range .PathParams }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }} = params["{{ .Name }}"]

	{{- else }}{{/* not string and not any */}}
		{
			{{ .VarName }}Raw := params["{{ .Name }}"]
			{{- template "path_conversion" . }}
		}

	{{- end }}
		{{- if .Validate }}
		{{ .Validate }}
		{{- end }}
{{- end }}

{{- range .QueryParams }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) .Required }}
		{{ .VarName }} = r.URL.Query().Get("{{ .Name }}")
		if {{ .VarName }} == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}

	{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }}Raw := r.URL.Query().Get("{{ .Name }}")
		if {{ .VarName }}Raw != "" {
			{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
		}

	{{- else if .StringSlice }}
		{{ .VarName }} = r.URL.Query()["{{ .Name }}"]
		{{- if .Required }}
		if {{ .VarName }} == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- end }}

	{{- else if .Slice }}
	{
		{{ .VarName }}Raw := r.URL.Query()["{{ .Name }}"]
		{{- if .Required }}
		if {{ .VarName }}Raw == nil {
			return goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- end }}

		{{- if not .Required }}
		if {{ .VarName }}Raw != nil {
		{{- end }}
		{{- template "slice_conversion" . }}
		{{- if not .Required }}
		}
		{{- end }}
	}

	{{- else if .MapStringSlice }}
		{{ .VarName }} = r.URL.Query()
		{{- if .Required }}
		if len({{ .VarName }}) == 0 {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- end }}

	{{- else if .Map }}
	{
		{{ .VarName }}Raw := r.URL.Query()
		{{- if .Required }}
		if len({{ .VarName }}Raw) == 0 {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- end }}

		{{- if not .Required }}
		if len({{ .VarName }}Raw) != 0 {
		{{- end }}
		{{- if eq .Type.ElemType.Type.Name "array" }}
			{{- if eq .Type.ElemType.Type.ElemType.Type.Name "string" }}
			{{- template "map_key_conversion" . }}
			{{- else }}
			{{- template "map_slice_conversion" . }}
			{{- end }}
		{{- else }}
			{{- template "map_conversion" . }}
		{{- end }}
		{{- if not .Required }}
		}
		{{- end }}
	}

	{{- else if .MapQueryParams }}
	{
		{{ .VarName }}Raw := r.URL.Query()
		{{- if .Required }}
		if len({{ .VarName }}Raw) == 0 {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- end }}

		{{- if not .Required }}
		if len({{ .VarName }}Raw) != 0 {
		{{- end }}
		{{- if eq .Type.ElemType.Type.Name "array" }}
			{{- if eq .Type.ElemType.Type.ElemType.Type.Name "string" }}
			{{- template "map_key_conversion" . }}
			{{- else }}
			{{- template "map_slice_conversion" . }}
			{{- end }}
		{{- else }}
			{{- template "map_conversion" . }}
		{{- end }}
		{{- if not .Required }}
		}
		{{- end }}
	}

	{{- else }}{{/* not string, not any, not slice and not map */}}
	{
		{{ .VarName }}Raw := r.URL.Query().Get("{{ .Name }}")
		{{- if .Required }}
		if {{ .VarName }}Raw == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "query string"))
		}
		{{- end }}
		{{- if not .Required }}
		if {{ .VarName }}Raw != "" {
		{{- end }}
		{{- template "type_conversion" . }}
		{{- if not .Required }}
		}
		{{- end }}
	}

	{{- end }}
		{{- if .Validate }}
		{{ .Validate }}
		{{- end }}
{{- end }}

{{- range .Headers }}
	{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) .Required }}
		{{ .VarName }} = r.Header.Get("{{ .Name }}")
		if {{ .VarName }} == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
		}

	{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }}Raw := r.Header.Get("{{ .Name }}")
		if {{ .VarName }}Raw != "" {
			{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
		}

	{{- else if .StringSlice }}
		{{ .VarName }} = r.Header["{{ .CanonicalName }}"]
		{{- if .Required }}
		if {{ .VarName }} == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
		}
		{{- end }}

	{{- else if .Slice }}
	{
		{{ .VarName }}Raw := r.Header["{{ .CanonicalName }}"]
		{{ if .Required }}if {{ .VarName }}Raw == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
		}
		{{- end }}

		{{- if not .Required }}
		if {{ .VarName }}Raw != nil {
		{{- end }}
		{{- template "slice_conversion" . }}
		{{- if not .Required }}
		}
		{{- end }}
	}

	{{- else }}{{/* not string, not any and not slice */}}
	{
		{{ .VarName }}Raw := r.Header.Get("{{ .Name }}")
		{{- if .Required }}
		if {{ .VarName }}Raw == "" {
			err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
		}
		{{- end }}

		{{- if not .Required }}
		if {{ .VarName }}Raw != "" {
		{{- end }}
		{{- template "type_conversion" . }}
		{{- if not .Required }}
		}
		{{- end }}
	}
	{{- end }}
	{{- if .Validate }}
		{{ .Validate }}
	{{- end }}
{{- end }}
{{- end }}
{{- end }}

{{- define "path_conversion" }}
	{{- if eq .Type.Name "array" }}
		{{ .VarName }}RawSlice := strings.Split({{ .VarName }}Raw, ",")
		{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}RawSlice))
		for i, rv := range {{ .VarName }}RawSlice {
			{{- template "slice_item_conversion" . }}
		}
	{{- else }}
		{{- template "type_conversion" . }}
	{{- end }}
{{- end }}

{{- define "slice_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for i, rv := range {{ .VarName }}Raw {
		{{- template "slice_item_conversion" . }}
	}
{{- end }}

{{- define "map_key_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for keyRaw, val := range {{ .VarName }}Raw {
		var key {{ goTypeRef .Type.KeyType.Type }}
		{
		{{- template "type_conversion" (conversionData "key" (printf "%q" "query") .Type.KeyType.Type) }}
		}
		{{ .VarName }}[key] = val
	}
{{- end }}

{{- define "map_slice_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for key{{ if not (eq .Type.KeyType.Type.Name "string") }}Raw{{ end }}, valRaw := range {{ .VarName }}Raw {

		{{- if not (eq .Type.KeyType.Type.Name "string") }}
		var key {{ goTypeRef .Type.KeyType.Type }}
		{
			{{- template "type_conversion" (conversionData "key" (printf "%q" "query") .Type.KeyType.Type) }}
		}
		{{- end }}
		var val {{ goTypeRef .Type.ElemType.Type }}
		{
		{{- template "slice_conversion" (conversionData "val" (printf "%q" "query") .Type.ElemType.Type) }}
		}
		{{ .VarName }}[key] = val
	}
{{- end }}

{{- define "map_conversion" }}
	{{ .VarName }} = make({{ goTypeRef .Type }}, len({{ .VarName }}Raw))
	for key{{ if not (eq .Type.KeyType.Type.Name "string") }}Raw{{ end }}, va := range {{ .VarName }}Raw {

		{{- if not (eq .Type.KeyType.Type.Name "string") }}
		var key {{ goTypeRef .Type.KeyType.Type }}
		{
			{{- if eq .Type.KeyType.Type.Name "string" }}
			key = keyRaw
			{{- else }}
			{{- template "type_conversion" (conversionData "key" (printf "%q" "query") .Type.KeyType.Type) }}
			{{- end }}
		}
		{{- end }}
		var val {{ goTypeRef .Type.ElemType.Type }}
		{
			{{- if eq .Type.ElemType.Type.Name "string" }}
			val = va[0]
			{{- else }}
			valRaw := va[0]
			{{- template "type_conversion" (conversionData "val" (printf "%q" "query") .Type.ElemType.Type) }}
			{{- end }}
		}
		{{ .VarName }}[key] = val
	}
{{- end }}

{{- define "type_conversion" }}
	{{- if eq .Type.Name "bytes" }}
		{{ .VarName }} = []byte({{.VarName}}Raw)
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
`

// input: EndpointData
const responseEncoderT = `{{ printf "%s returns an encoder for responses returned by the %s %s endpoint." .ResponseEncoder .ServiceName .Method.Name | comment }}
func {{ .ResponseEncoder }}(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
	{{- if and .Result.Ref .NeedServerResponse }}
		res := v.({{ .Result.Ref }})
		{{- range .Result.Responses }}
			{{- if and .ServerBody $.Method.ViewedResult }}
			w.Header().Set("goa-view", res.View)
			{{- end }}
			{{- if .TagName }}
		if res.{{ .TagName }} != nil && *res.{{ .TagName }} == {{ printf "%q" .TagValue }} {
			{{- end -}}
			{{ template "response" . }}
			{{- if .ServerBody }}
			return enc.Encode(body)
			{{- else }}
			return nil
			{{- end }}

			{{- if .TagName }}
		}
			{{- end }}

		{{- end }}

	{{- else }}

		{{- with (index .Result.Responses 0) }}
		w.WriteHeader({{ .StatusCode }})
		return nil

		{{- end }}

	{{- end }}
	}
}
` + responseT

// input: EndpointData
const errorEncoderT = `{{ printf "%s returns an encoder for errors returned by the %s %s endpoint." .ErrorEncoder .Method.Name .ServiceName | comment }}
func {{ .ErrorEncoder }}(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
	{{- range $gerr := .Errors }}
	{{- range $err := .Errors }}
		case {{ printf "%q" .Name }}:
			res := v.({{ $err.Ref }})
			{{- with .Response}}
				{{- template "response" . }}
				{{- if .ServerBody }}
				return enc.Encode(body)
				{{- end }}
			{{- end }}
	{{- end }}
	{{- end }}
		default:
			return encodeError(ctx, w, v)
		}
	}
}
` + responseT

// input: ResponseData
const responseT = `{{ define "response" -}}
	{{- if .ServerBody }}
	enc := encoder(ctx, w)
		{{- if .ServerBody.Init }}
	body := {{ .ServerBody.Init.Name }}({{ range .ServerBody.Init.ServerArgs }}{{ .Ref }}, {{ end }})
		{{- else }}
	body := res
		{{- end }}
	{{- end }}
	{{- range .Headers }}
		{{- $initDef := and .DefaultValue (not $.TagName) }}
		{{- if not $.TagName }}
	if res.{{ .FieldName }} != nil {
		{{- end }}
		{{- if eq .Type.Name "string" }}
	w.Header().Set("{{ .Name }}", *res{{ if .FieldName }}.{{ .FieldName }}{{ end }})
		{{- else }}
	val := *res{{ if .FieldName }}.{{ .FieldName }}{{ end }}
	{{ template "header_conversion" (headerConversionData .Type (printf "%ss" .VarName) "val") }}
	w.Header().Set("{{ .Name }}", {{ .VarName }}s)
		{{- end }}
		{{- if $initDef }}
	} else {
		w.Header().Set("{{ .Name }}", "{{ printValue .Type .DefaultValue }}")
		{{- end }}
		{{- if not $.TagName }}
	}
		{{- end }}
	{{- end }}
	{{- if .ErrorHeader }}
	w.Header().Set("goa-error", {{ printf "%q" .ErrorHeader }})
	{{- end }}
	w.WriteHeader({{ .StatusCode }})
{{- end }}

{{- define "header_conversion" }}
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
	{{- else if eq .Type.Name "array" -}}
		{{- if eq .Type.ElemType.Type.Name "string" -}}
		{{ .VarName }} := strings.Join({{ .Target }}, ", ")
		{{- else -}}
		{{ .VarName }}Slice := make([]string, len({{ .Target }}))
		for i, e := range {{ .Target }}  {
			{{ template "header_conversion" (headerConversionData .Type.ElemType.Type "es" "e") }}
			{{ .VarName }}Slice[i] = es	
		}
		{{ .VarName }} := strings.Join({{ .VarName }}Slice, ", ")
		{{- end }}
	{{- else }}
		// unsupported type {{ .Type.Name }} for header field {{ .FieldName }}
	{{- end }}
{{- end -}}
`

// input: multipartData
const multipartRequestDecoderTypeT = `{{ printf "%s is the type to decode multipart request for the %q service %q endpoint." .FuncName .ServiceName .MethodName | comment }}
type {{ .FuncName }} func(*multipart.Reader, *{{ .Payload.Ref }}) error
`

// input: multipartData
const multipartRequestDecoderT = `{{ printf "%s returns a decoder to decode the multipart request for the %q service %q endpoint." .InitName .ServiceName .MethodName | comment }}
func {{ .InitName }}(mux goahttp.Muxer, {{ .VarName }} {{ .FuncName }}) func(r *http.Request) goahttp.Decoder {
	return func(r *http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(v interface{}) error {
			mr, err := r.MultipartReader()
			if err != nil {
				return err
			}
			p := v.(*{{ .Payload.Ref }})
			if err := {{ .VarName }}(mr, p); err != nil {
				return err
			}
			{{- template "request_params_headers" .Payload.Request }}
			{{- if .Payload.Request.MustValidate }}
			if err != nil {
				return err
			}
			{{- end }}
			{{- if .Payload.Request.PayloadInit }}
				{{- range .Payload.Request.PayloadInit.ServerArgs }}
					{{- if .FieldName }}
			(*p).{{ .FieldName }} = {{ if .Pointer }}&{{ end }}{{ .Name }}
					{{- end }}
				{{- end }}
			{{- end }}
			return nil
		})
	}
}
` + requestParamsHeadersT
