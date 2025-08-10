{{ printf "%s configures the mux to serve the JSON-RPC %s service methods." .MountServer .Service.Name | comment }}
func {{ .MountServer }}(mux goahttp.Muxer, h *{{ .ServerStruct }}) {
{{- if .HasMixed }}
	// Mount ServeHTTP which handles both regular JSON-RPC and SSE based on Accept header
	{{- range (index .Endpoints 0).Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.ServeHTTP)
	{{- end }}
{{- else if .HasSSE }}
	// Mount SSE handler for all endpoint routes
	{{- range .Endpoints }}
		{{- range .Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.handleSSE)
		{{- end }}
	{{- end }}
{{- else }}
	{{- range (index .Endpoints 0).Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.ServeHTTP)
	{{- end }}
{{- end }}
}

{{ printf "%s configures the mux to serve the JSON-RPC %s service methods." .MountServer .Service.Name | comment }}
func (s *{{ .ServerStruct }}) {{ .MountServer }}(mux goahttp.Muxer) {
	{{ .MountServer }}(mux, s)
}
