package codegen

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
	httpdesign "goa.design/goa/http/design"
)

// ClientFiles returns the client HTTP transport files.
func ClientFiles(genpkg string, root *httpdesign.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, 2*len(root.HTTPServices))
	for i, r := range root.HTTPServices {
		fw[i] = client(genpkg, r)
	}
	for i, r := range root.HTTPServices {
		fw[i+len(root.HTTPServices)] = clientEncodeDecode(genpkg, r)
	}
	return fw
}

// client returns the client HTTP transport file
func client(genpkg string, svc *httpdesign.ServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "client", "client.go")
	data := HTTPServices.Get(svc.Name())
	title := fmt.Sprintf("%s client HTTP transport", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "net/http"},
			{Path: "strconv"},
			{Path: "strings"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: "goa.design/goa/http", Name: "goahttp"},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: data.Service.PkgName},
		}),
		{Name: "client-struct", Source: clientStructT, Data: data},
		{Name: "client-init", Source: clientInitT, Data: data},
	}
	for _, e := range data.Endpoints {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "client-endpoint-init",
			Source: endpointInitT,
			Data:   e,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

// clientEncodeDecode returns the file containing the HTTP client encoding and
// decoding logic.
func clientEncodeDecode(genpkg string, svc *httpdesign.ServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(svc.Name()), "client", "encode_decode.go")
	data := HTTPServices.Get(svc.Name())
	title := fmt.Sprintf("%s HTTP client encoders and decoders", svc.Name())
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", []*codegen.ImportSpec{
			{Path: "bytes"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "io/ioutil"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "strconv"},
			{Path: "strings"},
			{Path: "unicode/utf8"},
			{Path: "goa.design/goa", Name: "goa"},
			{Path: "goa.design/goa/http", Name: "goahttp"},
			{Path: genpkg + "/" + codegen.SnakeCase(svc.Name()), Name: data.Service.PkgName},
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
				},
				Data: e,
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
func typeConversionData(dt design.DataType, varName string, target string) map[string]interface{} {
	return map[string]interface{}{
		"Type":    dt,
		"VarName": varName,
		"Target":  target,
	}
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
	}
}
`

// input: EndpointData
const endpointInitT = `{{ printf "%s returns a endpoint that makes HTTP requests to the %s service %s server." .EndpointInit .ServiceName .Method.Name | comment }}
func (c *{{ .ClientStruct }}) {{ .EndpointInit }}() goa.Endpoint {
	var (
		{{- if .RequestEncoder }}
		encodeRequest  = {{ .RequestEncoder }}(c.encoder)
		{{- end }}
		decodeResponse = {{ .ResponseDecoder }}(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.{{ .RequestInit.Name }}({{ range .RequestInit.ClientArgs }}{{ .Ref }}{{ end }})
		if err != nil {
			return nil, err
		}
		{{- if .RequestEncoder }}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
		{{- end }}

		resp, err := c.{{ .Method.VarName }}Doer.Do(req)

		if err != nil {
			return nil, goahttp.ErrRequestError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		return decodeResponse(resp)
	}
}
`

// input: EndpointData
const requestBuilderT = `{{ comment .RequestInit.Description }}
func (c *{{ .ClientStruct }}) {{ .RequestInit.Name }}({{ range .RequestInit.ClientArgs }}{{ .Name }} {{ .TypeRef }}{{ end }}) (*http.Request, error) {
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
		req.Header.Set("{{ .Name }}", {{ if .Pointer }}*{{ end }}p.{{ .FieldName }})
			{{- if .Pointer }}
		}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- if or .Payload.Request.QueryParams }}
		values := req.URL.Query()
	{{- end }}
	{{- range .Payload.Request.QueryParams }}
		{{- if .FieldName }}
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
		{{- if .MapQueryParams }}
		for key, value := range p {
			{{ template "type_conversion" (typeConversionData .Type.KeyType.Type (printf "%sStr" "key") "key") }}
			{{- if eq .Type.ElemType.Type.Name "array" }}
			for _, val := range value {
				{{ template "type_conversion" (typeConversionData .Type.ElemType.Type.ElemType.Type (printf "%sStr" "val") "val") }}
				values.Add(keyStr, valStr)
			}
			{{- else }}
			{{ template "type_conversion" (typeConversionData .Type.ElemType.Type (printf "%sStr" "value") "value") }}
      values.Add(keyStr, valueStr)
			{{- end }}
    }
		{{- end }}
	{{- end }}
	{{- if or .Payload.Request.QueryParams }}
		req.URL.RawQuery = values.Encode()
	{{- end }}
	{{- if .Payload.Request.ClientBody }}
		{{- if .Payload.Request.ClientBody.Init }}
		body := {{ .Payload.Request.ClientBody.Init.Name }}({{ range .Payload.Request.ClientBody.Init.ClientArgs }}{{ if .Pointer }}&{{ end }}{{ .Name }}, {{ end }})
		{{- else }}
		body := p
		{{- end }}
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
	{{- end }}
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
{{ printf "%s may return the following error types:" .ResponseDecoder | comment }}
	{{- range $errors := .Errors }}
//	- {{ .Ref }}: {{ .Response.StatusCode }}{{ if .Response.Description }}, {{ .Response.Description }}{{ end }}
	{{- end }}
//	- error: generic transport error.
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
` + singleResponseT + `
		{{- if .ResultInit }}
			return {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }}), nil
		{{- else if .ClientBody }}
			return body, nil
		{{- else }}
			return nil, nil
		{{- end }}
	{{- end }}
	{{- range .Errors }}
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
			return nil, goahttp.ErrInvalidResponse("account", "create", resp.StatusCode, string(body))
		}
	}
}
`

const singleResponseT = `case {{ .StatusCode }}:
	{{- if .ClientBody }}
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
				return nil, fmt.Errorf("invalid response: %s", err)
			}
		{{- end }}
	{{ end }}

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

		{{- if and (or (eq .Type.Name "string") (eq .Type.Name "any")) .Required }}
			{{ .VarName }} = resp.Header.Get("{{ .Name }}")
			if {{ .VarName }} != "" {
				err = goa.MergeErrors(err, goa.MissingFieldError("{{ .Name }}", "header"))
			}

		{{- else if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
			{{ .VarName }}Raw := resp.Header.Get("{{ .Name }}")
			if {{ .VarName }}Raw != "" {
				{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
			}
			{{- if .DefaultValue }} else {
				{{ .VarName }} = {{ if eq .Type.Name "string" }}{{ printf "%q" .DefaultValue }}{{ else }}{{ printf "%#v" .DefaultValue }}{{ end }}
			}
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
				return nil, fmt.Errorf("invalid response: %s", goa.MissingFieldError("{{ .Name }}", "header"))
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
				return nil, fmt.Errorf("invalid response: %s", goa.MissingFieldError("{{ .Name }}", "header"))
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
				return nil, fmt.Errorf("invalid response: %s", err)
			}
	{{- end }}
`
