{{ printf "%s lists the %s service endpoint HTTP handlers." .ServerStructDeclaration.Name .Service.Name | comment }}
type {{ .ServerStructDeclaration.Name }} struct {
	Mounts []*{{ .MountPointStructDeclaration.Name }}
	{{- range .Endpoints }}
	{{ .Method.VarName }} http.Handler
	{{- end }}
	{{- range .FileServers }}
	{{ .VarName }} http.Handler
	{{- end }}
}
