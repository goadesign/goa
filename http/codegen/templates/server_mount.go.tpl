{{ printf "%s configures the mux to serve the %s endpoints." .MountServerDeclaration.Name .Service.Name | comment }}
func {{ .MountServerDeclaration.Name }}(mux goahttp.Muxer, h *{{ .ServerStructDeclaration.Name }}) {
	{{- range .Endpoints }}
	{{ .MountHandlerDeclaration.Name }}(mux, h.{{ .Method.VarName }})
	{{- end }}
	{{- range .FileServers }}
		{{- if .Redirect }}
	{{ .MountHandlerDeclaration.Name }}(mux, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "{{ .Redirect.URL }}", {{ .Redirect.StatusCode }})
		}))
		{{- else }}
			{{- $mountHandler := .MountHandlerDeclaration.Name }}
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

{{ printf "%s configures the mux to serve the %s endpoints." .MountServerDeclaration.Name .Service.Name | comment }}
func (s *{{ .ServerStructDeclaration.Name }}) {{ .MountServerDeclaration.Name }}(mux goahttp.Muxer) {
	{{ .MountServerDeclaration.Name }}(mux, s)
}
