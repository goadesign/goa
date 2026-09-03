// {{ .ClientDeclaration.Name }} is the {{ printf "%q" .Name }} service client.
type {{ .ClientDeclaration.Name }} struct {
{{- range .Methods}}
	{{ .EndpointField }} goa.Endpoint
	{{- if .HasMixedResults }}
	{{ .StreamEndpointField }} goa.Endpoint
	{{- end }}
{{- end }}
}
