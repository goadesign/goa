{{ printf "%s configures the mux to serve the JSON-RPC %s service methods." .MountServerDeclaration.Name .Service.Name | comment }}
func {{ .MountServerDeclaration.Name }}(mux goahttp.Muxer, h *{{ .ServerStructDeclaration.Name }}) {
{{- if .HasMixed }}
	// ServeHTTP chooses ordinary JSON-RPC handling or server-sent events.
	{{- range (index .Endpoints 0).Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.ServeHTTP)
	{{- end }}
{{- else if .HasSSE }}
	// This server handles every method through server-sent events.
	{{- range (index .Endpoints 0).Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.handleSSE)
	{{- end }}
{{- else }}
	// This server handles ordinary JSON-RPC request bodies.
	{{- range (index .Endpoints 0).Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.ServeHTTP)
	{{- end }}
{{- end }}
}

{{ printf "%s configures the mux to serve the JSON-RPC %s service methods." .MountServerDeclaration.Name .Service.Name | comment }}
func (s *{{ .ServerStructDeclaration.Name }}) {{ .MountServerDeclaration.Name }}(mux goahttp.Muxer) {
	{{ .MountServerDeclaration.Name }}(mux, s)
}
