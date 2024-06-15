// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
	{{- if streamingCmdExists .Commands }}
	dialer goahttp.Dialer,
		{{- range .Commands }}
			{{- if .NeedStream }}
				{{ .VarName }}Configurer *{{ .PkgName }}.ConnConfigurer,
			{{- end }}
		{{- end }}
	{{- end }}
	{{- range $c := .Commands }}
	{{- range .Subcommands }}
		{{- if .MultipartVarName }}
	{{ .MultipartVarName }} {{ $c.PkgName }}.{{ .MultipartFuncName }},
		{{- end }}
	{{- end }}
	{{- end }}
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
			c := {{ .PkgName }}.NewClient(scheme, host, doer, enc, dec, restore{{ if .NeedStream }}, dialer, {{ .VarName }}Configurer{{ end }})
			switch epn {
		{{- $pkgName := .PkgName }}{{ range .Subcommands }}
			case "{{ .Name }}":
				endpoint = c.{{ .MethodVarName }}({{ if .MultipartVarName }}{{ .MultipartVarName }}{{ end }})
			{{- if .BuildFunction }}
				data, err = {{ $pkgName}}.{{ .BuildFunction.Name }}({{ range .BuildFunction.ActualParams }}*{{ . }}Flag, {{ end }})
			{{- else if .Conversion }}
				{{ .Conversion }}
			{{- end }}
			{{- if .StreamFlag }}
				{{- if .BuildFunction }}
				if err == nil {
				{{- end }}
					data, err = {{ $pkgName }}.{{ .BuildStreamPayload }}({{ if or .BuildFunction .Conversion }}data, {{ end }}*{{ .StreamFlag.FullName }}Flag)
				{{- if .BuildFunction }}
				}
				{{- end }}
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
