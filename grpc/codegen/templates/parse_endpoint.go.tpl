// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func {{ .Declaration.Name }}(
	{{ .Variables.Connection }} *grpc.ClientConn,
{{-  range .Commands }}
	{{- if .Interceptors }}
	{{ .Interceptors.ParserVar }} {{ .Interceptors.PkgName }}.{{ .Interceptors.ClientInterceptorsDeclaration.Name }},
	{{- end }}
{{- end }}
	{{ .Variables.Options }} ...grpc.CallOption,
) (goa.Endpoint, any, error) {
	{{ .FlagsCode }}
	var (
		{{ .Variables.Data }}     any
		{{ .Variables.Endpoint }} goa.Endpoint
		{{ .Variables.Error }}      error
	)
	{
		switch {{ .Variables.ServiceName }} {
	{{- range .Commands }}
		case "{{ .Name }}":
			{{ $.Variables.Client }} := {{ .PkgName }}.{{ .ClientInit.Name }}({{ $.Variables.Connection }}, {{ $.Variables.Options }}...)
			switch {{ $.Variables.MethodName }} {
		{{- $pkgName := .PkgName }}
		{{- range .Subcommands }}
			case "{{ .Name }}":
				{{ $.Variables.Endpoint }} = {{ $.Variables.Client }}.{{ .MethodVarName }}()
			{{- if .Interceptors }}
				{{ $.Variables.Endpoint }} = {{ .Interceptors.PkgName }}.{{ .Interceptors.ClientEndpointWrapperDeclaration.Name }}({{ $.Variables.Endpoint }}, {{ .Interceptors.ParserVar }})
			{{- end }}
			{{- if .BuildFunction }}
				{{ $.Variables.Data }}, {{ $.Variables.Error }} = {{ $pkgName}}.{{ .BuildFunction.Name }}({{- if .ActualArgs }}{{ range $index, $argument := .ActualArgs }}{{ if $index }}, {{ end }}{{ $argument }}{{ end }}{{ else }}{{ range $index, $variable := .ActualPointerVars }}{{ if $index }}, {{ end }}*{{ $variable }}{{ end }}{{ end }})
			{{- else if .Conversion }}
				{{ .Conversion }}
			{{- end }}
		{{- end }}
			}
	{{- end }}
		}
	}
	if {{ .Variables.Error }} != nil {
		return nil, nil, {{ .Variables.Error }}
	}

	return {{ .Variables.Endpoint }}, {{ .Variables.Data }}, nil
}
