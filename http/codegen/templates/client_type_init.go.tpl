{{ comment .Description }}
func {{ .Declaration.Name }}({{- range .ClientArgs }}{{ .VarName }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
{{- if .OptionalBody }}
	res := &{{ .ReturnTypeName }}{}
	if body != nil {
		{{ .ClientCode }}
		res.{{ .ReturnTypeAttribute }} = {{ if .ReturnIsPrimitivePointer }}&{{ else if .ReturnIsUnion }}*{{ end }}v
	}
{{- else }}
{{- if .ClientCode }}
	{{ .ClientCode }}
	{{- if .ReturnTypeAttribute }}
		res := &{{ .ReturnTypeName }}{
			{{ .ReturnTypeAttribute }}: {{ if .ReturnIsPrimitivePointer }}&{{ end }}v,
		}
	{{- end }}
{{- end }}
{{- if .ReturnIsStruct }}
	{{- if not .ClientCode }}
	{{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }} := &{{ .ReturnTypeName }}{}
	{{- end }}
{{- end }}
	{{- end }}
	{{ fieldCode . "client" }}
	return {{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }}
}
