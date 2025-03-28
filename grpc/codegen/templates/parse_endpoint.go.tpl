// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	cc *grpc.ClientConn,
{{-  range .Commands }}
	{{- if .Interceptors }}
	{{ .Interceptors.VarName }} {{ .Interceptors.PkgName }}.ClientInterceptors,
	{{- end }}
{{- end }}
	opts ...grpc.CallOption,
) (goa.Endpoint, any, error) {
	{{ .FlagsCode }}
	var (
		data     any
		endpoint goa.Endpoint
		err      error
	)
	{
		switch svcn {
	{{- range .Commands }}
		case "{{ .Name }}":
			c := {{ .PkgName }}.NewClient(cc, opts...)
			switch epn {
		{{- $pkgName := .PkgName }}
		{{- range .Subcommands }}
			case "{{ .Name }}":
				endpoint = c.{{ .MethodVarName }}()
			{{- if .Interceptors }}
				endpoint = {{ .Interceptors.PkgName }}.Wrap{{ .MethodVarName }}ClientEndpoint(endpoint, {{ .Interceptors.VarName }})
			{{- end }}
			{{- if .BuildFunction }}
				data, err = {{ $pkgName}}.{{ .BuildFunction.Name }}({{ range .BuildFunction.ActualParams }}*{{ . }}Flag, {{ end }})
			{{- else if .Conversion }}
				{{ .Conversion }}
			{{- end }}
		{{- end }}
			}
	{{- end }}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
