{{ if and .IsJSONRPC (not .Payload.Ref) }}{{ printf "%s returns an encoder for requests sent to the %s service %s JSON-RPC method." .RequestEncoderDeclaration.Name .ServiceName .Method.Name | comment }}{{ else }}{{ printf "%s returns an encoder for requests sent to the %s %s server." .RequestEncoderDeclaration.Name .ServiceName .Method.Name | comment }}{{ end }}
func {{ .RequestEncoderDeclaration.Name }}(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, any) error {
	return func(req *http.Request, v any) error {
{{- if and .IsJSONRPC (not .Payload.Ref) }}
		{{- template "partial_jsonrpc_request_envelope" . }}
		if err := encoder(req).Encode(body); err != nil {
			return goahttp.ErrEncodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
		}
		return nil
{{- else }}
		{{- if .Method.SkipRequestBodyEncodeDecode }}
		data, ok := v.(*{{ .ServicePkgName }}.{{ .Method.RequestStruct }})
		if !ok {
			return goahttp.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "*{{ .ServicePkgName }}.{{ .Method.RequestStruct }}", v)
		}
		p := data.Payload
		{{- else }}
		p, ok := v.({{ .Payload.Ref }})
		if !ok {
			return goahttp.ErrInvalidType("{{ .ServiceName }}", "{{ .Method.Name }}", "{{ .Payload.Ref }}", v)
		}
		{{- end }}
	{{- range .Payload.Request.Headers }}
		{{- if .FieldName }}
			{{- if .FieldPointer }}
		if p.{{ .FieldName }} != nil {
			{{- else }}
			{
			{{- end }}
			head := {{ if .FieldPointer }}*{{ end }}p.{{ .FieldName }}
			{{- if (and (eq .HTTPName "Authorization") (isBearer $.HeaderSchemes)) }}
		if !strings.Contains(head, " ") {
			req.Header.Set({{ printf "%q" .HTTPName }}, "Bearer "+head)
		} else {
			{{- end }}
			{{- if eq .Type.Name "array" }}
			for _, val := range head {
				{{- if eq .Type.ElemType.Type.Name "string" }}
				req.Header.Add({{ printf "%q" .HTTPName }}, val)
				{{- else if (and (isAlias .Type.ElemType.Type) (eq (underlyingType .Type.ElemType.Type).Name "string")) }}
				req.Header.Set({{ printf "%q" .HTTPName }}, string(val))
				{{- else }}
				{{ template "partial_client_type_conversion" (typeConversionData .Type.ElemType.Type (aliasedType .FieldType).ElemType.Type "valStr" "val") }}
				req.Header.Add({{ printf "%q" .HTTPName }}, valStr)
				{{- end }}
			}
			{{- else if (and (isAlias .FieldType) (eq (underlyingType .FieldType).Name "string")) }}
			req.Header.Set({{ printf "%q" .HTTPName }}, string(head))
			{{- else if eq .Type.Name "string" }}
			req.Header.Set({{ printf "%q" .HTTPName }}, head)
			{{- else }}
			{{ template "partial_client_type_conversion" (typeConversionData .Type .FieldType "headStr" "head") }}
			req.Header.Set({{ printf "%q" .HTTPName }}, headStr)
			{{- end }}
			{{- if (and (eq .HTTPName "Authorization") (isBearer $.HeaderSchemes)) }}
		}
			{{- end }}
		}
		{{- end }}
	{{- end }}
	{{- range .Payload.Request.Cookies }}
		{{- if .FieldName }}
			{{- if .FieldPointer }}
		if p.{{ .FieldName }} != nil {
			{{- else }}
			{
			{{- end }}
			v{{ if not (eq .Type.Name "string") }}raw{{ end }} := {{ if .FieldPointer }}*{{ end }}p.{{ .FieldName }}
			{{- if not (eq .Type.Name "string" ) }}
			{{ template "partial_client_type_conversion" (typeConversionData .Type .FieldType "vraw" "v") }}
			{{- end }}
			req.AddCookie(&http.Cookie{
				Name: {{ printf "%q" .HTTPName }},
				Value: v,
				{{- if .MaxAge }}
				MaxAge: {{ .MaxAge }},
				{{- end }}
				{{- if .Path }}
				Path: {{ .Path }},
				{{- end }}
				{{- if .Domain }}
				Domain: {{ .Domain }},
				{{- end }}
				{{- if .Secure }}
				Secure: true,
				{{- end }}
				{{- if .HTTPOnly }}
				HttpOnly: true,
				{{- end }}
				{{- if .SameSite }}
				SameSite: {{ .SameSite }},
				{{- end }}
			})
		}
		{{- end }}
	{{- end }}
	{{- if or .Payload.Request.QueryParams }}
		values := req.URL.Query()
	{{- end }}
	{{- range .Payload.Request.QueryParams }}
		{{- if .MapQueryParams }}
		for key, value := range p{{ if .FieldName }}.{{ .FieldName }}{{ end }} {
			{{ template "partial_client_type_conversion" (typeConversionData .Type.KeyType.Type (aliasedType .FieldType).KeyType.Type "keyStr" "key") }}
			{{- if eq .Type.ElemType.Type.Name "array" }}
			for _, val := range value {
				{{ template "partial_client_type_conversion" (typeConversionData .Type.ElemType.Type.ElemType.Type (aliasedType (aliasedType .FieldType).ElemType.Type).ElemType.Type "valStr" "val") }}
				values.Add(keyStr, valStr)
			}
			{{- else }}
			{{ template "partial_client_type_conversion" (typeConversionData .Type.ElemType.Type (aliasedType .FieldType).ElemType.Type "valueStr" "value") }}
			values.Add(keyStr, valueStr)
			{{- end }}
    }
		{{- else if .StringSlice }}
			for _, value := range p{{ if .FieldName }}.{{ .FieldName }}{{ end }} {
				values.Add("{{ .HTTPName }}", value)
			}
		{{- else if .Slice }}
			for _, value := range p{{ if .FieldName }}.{{ .FieldName }}{{ end }} {
				{{ template "partial_client_type_conversion" (typeConversionData .Type.ElemType.Type (aliasedType .FieldType).ElemType.Type "valueStr" "value") }}
				values.Add("{{ .HTTPName }}", valueStr)
			}
		{{- else if .Map }}
			{{- template "partial_client_map_conversion" (mapConversionData .Type .FieldType .HTTPName "p" .FieldName true) }}
		{{- else if .FieldName }}
			{{- if .FieldPointer }}
		if p.{{ .FieldName }} != nil {
			{{- end }}
			{{- $target := printf "p.%s" .FieldName }}
			{{- if .FieldPointer }}
				{{- $target = printf "*p.%s" .FieldName }}
			{{- end }}
		values.Add("{{ .HTTPName }}", {{ template "partial_client_type_expression" (typeConversionData .Type .FieldType "" $target) }})
			{{- if .FieldPointer }}
		}
			{{- end }}
		{{- else }}
			{{- if eq .Type.Name "string" }}
				values.Add("{{ .HTTPName }}", p)
			{{- else if (and (isAlias .Type) (eq (underlyingType .Type).Name "string")) }}
				values.Add("{{ .HTTPName }}", string(p))
			{{- else }}
				{{ template "partial_client_type_conversion" (typeConversionData .Type .FieldType "pStr" "p") }}
				values.Add("{{ .HTTPName }}", pStr)
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
		{{ if .IsJSONRPC }}b{{ else }}body{{ end }} := {{ .Payload.Request.ClientBody.Init.Declaration.Name }}({{ range .Payload.Request.ClientBody.Init.ClientArgs }}{{ if .FieldPointer }}&{{ end }}{{ .VarName }}, {{ end }})
		{{- else }}
		{{ if .IsJSONRPC }}b{{ else }}body{{ end }} := p{{ if .Payload.Request.PayloadAttr }}.{{ .Payload.Request.PayloadAttr }}{{ end }}
		{{- end }}
		{{- if .IsJSONRPC }}
		{{- template "partial_jsonrpc_request_envelope" . }}
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
{{- end }}
	}
}
