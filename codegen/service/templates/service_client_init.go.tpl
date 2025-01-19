{{ printf "New%s initializes a %q service client given the endpoints." .ClientVarName .Name | comment }}
func New{{ .ClientVarName }}({{ .ClientInitArgs }} goa.Endpoint{{ if .HasClientInterceptors }}, ci ClientInterceptors{{ end }}) *{{ .ClientVarName }} {
    return &{{ .ClientVarName }}{
    {{- range .Methods }}
        {{ .VarName }}Endpoint: {{ if .ClientInterceptors }}Wrap{{ .VarName }}ClientEndpoint({{ end }}{{ .ArgName }}{{ if .ClientInterceptors }}, ci){{ end }},
    {{- end }}
    }
}
