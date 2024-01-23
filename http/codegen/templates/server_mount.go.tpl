{{ printf "%s configures the mux to serve the %s endpoints." .MountServer .Service.Name | comment }}
func {{ .MountServer }}(mux goahttp.Muxer, h *{{ .ServerStruct }}) {
	{{- range .Endpoints }}
	{{ .MountHandler }}(mux, h.{{ .Method.VarName }})
	{{- end }}
	{{- range .FileServers }}
		{{- if .Redirect }}
	{{ .MountHandler }}(mux, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "{{ .Redirect.URL }}", {{ .Redirect.StatusCode }})
		}))
	 	{{- else if .IsDir }}
			{{- $filepath := addLeadingSlash (removeTrailingIndexHTML .FilePath) }}
	{{ .MountHandler }}(mux, {{ range .RequestPaths }}{{if ne . $filepath }}goahttp.Replace("{{ . }}", "{{ $filepath }}", {{ end }}{{ end }}h.{{ .VarName }}){{ range .RequestPaths }}{{ if ne . $filepath }}){{ end}}{{ end }}
		{{- else }}
			{{- $filepath := addLeadingSlash (removeTrailingIndexHTML .FilePath) }}
	{{ .MountHandler }}(mux, {{ range .RequestPaths }}{{if ne . $filepath }}goahttp.Replace("", "{{ $filepath }}", {{ end }}{{ end }}h.{{ .VarName }}){{ range .RequestPaths }}{{ if ne . $filepath }}){{ end}}{{ end }}
		{{- end }}
	{{- end }}
}

{{ printf "%s configures the mux to serve the %s endpoints." .MountServer .Service.Name | comment }}
func (s *{{ .ServerStruct }}) {{ .MountServer }}(mux goahttp.Muxer) {
	{{ .MountServer }}(mux, s)
}
