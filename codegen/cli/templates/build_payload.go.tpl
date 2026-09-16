{{ printf "%s builds the payload for the %s %s endpoint from CLI flags." .Name .ServiceName .MethodName | comment }}
func {{ .Name }}({{ range $index, $param := .FormalParams }}{{ $param }} {{ formalParamType $ $index }}, {{ end }}) ({{ .ResultType }}, error) {
{{- if .CheckErr }}
	var err error
{{- end }}
{{- range .Fields }}
	{{- if .VarName }}
		var {{ .VarName }} {{ .TypeRef }}
		{
			{{ .Init }}
		}
	{{- end }}
{{- end }}
{{- with .PayloadInit }}
	{{- if .OptionalBody }}
		res := &{{ .ReturnTypeName }}{}
		if body != nil {
			{{ .Code }}
			res.{{ .ReturnTypeAttribute }} = {{ if .ReturnTypeAttributePointer }}&{{ else if .ReturnTypeAttributeUnion }}*{{ end }}v
		}
	{{- else }}
	{{- if .Code }}
		{{ .Code }}
		{{- if .ReturnTypeAttribute }}
			res := &{{ .ReturnTypeName }}{
				{{ .ReturnTypeAttribute }}: {{ if .ReturnTypeAttributePointer }}&{{ end }}v,
			}
		{{- end }}
	{{- end }}
	{{- if .ReturnIsStruct }}
		{{- if not .Code }}
		{{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }} := &{{ .ReturnTypeName }}{}
		{{- end }}
		{{ fieldCode . }}
	{{- end }}
	{{- end }}
	{{- if and .OptionalBody .ReturnIsStruct }}
		{{ fieldCode . }}
	{{- end }}
	return {{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }}, nil
{{- end }}
}
