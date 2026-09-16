{{ printf "%s implements the %s.%s interface." .ServerStructDeclaration.Name .ServerProtobufPkgName .ServerInterface | comment }}
type {{ .ServerStructDeclaration.Name }} struct {
{{- range .Endpoints }}
	{{ .Method.VarName }}H {{ if .ServerStream }}goagrpc.StreamHandler{{ else }}goagrpc.UnaryHandler{{ end }}
{{- end }}
	{{ .ServerProtobufPkgName }}.{{ .UnimplementedServer }}
}
