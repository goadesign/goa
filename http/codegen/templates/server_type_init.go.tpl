{{ comment .Description }}
func {{ .Name }}({{- range .ServerArgs }}{{ .VarName }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
{{- if .ServerCode }}
	{{ .ServerCode }}
	{{- if .ReturnTypeAttribute }}
		res := &{{ .ReturnTypeName }}{
			{{ .ReturnTypeAttribute }}: {{ if .ReturnIsPrimitivePointer }}&{{ end }}v,
		}
	{{- end }}
{{- end }}
{{- if .ReturnIsStruct }}
	{{- if not .ServerCode }}
	{{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }} := &{{ .ReturnTypeName }}{}
	{{- end }}
	{{ fieldCode . "server" }}
{{- end }}
	return {{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }}
}
