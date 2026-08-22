{{ printf "%s configures the mux to serve the JSON-RPC %s service methods." .MountServerDeclaration.Name .Service.Name | comment }}
func {{ .MountServerDeclaration.Name }}(mux goahttp.Muxer, h *{{ .ServerStructDeclaration.Name }}) {
{{- if .HasMixed }}
	// ServeHTTP checks the Accept header and chooses an ordinary response or server-sent events.
	{{- range (index .Endpoints 0).Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.ServeHTTP)
	{{- end }}
{{- else if .HasSSE }}
	// Every method in this server writes server-sent events.
	{{- range .Endpoints }}
		{{- range .Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.handleSSE)
		{{- end }}
	{{- end }}
{{- else }}
	// Every method in this server writes one ordinary JSON-RPC response.
	{{- range (index .Endpoints 0).Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.ServeHTTP)
	{{- end }}
{{- end }}
}

{{ printf "%s configures the mux to serve the JSON-RPC %s service methods." .MountServerDeclaration.Name .Service.Name | comment }}
func (s *{{ .ServerStructDeclaration.Name }}) {{ .MountServerDeclaration.Name }}(mux goahttp.Muxer) {
	{{ .MountServerDeclaration.Name }}(mux, s)
}
