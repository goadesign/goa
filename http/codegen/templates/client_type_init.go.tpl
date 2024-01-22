{{ comment .Description }}
func {{ .Name }}({{- range .ClientArgs }}{{ .VarName }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
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
	{{ fieldCode . "client" }}
	return {{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }}
}
