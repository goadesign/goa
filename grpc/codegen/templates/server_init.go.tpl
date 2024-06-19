{{ printf "%s instantiates the server struct with the %s service endpoints." .ServerInit .Service.Name | comment }}
func {{ .ServerInit }}(e *{{ .Service.PkgName }}.Endpoints{{ if .HasUnaryEndpoint }}, uh goagrpc.UnaryHandler{{ end }}{{ if .HasStreamingEndpoint }}, sh goagrpc.StreamHandler{{ end }}) *{{ .ServerStruct }} {
	return &{{ .ServerStruct }}{
	{{- range .Endpoints }}
		{{ .Method.VarName }}H: New{{ .Method.VarName }}Handler(e.{{ .Method.VarName }}{{ if .ServerStream }}, sh{{ else }}, uh{{ end }}),
	{{- end }}
	}
}
