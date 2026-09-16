{{ printf "%s initializes a %q service client given the endpoints." .NewClientDeclaration.Name .Name | comment }}
func {{ .NewClientDeclaration.Name }}({{ if .ClientInitArgs }}{{ .ClientInitArgs }} goa.Endpoint{{ if .HasClientInterceptors }}, ci {{ .ClientInterceptorsDeclaration.Name }}{{ end }}{{ else }}{{ if .HasClientInterceptors }}ci {{ .ClientInterceptorsDeclaration.Name }}{{ end }}{{ end }}) *{{ .ClientDeclaration.Name }} {
    return &{{ .ClientDeclaration.Name }}{
    {{- range .Methods }}
        {{ .EndpointField }}: {{ if .ClientInterceptors }}{{ .ClientEndpointWrapperDeclaration.Name }}({{ end }}{{ .ArgName }}{{ if .ClientInterceptors }}, ci){{ end }},
		{{- if .HasMixedResults }}
        {{ .StreamEndpointField }}: {{ if .ClientInterceptors }}{{ .ClientEndpointWrapperDeclaration.Name }}({{ end }}{{ .StreamArgName }}{{ if .ClientInterceptors }}, ci){{ end }},
		{{- end }}
    {{- end }}
    }
}
