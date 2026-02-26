{{ printf "New%s initializes a %q service client given the endpoints." .ClientVarName .Name | comment }}
func New{{ .ClientVarName }}({{ if .ClientInitArgs }}{{ .ClientInitArgs }} goa.Endpoint{{ if .HasClientInterceptors }}, ci ClientInterceptors{{ end }}{{ else }}{{ if .HasClientInterceptors }}ci ClientInterceptors{{ end }}{{ end }}) *{{ .ClientVarName }} {
    return &{{ .ClientVarName }}{
    {{- range .Methods }}
        {{ .EndpointField }}: {{ if .ClientInterceptors }}Wrap{{ .VarName }}ClientEndpoint({{ end }}{{ .ArgName }}{{ if .ClientInterceptors }}, ci){{ end }},
		{{- if .HasMixedResults }}
        {{ .StreamEndpointField }}: {{ if .ClientInterceptors }}Wrap{{ .VarName }}ClientEndpoint({{ end }}{{ .StreamArgName }}{{ if .ClientInterceptors }}, ci){{ end }},
		{{- end }}
    {{- end }}
    }
}
