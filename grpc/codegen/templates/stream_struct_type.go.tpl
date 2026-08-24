{{ printf "%s implements the %s interface." .Declaration.Name .ServiceInterface | comment }}
type {{ .Declaration.Name }} struct {
	stream {{ .Interface }}
{{- if and (eq .Type "server") .Endpoint.Request.LegacyDecode }}
	// legacy indicates that the client speaks the legacy stream protocol
	// which sends raw stream item frames instead of typed envelopes.
	legacy bool
{{- end }}
{{- if and .Endpoint.Method.ViewedResult (not .Endpoint.Method.ViewedResult.ViewName) }}
	view string
	{{- if eq .Type "server" }}
	{{ comment "sentView is the result view named in the response header. Later sends must use the same view." }}
	sentView string
	{{- else }}
	viewSet bool
	{{- end }}
{{- end }}
}
