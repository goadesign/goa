{{- range .Headers }}
	{{- $hasDefault := and (or .FieldPointer .Slice) .DefaultValue }}
	{{- $checkNil := or .FieldPointer .Slice (eq .TypeName "bytes") (eq .TypeName "any") $hasDefault }}
	{{- if $checkNil }}
	if res.Projected.{{ .FieldName }} != nil {
	{{- end }}
	{{- if and (eq .TypeName "string") (not .IsAliased) }}
	w.Header().Set("{{ .CanonicalName }}", {{ if .FieldPointer }}*{{ end }}res.Projected.{{ .FieldName }})
	{{- else }}
	{{- if not $checkNil }}
	{
	{{- end }}
		{{- if .IsAliased }}
		val := {{ goTypeRef .TypeName .ElemTypeName }}({{ if .FieldPointer }}*{{ end }}res.Projected.{{ .FieldName }})
		{{ template "partial_header_conversion" (headerConversionData .TypeName .ElemTypeName (printf "%ss" .VarName) true "val") }}
		{{- else }}
		val := res.Projected.{{ .FieldName }}
		{{ template "partial_header_conversion" (headerConversionData .TypeName .ElemTypeName (printf "%ss" .VarName) (not .FieldPointer) "val") }}
		{{- end }}
		w.Header().Set("{{ .CanonicalName }}", {{ .VarName }}s)
	{{- if not $checkNil }}
	}
	{{- end }}
	{{- end }}
	{{- if $hasDefault }}
	} else {
		w.Header().Set("{{ .CanonicalName }}", "{{ printValue .TypeName .ElemTypeName .DefaultValue }}")
	{{- end }}
	{{- if or $checkNil $hasDefault }}
	}
	{{- end }}
{{- end }}
{{- range .Cookies }}
	{{- $hasDefault := and (or .FieldPointer .Slice) .DefaultValue }}
	{{- $checkNil := or .FieldPointer .Slice (eq .TypeName "bytes") (eq .TypeName "any") $hasDefault }}
	{{- if $checkNil }}
	if res.Projected.{{ .FieldName }} != nil {
	{{- end }}
	{{- if eq .TypeName "string" }}
	{{ .VarName }} := {{ if .FieldPointer }}*{{ end }}res.Projected.{{ .FieldName }}
	{{- else if .IsAliased }}
	{{ .VarName }}raw := {{ goTypeRef .TypeName .ElemTypeName }}({{ if .FieldPointer }}*{{ end }}res.Projected.{{ .FieldName }})
	{{ template "partial_header_conversion" (headerConversionData .TypeName .ElemTypeName (printf "%sraw" .VarName) true .VarName) }}
	{{- else }}
	{{ .VarName }}raw := res.Projected.{{ .FieldName }}
	{{ template "partial_header_conversion" (headerConversionData .TypeName .ElemTypeName (printf "%sraw" .VarName) (not .FieldPointer) .VarName) }}
	{{- end }}
	{{- if $hasDefault }}
	} else {
		{{ .VarName }} := "{{ printValue .TypeName .ElemTypeName .DefaultValue }}"
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
	{{- if or $checkNil $hasDefault }}
	}
	{{- end }}
{{- end }}
