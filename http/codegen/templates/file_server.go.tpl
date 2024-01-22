{{ printf "%s configures the mux to serve GET request made to %q." .MountHandler (join .RequestPaths ", ") | comment }}
func {{ .MountHandler }}(mux goahttp.Muxer, h http.Handler) {
	{{- if .IsDir }}
		{{- range .RequestPaths }}
	mux.Handle("GET", "{{ . }}{{if ne . "/"}}/{{end}}", h.ServeHTTP)
	mux.Handle("GET", "{{ . }}{{if ne . "/"}}/{{end}}{*{{ $.PathParam }}}", h.ServeHTTP)
		{{- end }}
	{{- else }}
		{{- range .RequestPaths }}
	mux.Handle("GET", "{{ . }}", h.ServeHTTP)
		{{- end }}
	{{- end }}
}
