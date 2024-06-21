{{ printf "%s implements the %s.%s interface." .ServerStruct .PkgName .ServerInterface | comment }}
type {{ .ServerStruct }} struct {
{{- range .Endpoints }}
	{{ .Method.VarName }}H {{ if .ServerStream }}goagrpc.StreamHandler{{ else }}goagrpc.UnaryHandler{{ end }}
{{- end }}
	{{ .PkgName }}.Unimplemented{{ .ServerInterface }}
}
