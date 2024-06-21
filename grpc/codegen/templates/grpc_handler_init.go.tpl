{{ printf "New%sHandler creates a gRPC handler which serves the %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func New{{ .Method.VarName }}Handler(endpoint goa.Endpoint, h goagrpc.{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler) goagrpc.{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler {
	if h == nil {
		h = goagrpc.New{{ if .ServerStream }}Stream{{ else }}Unary{{ end }}Handler(endpoint, {{ if .Method.Payload }}Decode{{ .Method.VarName }}Request{{ else }}nil{{ end }}{{ if not .ServerStream }}, Encode{{ .Method.VarName }}Response{{ end }})
	}
	return h
}
