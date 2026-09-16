{{ printf "%s creates HTTP clients for all the %s service servers." .ClientInitDeclaration.Name .Service.Name | comment }}
func {{ .ClientInitDeclaration.Name }}(
	scheme string,
	host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restoreBody bool,
) *{{ .ClientStructDeclaration.Name }} {
	return &{{ .ClientStructDeclaration.Name }}{
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
	}
}
