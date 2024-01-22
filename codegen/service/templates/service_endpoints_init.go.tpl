

{{ printf "New%s wraps the methods of the %q service with endpoints." .VarName .Name | comment }}
func New{{ .VarName }}(s {{ .ServiceVarName }}) *{{ .VarName }} {
{{- if .Schemes }}
	// Casting service to Auther interface
	a := s.(Auther)
{{- end }}
	return &{{ .VarName }}{
{{- range .Methods }}
		{{ .VarName }}: New{{ .VarName }}Endpoint(s{{ range .Schemes }}, a.{{ .Type }}Auth{{ end }}),
{{- end }}
	}
}