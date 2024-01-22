{{ printf "%s lists the %s service endpoint HTTP handlers." .ServerStruct .Service.Name | comment }}
type {{ .ServerStruct }} struct {
	Mounts []*{{ .MountPointStruct }}
	{{- range .Endpoints }}
	{{ .Method.VarName }} http.Handler
	{{- end }}
	{{- range .FileServers }}
	{{ .VarName }} http.Handler
	{{- end }}
}
