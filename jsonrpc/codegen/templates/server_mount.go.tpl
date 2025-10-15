{{ printf "%s configures the mux to serve the JSON-RPC %s service methods." .MountServer .Service.Name | comment }}
func {{ .MountServer }}(mux goahttp.Muxer, h *{{ .ServerStruct }}) {
{{- if .HasMixed }}
	// Mixed transports: mount unified handler that negotiates HTTP vs SSE by Accept header
	{{- range (index .Endpoints 0).Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.ServeHTTP)
	{{- end }}
{{- else if .HasSSE }}
	// SSE only: mount SSE handler
	{{- range .Endpoints }}
		{{- range .Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.handleSSE)
		{{- end }}
	{{- end }}
{{- else }}
	// HTTP only
	{{- range (index .Endpoints 0).Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", h.ServeHTTP)
	{{- end }}
{{- end }}
}

{{ printf "%s configures the mux to serve the JSON-RPC %s service methods." .MountServer .Service.Name | comment }}
func (s *{{ .ServerStruct }}) {{ .MountServer }}(mux goahttp.Muxer) {
	{{ .MountServer }}(mux, s)
}
