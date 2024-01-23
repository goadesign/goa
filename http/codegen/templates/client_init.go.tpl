{{ printf "New%s instantiates HTTP clients for all the %s service servers." .ClientStruct .Service.Name | comment }}
func New{{ .ClientStruct }}(
	scheme string,
	host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restoreBody bool,
	{{- if hasWebSocket . }}
	dialer goahttp.Dialer,
	cfn *ConnConfigurer,
	{{- end }}
) *{{ .ClientStruct }} {
{{- if hasWebSocket . }}
	if cfn == nil {
		cfn = &ConnConfigurer{}
	}
{{- end }}
	return &{{ .ClientStruct }}{
		{{- range .Endpoints }}
		{{ .Method.VarName }}Doer: doer,
		{{- end }}
		RestoreResponseBody: restoreBody,
		scheme:            scheme,
		host:              host,
		decoder:           dec,
		encoder:           enc,
		{{- if hasWebSocket . }}
		dialer: dialer,
		configurer: cfn,
		{{- end }}
	}
}
