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
		{{- else }}
			{{- $mountHandler := .MountHandler }}
			{{- $varName := .VarName }}
			{{- $isDir := .IsDir }}
			{{- range .RequestPaths }}
				{{- $stripped := addLeadingSlash . }}
				{{- if not $isDir }}
					{{- $stripped = (dir $stripped) }}
				{{- end }}
				{{- if eq $stripped "/" }}
	{{ $mountHandler }}(mux, h.{{ $varName }}) 
				{{- else }}
	{{ $mountHandler }}(mux, http.StripPrefix("{{ $stripped }}", h.{{ $varName }}))
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
}

{{ printf "%s configures the mux to serve the %s endpoints." .MountServer .Service.Name | comment }}
func (s *{{ .ServerStruct }}) {{ .MountServer }}(mux goahttp.Muxer) {
	{{ .MountServer }}(mux, s)
}
