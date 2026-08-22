{{ printf "%s configures the mux to serve the %q service %q endpoint." .MountHandlerDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .MountHandlerDeclaration.Name }}(mux goahttp.Muxer, h http.Handler) {
	f, ok := h.(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	{{- range .Routes }}
	mux.Handle("{{ .Verb }}", "{{ .Path }}", f)
	{{- end }}
}
