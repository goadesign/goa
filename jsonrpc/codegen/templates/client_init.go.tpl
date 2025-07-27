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
	cfn goahttp.ConnConfigureFunc,
	configurer *ConnConfigurer,
	{{- end }}
) *{{ .ClientStruct }} {
	return &{{ .ClientStruct }}{
		Doer:                doer,
		RestoreResponseBody: restoreBody,
		scheme:              scheme,
		host:                host,
		decoder:             dec,
		encoder:             enc,
		{{- if hasWebSocket . }}
		dialer:              dialer,
		configfn:            cfn,
		configurer:          configurer,
		{{- end }}
	}
}
