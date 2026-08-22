{{ range .Routes }}// {{ .PathInit.Description }}
func {{ if $.Client }}{{ .PathInit.ClientDeclaration.Name }}{{ else }}{{ .PathInit.Declaration.Name }}{{ end }}({{ range .PathInit.ServerArgs }}{{ .VarName }} {{ .TypeRef }}, {{ end }}) {{ .PathInit.ReturnTypeRef }} {
{{- .PathInit.ServerCode }}
}
{{ end }}
