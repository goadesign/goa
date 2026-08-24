{{ printf "%s creates a gRPC handler which serves the %q service %q endpoint." .ServerHandlerDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .ServerHandlerDeclaration.Name }}(endpoint goa.Endpoint, h goagrpc.{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler) goagrpc.{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler {
	if h == nil {
		h = goagrpc.New{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler(endpoint, {{ if .Method.Payload }}{{ .ServerDecodeDeclaration.Name }}{{ else }}nil{{ end }}{{ if not .ServerStream }}, {{ .ServerEncodeDeclaration.Name }}{{ end }})
	}
	return h
}
