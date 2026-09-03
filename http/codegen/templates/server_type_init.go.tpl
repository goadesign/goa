{{ comment .Description }}
func {{ .Declaration.Name }}({{- range .ServerArgs }}{{ .VarName }} {{ .TypeRef }}, {{ end }}) {{ .ReturnTypeRef }} {
{{- if .OptionalBody }}
	res := &{{ .ReturnTypeName }}{}
	if body != nil {
		{{ .ServerCode }}
		res.{{ .ReturnTypeAttribute }} = {{ if .ReturnIsPrimitivePointer }}&{{ else if .ReturnIsUnion }}*{{ end }}v
	}
	{{- if .BodyDefault }} else {
		{{- range .BodyDefault.Declarations }}
		{{ . }}
		{{- end }}
		res.{{ .ReturnTypeAttribute }} = {{ .BodyDefault.Expression }}
	}
	{{- end }}
{{- else }}
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
	{{- end }}
	{{- if and .OptionalBody .ReturnIsStruct }}
	{{ fieldCode . "server" }}
	{{- end }}
	return {{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }}
}
