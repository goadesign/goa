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
	streamOpts ...jsonrpc.StreamConfigOption,
	{{- end }}
) *{{ .ClientStruct }} {
	{{- if hasWebSocket . }}
	// Create stream configuration from options
	streamConfig := jsonrpc.NewStreamConfig(streamOpts...)
	{{- end }}
	
	return &{{ .ClientStruct }}{
		Doer:                doer,
		{{- range .Endpoints }}
		{{- if isSSEEndpoint . }}
		{{ .Method.VarName }}Doer: doer,
		{{- end }}
		{{- end }}
		RestoreResponseBody: restoreBody,
		scheme:              scheme,
		host:                host,
		decoder:             dec,
		encoder:             enc,
		{{- if hasWebSocket . }}
		dialer:              dialer,
		configfn:            cfn,
		streamConfig:        streamConfig,
		{{- end }}
	}
}
