	{{- $servBodyLen := len .ServerBody }}
	{{- if gt $servBodyLen 0 }}
	enc := encoder(ctx, w)
	{{- end }}
	{{- if gt $servBodyLen 0 }}
		{{- if and (gt $servBodyLen 1) $.ViewedResult }}
	var body any
	switch res.View	{
			{{- range $.ViewedResult.Views }}
	case {{ printf "%q" .Name }}{{ if eq .Name "default" }}, ""{{ end }}:
		{{- $vsb := (viewedServerBody $.ServerBody .Name) }}
		body = {{ $vsb.Init.Name }}({{ range $vsb.Init.ServerArgs }}{{ .Ref }}, {{ end }})
			{{- end }}
	}
		{{- else if (index .ServerBody 0).Init }}
			{{- if .ErrorHeader }}
	var body any
	if formatter != nil {
		body = formatter(ctx, {{ (index (index .ServerBody 0).Init.ServerArgs 0).Ref }})
	} else {
			{{- end }}
	body {{ if not .ErrorHeader}}:{{ end }}= {{ (index .ServerBody 0).Init.Name }}({{ range (index .ServerBody 0).Init.ServerArgs }}{{ .Ref }}, {{ end }})
			{{- if .ErrorHeader }}
	}
			{{- end }}
		{{- else }}
	body := res{{ if $.ViewedResult }}.Projected{{ end }}{{ if .ResultAttr }}.{{ .ResultAttr }}{{ end }}
		{{- end }}
	{{- end }}
	{{- range .Headers }}
		{{- $initDef := and (or .FieldPointer .Slice) .DefaultValue (not $.TagName) }}
		{{- $checkNil := and (or .FieldPointer .Slice (eq .Type.Name "bytes") (eq .Type.Name "any") $initDef) (not $.TagName) }}
		{{- if $checkNil }}
	if res{{ if .FieldName }}.{{ end }}{{ if $.ViewedResult }}Projected.{{ end }}{{ if .FieldName }}{{ .FieldName }}{{ end }} != nil {
		{{- end }}

		{{- if and (eq .Type.Name "string") (not (isAliased .FieldType)) }}
	w.Header().Set("{{ .CanonicalName }}", {{ if or .FieldPointer $.ViewedResult }}*{{ end }}res{{ if $.ViewedResult }}.Projected{{ end }}{{ if .FieldName }}.{{ .FieldName }}{{ end }})
		{{- else }}
{{- if not $checkNil }}
{
{{- end }}
			{{- if isAliased .FieldType }}
	val := {{ goTypeRef .Type }}({{ if .FieldPointer }}*{{ end }}res{{ if $.ViewedResult }}.Projected{{ end }}{{ if .FieldName }}.{{ .FieldName }}{{ end }})
	{{ template "partial_header_conversion" (headerConversionData .Type (printf "%ss" .VarName) true "val") }}
			{{- else }}
	val := res{{ if $.ViewedResult }}.Projected{{ end }}{{ if .FieldName }}.{{ .FieldName }}{{ end }}
	{{ template "partial_header_conversion" (headerConversionData .Type (printf "%ss" .VarName) (not .FieldPointer) "val") }}
			{{- end }}
	w.Header().Set("{{ .CanonicalName }}", {{ .VarName }}s)
{{- if not $checkNil }}
}
{{- end }}
		{{- end }}

		{{- if $initDef }}
	{{ if $checkNil }} } else { {{ else }}if res{{ if $.ViewedResult }}.Projected{{ end }}.{{ .FieldName }} == nil { {{ end }}
		w.Header().Set("{{ .CanonicalName }}", "{{ printValue .Type .DefaultValue }}")
		{{- end }}

		{{- if or $checkNil $initDef }}
	}
		{{- end }}

	{{- end }}

	{{- range .Cookies }}
		{{- $initDef := and (or .FieldPointer .Slice) .DefaultValue }}
		{{- $checkNil := and (or .FieldPointer .Slice (eq .Type.Name "bytes") (eq .Type.Name "any") $initDef) }}
		{{- if $checkNil }}
	if res.{{ if $.ViewedResult }}Projected.{{ end }}{{ .FieldName }} != nil {
		{{- end }}

		{{- if eq .Type.Name "string" }}
	{{ .VarName }} := {{ if or .FieldPointer $.ViewedResult }}*{{ end }}res{{ if $.ViewedResult }}.Projected{{ end }}{{ if .FieldName }}.{{ .FieldName }}{{ end }}
		{{- else }}
			{{- if isAliased .FieldType }}
	{{ .VarName }}raw := {{ goTypeRef .Type }}({{ if .FieldPointer }}*{{ end }}res{{ if $.ViewedResult }}.Projected{{ end }}{{ if .FieldName }}.{{ .FieldName }}{{ end }})
	{{ template "partial_header_conversion" (headerConversionData .Type (printf "%sraw" .VarName) true .VarName) }}
			{{- else }}
	{{ .VarName }}raw := res{{ if $.ViewedResult }}.Projected{{ end }}{{ if .FieldName }}.{{ .FieldName }}{{ end }}
	{{ template "partial_header_conversion" (headerConversionData .Type (printf "%sraw" .VarName) (not .FieldPointer) .VarName) }}
			{{- end }}
		{{- end }}

		{{- if $initDef }}
	{{ if $checkNil }} } else { {{ else }}if res{{ if $.ViewedResult }}.Projected{{ end }}.{{ .FieldName }} == nil { {{ end }}
		{{ .VarName }} := "{{ printValue .Type .DefaultValue }}"
		{{- end }}
		http.SetCookie(w, &http.Cookie{
			Name: {{ printf "%q" .HTTPName }},
			Value: {{ .VarName }},
			{{- if .MaxAge }}
			MaxAge: {{ .MaxAge }},
			{{- end }}
			{{- if .Path }}
			Path: {{ printf "%q" .Path }},
			{{- end }}
			{{- if .Domain }}
			Domain: {{ printf "%q" .Domain }},
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
		{{- if or $checkNil $initDef }}
	}
		{{- end }}

	{{- end }}

	{{- if .ErrorHeader }}
	w.Header().Set("goa-error", res.GoaErrorName())
	{{- end }}
	w.WriteHeader({{ .StatusCode }})