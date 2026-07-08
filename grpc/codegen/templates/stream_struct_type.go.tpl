{{ printf "%s implements the %s interface." .VarName .ServiceInterface | comment }}
type {{ .VarName }} struct {
	stream {{ .Interface }}
{{- if and (eq .Type "server") .Endpoint.Request.LegacyDecode }}
	// legacy indicates that the client speaks the legacy stream protocol
	// which sends raw stream item frames instead of typed envelopes.
	legacy bool
{{- end }}
{{- if .Endpoint.Method.ViewedResult }}
	view string
{{- end }}
}
