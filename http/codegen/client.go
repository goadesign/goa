package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/codegen/service"
	"goa.design/goa/expr"
)

// ClientFiles returns the client HTTP transport files.
func ClientFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, 2*len(root.API.HTTP.Services))
	for i, r := range root.API.HTTP.Services {
		fw[i] = client(genpkg, r)
	}
	for i, r := range root.API.HTTP.Services {
		fw[i+len(root.API.HTTP.Services)] = clientEncodeDecode(genpkg, r)
	}
	return fw
}

// client returns the client HTTP transport file
func client(genpkg string, svc *expr.HTTPServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "client", "client.go")
	data := HTTPServices.Get(svc.Name())
	title := fmt.Sprintf("%s client HTTP transport", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "mime/multipart"},
			{Path: "net/http"},
			{Path: "strconv"},
			{Path: "strings"},
			{Path: "sync"},
			{Path: "time"},
			{Path: "github.com/gorilla/websocket"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: "goa.design/goa/http", Name: "goahttp"},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: data.Service.PkgName},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()) + "/" + "views", Name: data.Service.ViewsPkg},
		}),
	}
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "client-struct",
		Source: clientStructT,
		Data:   data,
		FuncMap: map[string]interface{}{
			"streamingEndpointExists": streamingEndpointExists,
		},
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

	for _, e := range data.Endpoints {
		if e.MultipartRequestEncoder != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "multipart-request-encoder-type",
				Source: multipartRequestEncoderTypeT,
				Data:   e.MultipartRequestEncoder,
			})
		}
	}

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "client-init",
		Source: clientInitT,
		Data:   data,
		FuncMap: map[string]interface{}{
			"streamingEndpointExists": streamingEndpointExists,
		},
	})

	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-endpoint-init",
			Source: endpointInitT,
			Data:   e,
		})
		if e.ClientStream != nil {
			if e.ClientStream.RecvTypeRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-recv",
					Source: streamRecvT,
					Data:   e.ClientStream,
					FuncMap: map[string]interface{}{
						"upgradeParams": upgradeParams,
					},
				})
			}
			switch e.ClientStream.Kind {
			case expr.ClientStreamKind, expr.BidirectionalStreamKind:
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-send",
					Source: streamSendT,
					Data:   e.ClientStream,
					FuncMap: map[string]interface{}{
						"upgradeParams":    upgradeParams,
						"viewedServerBody": viewedServerBody,
					},
				})
			}
			if e.ClientStream.MustClose {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-close",
					Source: streamCloseT,
					Data:   e.ClientStream,
					FuncMap: map[string]interface{}{
						"upgradeParams": upgradeParams,
					},
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

// clientEncodeDecode returns the file containing the HTTP client encoding and
// decoding logic.
func clientEncodeDecode(genpkg string, svc *expr.HTTPServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "client", "encode_decode.go")
	data := HTTPServices.Get(svc.Name())
	title := fmt.Sprintf("%s HTTP client encoders and decoders", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", []*codegen.ImportSpec{
			{Path: "bytes"},
			{Path: "context"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "io/ioutil"},
			{Path: "mime/multipart"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "strconv"},
			{Path: "strings"},
			{Path: "unicode/utf8"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: "goa.design/goa/http", Name: "goahttp"},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: data.Service.PkgName},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()) + "/" + "views", Name: data.Service.ViewsPkg},
		}),
	}

	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "request-builder",
			Source: requestBuilderT,
			Data:   e,
		})
		if e.RequestEncoder != "" {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "request-encoder",
				Source: requestEncoderT,
				FuncMap: map[string]interface{}{
					"typeConversionData": typeConversionData,
					"isBearer":           isBearer,
				},
				Data: e,
			})
		}
		if e.MultipartRequestEncoder != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "multipart-request-encoder",
				Source: multipartRequestEncoderT,
				Data:   e.MultipartRequestEncoder,
			})
		}
		if e.Result != nil || len(e.Errors) > 0 {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "response-decoder",
				Source: responseDecoderT,
				Data:   e,
			})
		}
	}
	for _, h := range data.ClientTransformHelpers {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-transform-helper",
			Source: transformHelperT,
			Data:   h,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// typeConversionData produces the template data suitable for executing the
// "header_conversion" template.
func typeConversionData(dt expr.DataType, varName string, target string) map[string]interface{} {
	return map[string]interface{}{
		"Type":    dt,
		"VarName": varName,
		"Target":  target,
	}
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
const clientStructT = `{{ printf "%s lists the %s service endpoint HTTP clients." .ClientStruct .Service.Name | comment }}
type {{ .ClientStruct }} struct {
	{{- range .Endpoints }}
	{{ printf "%s Doer is the HTTP client used to make requests to the %s endpoint." .Method.VarName .Method.Name | comment }}
	{{ .Method.VarName }}Doer goahttp.Doer
	{{ end }}
	// RestoreResponseBody controls whether the response bodies are reset after
	// decoding so they can be read again.
	RestoreResponseBody bool

	scheme     string
	host       string
	encoder    func(*http.Request) goahttp.Encoder
	decoder    func(*http.Response) goahttp.Decoder
	{{- if streamingEndpointExists . }}
	dialer goahttp.Dialer
	connConfigFn goahttp.ConnConfigureFunc
	{{- end }}
}
`

// input: ServiceData
const clientInitT = `{{ printf "New%s instantiates HTTP clients for all the %s service servers." .ClientStruct .Service.Name | comment }}
func New{{ .ClientStruct }}(
	scheme string,
	host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restoreBody bool,
	{{- if streamingEndpointExists . }}
	dialer goahttp.Dialer,
	connConfigFn goahttp.ConnConfigureFunc,
	{{- end }}
) *{{ .ClientStruct }} {
	return &{{ .ClientStruct }}{
		{{- range .Endpoints }}
		{{ .Method.VarName }}Doer: doer,
		{{- end }}
		RestoreResponseBody: restoreBody,
		scheme:            scheme,
		host:              host,
		decoder:           dec,
		encoder:           enc,
		{{- if streamingEndpointExists . }}
		dialer: dialer,
		connConfigFn: connConfigFn,
		{{- end }}
	}
}
`

// input: EndpointData
const endpointInitT = `{{ printf "%s returns an endpoint that makes HTTP requests to the %s service %s server." .EndpointInit .ServiceName .Method.Name | comment }}
func (c *{{ .ClientStruct }}) {{ .EndpointInit }}({{ if .MultipartRequestEncoder }}{{ .MultipartRequestEncoder.VarName }} {{ .MultipartRequestEncoder.FuncName }}{{ end }}) goa.Endpoint {
	var (
		{{- if and .ClientStream .RequestEncoder }}
		encodeRequest  = {{ .RequestEncoder }}({{ if .MultipartRequestEncoder }}{{ .MultipartRequestEncoder.InitName }}({{ .MultipartRequestEncoder.VarName }}){{ else }}c.encoder{{ end }})
		{{- else }}
			{{- if .RequestEncoder }}
		encodeRequest  = {{ .RequestEncoder }}({{ if .MultipartRequestEncoder }}{{ .MultipartRequestEncoder.InitName }}({{ .MultipartRequestEncoder.VarName }}){{ else }}c.encoder{{ end }})
			{{- end }}
		{{- end }}
		decodeResponse = {{ .ResponseDecoder }}(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.{{ .RequestInit.Name }}(ctx, {{ range .RequestInit.ClientArgs }}{{ .Ref }}{{ end }})
		if err != nil {
			return nil, err
		}
	{{- if .RequestEncoder }}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
	{{- end }}

	{{- if .ClientStream }}
		conn, resp, err := c.dialer.Dial(req.URL.String(), req.Header)
		if err != nil {
			if resp != nil {
				return decodeResponse(resp)
			}
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		if c.connConfigFn != nil {
			conn = c.connConfigFn(conn)
		}
		stream := &{{ .ClientStream.VarName }}{conn: conn}
		{{- if .Method.ViewedResult }}
			{{- if not .Method.ViewedResult.ViewName }}
		view := resp.Header.Get("goa-view")
		stream.SetView(view)
			{{- end }}
		{{- end }}
		return stream, nil
	{{- else }}
		resp, err := c.{{ .Method.VarName }}Doer.Do(req)

		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		return decodeResponse(resp)
	{{- end }}
	}
}
`

// input: EndpointData
const requestBuilderT = `{{ comment .RequestInit.Description }}
func (c *{{ .ClientStruct }}) {{ .RequestInit.Name }}(ctx context.Context, {{ range .RequestInit.ClientArgs }}{{ .Name }} {{ .TypeRef }}{{ end }}) (*http.Request, error) {
	{{- .RequestInit.ClientCode }}
}
`

// input: EndpointData
const requestEncoderT = `{{ printf "%s returns an encoder for requests sent to the %s %s server." .RequestEncoder .ServiceName .Method.Name | comment }}
func {{ .RequestEncoder }}(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.({{ .Payload.Ref }})
		if !ok {
			return goahttp.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .Payload.Ref }}", v)
		}
	{{- range .Payload.Request.Headers }}
		{{- if .FieldName }}
			{{- if .Pointer }}
		if p.{{ .FieldName }} != nil {
			{{- end }}
			{{- if (and (eq .Name "Authorization") (isBearer $.HeaderSchemes)) }}
		if !strings.Contains({{ if .Pointer }}*{{ end }}p.{{ .FieldName }}, " ") {
			req.Header.Set({{ printf "%q" .Name }}, "Bearer "+{{ if .Pointer }}*{{ end }}p.{{ .FieldName }})
		} else {
			{{- end }}
			req.Header.Set({{ printf "%q" .Name }}, {{ if .Pointer }}*{{ end }}p.{{ .FieldName }})
			{{- if (and (eq .Name "Authorization") (isBearer $.HeaderSchemes)) }}
		}
			{{- end }}
			{{- if .Pointer }}
		}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- if or .Payload.Request.QueryParams }}
		values := req.URL.Query()
	{{- end }}
	{{- range .Payload.Request.QueryParams }}
		{{- if .MapQueryParams }}
		for key, value := range p{{ if .FieldName }}.{{ .FieldName }}{{ end }} {
			{{ template "type_conversion" (typeConversionData .Type.KeyType.Type "keyStr" "key") }}
			{{- if eq .Type.ElemType.Type.Name "array" }}
			for _, val := range value {
				{{ template "type_conversion" (typeConversionData .Type.ElemType.Type.ElemType.Type "valStr" "val") }}
				values.Add(keyStr, valStr)
			}
			{{- else }}
			{{ template "type_conversion" (typeConversionData .Type.ElemType.Type "valueStr" "value") }}
			values.Add(keyStr, valueStr)
			{{- end }}
    }
		{{- else if .StringSlice }}
			for _, value := range p{{ if .FieldName }}.{{ .FieldName }}{{ end }} {
				values.Add("{{ .Name }}", value)
			}
		{{- else if .Slice }}
			for _, value := range p{{ if .FieldName }}.{{ .FieldName }}{{ end }} {
				{{ template "type_conversion" (typeConversionData .Type.ElemType.Type "valueStr" "value") }}
				values.Add("{{ .Name }}", valueStr)
			}
		{{- else if .FieldName }}
			{{- if .Pointer }}
		if p.{{ .FieldName }} != nil {
			{{- end }}
		values.Add("{{ .Name }}",
			{{- if eq .Type.Name "bytes" }} string(
			{{- else if not (eq .Type.Name "string") }} fmt.Sprintf("%v", 
			{{- end }}
			{{- if .Pointer }}*{{ end }}p.{{ .FieldName }}
			{{- if or (eq .Type.Name "bytes") (not (eq .Type.Name "string")) }})
			{{- end }})
			{{- if .Pointer }}
		}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- if .Payload.Request.QueryParams }}
		req.URL.RawQuery = values.Encode()
	{{- end }}
	{{- if .MultipartRequestEncoder }}
		if err := encoder(req).Encode(p); err != nil {
			return goahttp.ErrEncodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
	{{- else if .Payload.Request.ClientBody }}
		{{- if .Payload.Request.ClientBody.Init }}
		body := {{ .Payload.Request.ClientBody.Init.Name }}({{ range .Payload.Request.ClientBody.Init.ClientArgs }}{{ if .Pointer }}&{{ end }}{{ .Name }}, {{ end }})
		{{- else }}
		body := p
		{{- end }}
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
	{{- end }}
	{{- if .BasicScheme }}{{ with .BasicScheme }}
		{{- if not .UsernameRequired }}
		if p.{{ .UsernameField }} != nil {
		{{- end }}
		{{- if not .PasswordRequired }}
		if p.{{ .PasswordField }} != nil {
		{{- end }}
		req.SetBasicAuth({{ if .UsernamePointer }}*{{ end }}p.{{ .UsernameField }}, {{ if .PasswordPointer }}*{{ end }}p.{{ .PasswordField }})
		{{- if not .UsernameRequired }}
		}
		{{- end }}
		{{- if not .PasswordRequired }}
		}
		{{- end }}
	{{- end }}{{ end }}
		return nil
	}
}

{{- define "type_conversion" }}
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
    // unsupported type {{ .Type.Name }} for header field {{ .FieldName }}
  {{- end }}
{{- end }}
`

// input: EndpointData
const responseDecoderT = `{{ printf "%s returns a decoder for responses returned by the %s %s endpoint. restoreBody controls whether the response body should be restored after having been read." .ResponseDecoder .ServiceName .Method.Name | comment }}
{{- if .Errors }}
{{ printf "%s may return the following errors:" .ResponseDecoder | comment }}
	{{- range $gerr := .Errors }}
	{{- range $errors := .Errors }}
//	- {{ printf "%q" .Name }} (type {{ .Ref }}): {{ .Response.StatusCode }}{{ if .Response.Description }}, {{ .Response.Description }}{{ end }}
	{{- end }}
	{{- end }}
//	- error: internal error
{{- end }}
func {{ .ResponseDecoder }}(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
	{{- range .Result.Responses }}
		case {{ .StatusCode }}:
` + singleResponseT + `
		{{- if .ResultInit }}
			{{- if .ViewedResult }}
			p := {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
				{{- if $.Method.ViewedResult.ViewName }}
			view := {{ printf "%q" $.Method.ViewedResult.ViewName }}
				{{- else }}
			view := resp.Header.Get("goa-view")
				{{- end }}
			vres := {{ if not $.Method.ViewedResult.IsCollection }}&{{ end }}{{ $.Method.ViewedResult.ViewsPkg}}.{{ $.Method.ViewedResult.VarName }}{p, view}
			{{- if .ClientBody }}
				if err = {{ $.Method.ViewedResult.ViewsPkg}}.Validate{{ $.Method.Result }}(vres); err != nil {
					return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
				}
			{{- end }}
			res := {{ $.ServicePkgName }}.{{ $.Method.ViewedResult.ResultInit.Name }}(vres)
			{{- else }}
			res := {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
			{{- end }}
			{{- if .TagName }}
				{{- if .TagPointer }}
					tmp := {{ printf "%q" .TagValue }}
					res.{{ .TagName }} = &tmp
				{{- else }}
					res.{{ .TagName }} = {{ printf "%q" .TagValue }}
				{{- end }}
			{{- end }}
			return res, nil
		{{- else if .ClientBody }}
			return body, nil
		{{- else }}
			return nil, nil
		{{- end }}
	{{- end }}
	{{- range .Errors }}
		case {{ .StatusCode }}:
		{{- if gt (len .Errors) 1 }}
		en := resp.Header.Get("goa-error")
		switch en {
			{{- range .Errors }}
		case {{ printf "%q" .Name }}:
				{{- with .Response }}
` + singleResponseT + `
					{{- if .ResultInit }}
			return nil, {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
					{{- else if .ClientBody }}
			return nil, body
					{{- else }}
			return nil, nil
					{{- end }}
				{{- end }}
			{{- end }}
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse({{ printf "%q" $.ServiceName }}, {{ printf "%q" $.Method.Name }}, resp.StatusCode, string(body))
		}
		{{- else }}
			{{- with (index .Errors 0).Response }}
` + singleResponseT + `
				{{- if .ResultInit }}
			return nil, {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
				{{- else if .ClientBody }}
			return nil, body
				{{- else }}
			return nil, nil
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse({{ printf "%q" .ServiceName }}, {{ printf "%q" .Method.Name }}, resp.StatusCode, string(body))
		}
	}
}
`

// input: ResponseData
const singleResponseT = ` {{- if .ClientBody }}
			var (
				body {{ .ClientBody.VarName }}
				err error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
			}
		{{- if .ClientBody.ValidateRef }}
			{{ .ClientBody.ValidateRef }}
			if err != nil {
				return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
			}
		{{- end }}
	{{- end }}

	{{- if .Headers }}
			var (
		{{- range .Headers }}
				{{ .VarName }} {{ .TypeRef }}
		{{- end }}
		{{- if not .ClientBody }}
			{{- if .MustValidate }}
				err error
			{{- end }}
		{{- end }}
			)
		{{- range .Headers }}

		{{- if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
			{{ .VarName }}Raw := resp.Header.Get("{{ .Name }}")
			{{- if .Required }}
				if {{ .VarName }}Raw == "" {
					err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
				}
				{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
			{{- else }}
				if {{ .VarName }}Raw != "" {
					{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
				}
				{{- if .DefaultValue }} else {
					{{ .VarName }} = {{ if eq .Type.Name "string" }}{{ printf "%q" .DefaultValue }}{{ else }}{{ printf "%#v" .DefaultValue }}{{ end }}
				}
				{{- end }}
			{{- end }}

		{{- else if .StringSlice }}
			{{ .VarName }} = resp.Header["{{ .CanonicalName }}"]
			{{ if .Required }}
			if {{ .VarName }} == nil {
				err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
			}
			{{- else if .DefaultValue }}
			if {{ .VarName }} == nil {
				{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
			}
			{{- end }}

		{{- else if .Slice }}
			{{ .VarName }}Raw := resp.Header["{{ .CanonicalName }}"]
				{{ if .Required }} if {{ .VarName }}Raw == nil {
				return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "header"))
			}
			{{- else if .DefaultValue }}
			if {{ .VarName }}Raw == nil {
				{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
			}
			{{- end }}

			{{- if .DefaultValue }}else {
			{{- else if not .Required }}
			if {{ .VarName }}Raw != nil {
			{{- end }}
				{{- template "slice_conversion" . }}
			{{- if or .DefaultValue (not .Required) }}
			}
			{{- end }}

		{{- else }}{{/* not string, not any and not slice */}}
			{{ .VarName }}Raw := resp.Header.Get("{{ .Name }}")
			{{- if .Required }}
			if {{ .VarName }}Raw == "" {
				return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "header"))
			}
			{{- else if .DefaultValue }}
			if {{ .VarName }}Raw == "" {
				{{ .VarName }} = {{ printf "%#v" .DefaultValue }}
			}
			{{- end }}

			{{- if .DefaultValue }}else {
				{{- else if not .Required }}
			if {{ .VarName }}Raw != "" {
			{{- end }}
				{{- template "type_conversion" . }}
			{{- if or .DefaultValue (not .Required) }}
			}
			{{- end }}
		{{- end }}
		{{- if .Validate }}
			{{ .Validate }}
		{{- end }}
		{{- end }}{{/* range .Headers */}}
	{{- end }}

	{{- if .MustValidate }}
			if err != nil {
				return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
			}
	{{- end }}
`

// input: multipartData
const multipartRequestEncoderTypeT = `{{ printf "%s is the type to encode multipart request for the %q service %q endpoint." .FuncName .ServiceName .MethodName | comment }}
type {{ .FuncName }} func(*multipart.Writer, {{ .Payload.Ref }}) error
`

// input: multipartData
const multipartRequestEncoderT = `{{ printf "%s returns an encoder to encode the multipart request for the %q service %q endpoint." .InitName .ServiceName .MethodName | comment }}
func {{ .InitName }}(encoderFn {{ .FuncName }}) func(r *http.Request) goahttp.Encoder {
	return func(r *http.Request) goahttp.Encoder {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		return goahttp.EncodingFunc(func(v interface{}) error {
			p := v.({{ .Payload.Ref }})
			if err := encoderFn(mw, p); err != nil {
				return err
			}
			r.Body = ioutil.NopCloser(body)
			r.Header.Set("Content-Type", mw.FormDataContentType())
			return mw.Close()
		})
	}
}
`

// input: EndpointData
const clientStreamRecvT = `{{ printf "Recv receives a %s type from the %q endpoint websocket connection." .Result.Name .Method.Name | comment }}
func (c *{{ .ClientStream.VarName }}) Recv() ({{ .Result.Ref }}, error) {
	{{- if .Method.ViewedResult }}
	var vres {{ .Method.ViewedResult.ViewsPkg }}.{{ .Result.Name }}
	{{- else }}
	var res {{ .Result.Name }}
	{{- end }}
	err := c.conn.ReadJSON(&{{ if .Method.ViewedResult }}vres{{ else }}res{{ end }})
	if websocket.IsCloseError(err, goahttp.NormalSocketCloseErrors...) {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}
	{{- if .Method.ViewedResult }}
	if err := {{ .Method.ViewedResult.ViewsPkg }}.Validate{{ .Result.Name }}(vres); err != nil {
		return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
	return {{ $.ServicePkgName }}.{{ .Method.ViewedResult.ResultInit.Name }}(vres), nil
	{{- else }}
	return &res, nil
	{{- end }}
}
`
