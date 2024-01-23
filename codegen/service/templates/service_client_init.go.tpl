{{ printf "New%s initializes a %q service client given the endpoints." .ClientVarName .Name | comment }}
func New{{ .ClientVarName }}({{ .ClientInitArgs }} goa.Endpoint) *{{ .ClientVarName }} {
	return &{{ .ClientVarName }}{
{{- range .Methods }}
		{{ .VarName }}Endpoint: {{ .ArgName }},
{{- end }}
	}
}
