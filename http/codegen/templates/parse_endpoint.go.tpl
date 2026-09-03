// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func {{ .Declaration.Name }}(
	{{ .Variables.Scheme }}, {{ .Variables.Host }} string,
	{{ .Variables.Doer }} goahttp.Doer,
	{{ .Variables.Encoder }} func(*http.Request) goahttp.Encoder,
	{{ .Variables.Decoder }} func(*http.Response) goahttp.Decoder,
	{{ .Variables.Restore }} bool,
	{{- if streamingCmdExists .Commands }}
	{{ .Variables.Dialer }} goahttp.Dialer,
		{{- range .Commands }}
			{{- if .NeedDialer }}
				{{ if .JSONRPC }}{{ .ConfigurerLocal.VarName }} goahttp.ConnConfigureFunc,{{ else }}{{ .ConfigurerLocal.VarName }} *{{ .PkgName }}.{{ .Configurer.Name }},{{ end }}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- range $i, $c := .Commands }}
	{{- range .Subcommands }}
		{{- if .MultipartVarName }}
	{{ .MultipartLocal.VarName }} {{ $c.PkgName }}.{{ .MultipartFuncDeclaration.Name }},
		{{- end }}
	{{- end }}
	{{- if .Interceptors }}
	{{ .Interceptors.ParserVar }} {{ .Interceptors.PkgName }}.{{ .Interceptors.ClientInterceptorsDeclaration.Name }},
	{{- end }}
	{{- end }}
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
			{{ $.Variables.Client }} := {{ .PkgName }}.{{ .ClientInit.Name }}({{ $.Variables.Scheme }}, {{ $.Variables.Host }}, {{ $.Variables.Doer }}, {{ $.Variables.Encoder }}, {{ $.Variables.Decoder }}, {{ $.Variables.Restore }}{{ if .NeedDialer }}, {{ $.Variables.Dialer }}, {{ .ConfigurerLocal.VarName }}{{ end }})
			switch {{ $.Variables.MethodName }} {
		{{- $pkgName := .PkgName }}
		{{- range .Subcommands }}
			case "{{ .Name }}":
				{{ $.Variables.Endpoint }} = {{ $.Variables.Client }}.{{ .MethodVarName }}({{ if .MultipartLocal }}{{ .MultipartLocal.VarName }}{{ end }})
			{{- if .Interceptors }}
				{{ $.Variables.Endpoint }} = {{ .Interceptors.PkgName }}.{{ .Interceptors.ClientEndpointWrapperDeclaration.Name }}({{ $.Variables.Endpoint }}, {{ .Interceptors.ParserVar }})
			{{- end }}
			{{- if .BuildFunction }}
				{{ $.Variables.Data }}, {{ $.Variables.Error }} = {{ $pkgName }}.{{ .BuildFunction.Name }}({{- if .ActualArgs }}{{ range $index, $argument := .ActualArgs }}{{ if $index }}, {{ end }}{{ $argument }}{{ end }}{{ else }}{{ range $index, $variable := .ActualPointerVars }}{{ if $index }}, {{ end }}*{{ $variable }}{{ end }}{{ end }})
			{{- else if .Conversion }}
				{{ .Conversion }}
			{{- end }}
			{{- if .StreamFlag }}
				if {{ $.Variables.Error }} == nil {
					{{- if .StreamFlag.TracksPresence }}
					if {{ .StreamPointerVar }}.value == nil {
						{{ $.Variables.Error }} = fmt.Errorf("missing required flag --{{ .StreamFlag.Name }}")
					} else {
						{{ $.Variables.Data }}, {{ $.Variables.Error }} = {{ $pkgName }}.{{ .BuildStreamPayloadDeclaration.Name }}({{ if or .BuildFunction .Conversion }}{{ $.Variables.Data }}, {{ end }}*{{ .StreamPointerVar }}.value)
					}
					{{- else }}
					{{ $.Variables.Data }}, {{ $.Variables.Error }} = {{ $pkgName }}.{{ .BuildStreamPayloadDeclaration.Name }}({{ if or .BuildFunction .Conversion }}{{ $.Variables.Data }}, {{ end }}*{{ .StreamPointerVar }})
					{{- end }}
				}
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
